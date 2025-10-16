package cluster

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func HelmInstallChart(config *rest.Config) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("helm_install_chart",
			mcp.WithDescription("Add Helm repo, update it, and install chart."),
			mcp.WithString("release_name", mcp.Description("Name of the Helm release"), mcp.Required()),
			mcp.WithString("chart", mcp.Description("Chart name (e.g. kyverno/kyverno)"), mcp.Required()),
			mcp.WithString("project", mcp.Description("project to install the chart into"), mcp.Required()),
			mcp.WithString("repo_name", mcp.Description("Name of the Helm repo (e.g. kyverno)"), mcp.Required()),
			mcp.WithString("repo_url", mcp.Description("URL of the Helm repo (e.g. https://kyverno.github.io/helm)"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			releaseName := request.Params.Arguments["release_name"].(string)
			chart := request.Params.Arguments["chart"].(string)
			project := request.Params.Arguments["project"].(string)
			repoName := request.Params.Arguments["repo_name"].(string)
			repoURL := request.Params.Arguments["repo_url"].(string)

			// Setup Helm environment
			settings := cli.New()
			repoFile := settings.RepositoryConfig
			repoCache := settings.RepositoryCache

			// Step 1: Add repo
			repoEntry := &repo.Entry{
				Name: repoName,
				URL:  repoURL,
			}
			repoFileObj := repo.NewFile()
			repoFileObj.Update(repoEntry)

			if err := os.MkdirAll(repoCache, os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create repo cache dir: %w", err)
			}

			if err := repoFileObj.WriteFile(repoFile, 0644); err != nil {
				return nil, fmt.Errorf("failed to write repo file: %w", err)
			}

			// Step 2: Update repos
			providers := getter.All(settings)

			chartRepo, err := repo.NewChartRepository(repoEntry, providers)
			if err != nil {
				return nil, fmt.Errorf("failed to create chart repo: %w", err)
			}

			if _, err := chartRepo.DownloadIndexFile(); err != nil {
				return nil, fmt.Errorf("failed to download repo index: %w", err)
			}

			// Step 3: Install chart
			kubeConfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
			flags := genericclioptions.NewConfigFlags(false)
			flags.KubeConfig = &kubeConfig
			flags.Namespace = &project

			actionConfig := new(action.Configuration)
			if err := actionConfig.Init(flags, project, "secrets", func(format string, v ...interface{}) {}); err != nil {
				return nil, fmt.Errorf("failed to init helm config: %w", err)
			}

			install := action.NewInstall(actionConfig)
			install.ReleaseName = releaseName
			install.Namespace = project
			install.CreateNamespace = true

			chartPath, err := install.LocateChart(chart, settings)
			if err != nil {
				return nil, fmt.Errorf("failed to locate chart: %w", err)
			}

			ch, err := loader.Load(chartPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load chart: %w", err)
			}

			rel, err := install.Run(ch, nil)
			if err != nil {
				return nil, fmt.Errorf("helm install failed: %w", err)
			}

			return mcp.NewToolResultText(fmt.Sprintf("Helm release '%s' installed successfully in project '%s'", rel.Name, rel.Namespace)), nil
		},
	}
}

func HelmUninstallChart(config *rest.Config) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("helm_uninstall_chart",
			mcp.WithDescription("Uninstall a Helm release from a Kubernetes namespace."),
			mcp.WithString("release_name", mcp.Description("Name of the Helm release to uninstall"), mcp.Required()),
			mcp.WithString("project", mcp.Description("Namespace where the release is installed"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			releaseName := request.Params.Arguments["release_name"].(string)
			project := request.Params.Arguments["project"].(string)

			kubeConfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
			flags := genericclioptions.NewConfigFlags(false)
			flags.KubeConfig = &kubeConfig
			flags.Namespace = &project

			actionConfig := new(action.Configuration)
			if err := actionConfig.Init(flags, project, "secrets", func(format string, v ...interface{}) {}); err != nil {
				return nil, fmt.Errorf("failed to init helm config: %w", err)
			}

			uninstall := action.NewUninstall(actionConfig)
			resp, err := uninstall.Run(releaseName)
			if err != nil {
				return nil, fmt.Errorf("helm uninstall failed: %w", err)
			}

			return mcp.NewToolResultText(fmt.Sprintf("Helm release '%s' uninstalled from project '%s'. Info: %s", releaseName, project, resp.Info)), nil
		},
	}
}
