package app

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"kubesphere.io/ks-mcp-server/cmd/app/options"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
	"kubesphere.io/ks-mcp-server/pkg/tools/cluster"
	"kubesphere.io/ks-mcp-server/pkg/tools/userrole"
	"kubesphere.io/ks-mcp-server/pkg/tools/workspace"
	"kubesphere.io/ks-mcp-server/pkg/version"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ks-mcp-server",
		Short:        "KubeSphere MCP Server",
		Version:      fmt.Sprint(version.Get()),
		SilenceUsage: true,
	}

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newStudioCommand())
	cmd.AddCommand(newSSECommand())

	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of KubeSphere MCP Server",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(version.Get())
		},
	}
}

func newStudioCommand() *cobra.Command {
	o := options.NewStdioOptions()
	cmd := &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// get restconfig
			restconfig, err := kubesphere.LoadKSConfig(o.KSConfig)
			if err != nil {
				return err
			}
			ksconfig, err := kubesphere.NewKSConfig(restconfig, o.KSApiserver)
			if err != nil {
				return err
			}
			// create MCP server
			s := server.NewMCPServer(
				"KubeSphere",
				"4.1.3",
			)
			s.SetTools(
				// register userrole
				userrole.ListUsers(ksconfig), userrole.GetUser(ksconfig),
				userrole.ListRoles(ksconfig), userrole.GetRole(ksconfig), userrole.ListPermissions(ksconfig),
				// register workspace
				workspace.ListWorkspaces(ksconfig), workspace.GetWorkspace(ksconfig), workspace.ListWorkspaceMembers(ksconfig), workspace.GetWorkspaceQuotas(ksconfig),
				workspace.ListApplicationRepos(ksconfig), workspace.ListApplications(ksconfig), workspace.GetApplication(ksconfig), workspace.GetApplicationVersion(ksconfig),
				workspace.GetWorkspace(ksconfig), workspace.ListProjectMembers(ksconfig),
				// register cluster
				cluster.ListClusters(ksconfig), cluster.GetCluster(ksconfig), cluster.ListClusterMembers(ksconfig),
				cluster.ListNodes(ksconfig), cluster.ListProjects(ksconfig), cluster.ListDeployments(ksconfig), cluster.ListStatefulsets(ksconfig), cluster.ListDaemonsets(ksconfig),
				cluster.ListJobs(ksconfig), cluster.ListCronJobs(ksconfig), cluster.ListPods(ksconfig), cluster.ListServices(ksconfig), cluster.ListIngresses(ksconfig),
				cluster.ListSecrets(ksconfig), cluster.ListConfigmaps(ksconfig), cluster.ListServiceAccounts(ksconfig), cluster.ListCustomResourceDefinitions(ksconfig),
				cluster.ListPersistentVolumeClaims(ksconfig), cluster.ListPersistentVolumes(ksconfig), cluster.ListStorageClasses(ksconfig),
				cluster.CreateDeployment(ksconfig), cluster.DeleteDeployment(ksconfig),
				// register extension

			)

			// Start the stdio server
			return server.ServeStdio(s)
		},
	}

	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}
	return cmd
}

func newSSECommand() *cobra.Command {
	o := options.NewStdioOptions()
	cmd := &cobra.Command{
		Use:   "sse",
		Short: "Start sse server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// get restconfig
			restconfig, err := kubesphere.LoadKSConfig(o.KSConfig)
			if err != nil {
				return err
			}
			ksconfig, err := kubesphere.NewKSConfig(restconfig, o.KSApiserver)
			if err != nil {
				return err
			}
			// create MCP server
			s := server.NewMCPServer(
				"KubeSphere",
				"4.1.3",
			)
			s.SetTools(
				// register userrole
				userrole.ListUsers(ksconfig), userrole.GetUser(ksconfig),
				userrole.ListRoles(ksconfig), userrole.GetRole(ksconfig), userrole.ListPermissions(ksconfig),
				// register workspace
				workspace.ListWorkspaces(ksconfig), workspace.GetWorkspace(ksconfig), workspace.ListWorkspaceMembers(ksconfig), workspace.GetWorkspaceQuotas(ksconfig),
				workspace.ListApplicationRepos(ksconfig), workspace.ListApplications(ksconfig), workspace.GetApplication(ksconfig), workspace.GetApplicationVersion(ksconfig),
				workspace.GetWorkspace(ksconfig), workspace.ListProjectMembers(ksconfig),
				// register cluster
				cluster.ListClusters(ksconfig), cluster.GetCluster(ksconfig), cluster.ListClusterMembers(ksconfig),
				cluster.ListNodes(ksconfig), cluster.ListProjects(ksconfig), cluster.ListDeployments(ksconfig), cluster.ListStatefulsets(ksconfig), cluster.ListDaemonsets(ksconfig),
				cluster.ListJobs(ksconfig), cluster.ListCronJobs(ksconfig), cluster.ListPods(ksconfig), cluster.ListServices(ksconfig), cluster.ListIngresses(ksconfig),
				cluster.ListSecrets(ksconfig), cluster.ListConfigmaps(ksconfig), cluster.ListServiceAccounts(ksconfig), cluster.ListCustomResourceDefinitions(ksconfig),
				cluster.ListPersistentVolumeClaims(ksconfig), cluster.ListPersistentVolumes(ksconfig), cluster.ListStorageClasses(ksconfig),
				cluster.CreateDeployment(ksconfig), cluster.DeleteDeployment(ksconfig),
				// register extension

			)
			// Start the sse server
			sse := server.NewSSEServer(s)
			return sse.Start(":8080")
		},
	}

	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}
	return cmd
}
