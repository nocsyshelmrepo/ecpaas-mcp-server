package cluster

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	storagev1 "k8s.io/api/storage/v1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/constants/v1alpha3"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
)

func ListNodes(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_nodes", mcp.WithDescription(`
Retrieve the paginated nodes list. The response will include:
1. items: An array of nodes objects containing:
the item actual is node resource in kubernetes.
  - specific metadata.labels fields indicate:
   - node-role.kubernetes.io/edge: is the edge node.
2. totalItems: The total number of nodes in KubeSphere.		
`),
			mcp.WithNumber("limit", mcp.Description("Number of nodes displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of nodes to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("nodes").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func ListProjects(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_projects", mcp.WithDescription(`
Get project from cluster or workspace.
when Get from cluster. should set cluster param.
when Get from workspace. should set workspace and cluster which has assign to this workspace.
Retrieve the paginated projects list. The response will include:
1. items: An array of projects objects containing:
the item actual is namespace resource in kubernetes.
  - specific metadata.labels fields indicate:
   - node-role.kubernetes.io/edge: is the edge node.
2. totalItems: The total number of projects in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of projects displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of projects to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("workspace", mcp.Description("the given workspaceName")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			workspace := ""
			if reqWorkspace, ok := request.Params.Arguments["workspace"].(string); ok && reqWorkspace != "" {
				workspace = reqWorkspace
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			switch workspace {
			case "": // from cluster
				client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
				if err != nil {
					return nil, err
				}
				data, err := client.Get().Resource("namespaces").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default: // from workspace
				client, err := ksconfig.RestClient(tenantv1beta1.SchemeGroupVersion, cluster)
				if err != nil {
					return nil, err
				}
				data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).SubResource("namespaces").
					Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListDeployments(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_deployments", mcp.WithDescription(`
Retrieve the paginated deployments list. The response will include:
1. items: An array of deployments objects containing:
the item actual is deployments resource in kubernetes.
2. totalItems: The total number of deployments in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of deployments displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of deployments to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project deployments")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("deployments").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("deployments").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListStatefulsets(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_statefulsets", mcp.WithDescription(`
Retrieve the paginated statefulsets list. The response will include:
1. items: An array of statefulsets objects containing:
the item actual is statefulsets resource in kubernetes.
2. totalItems: The total number of statefulsets in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of statefulsets displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of statefulsets to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project statefulsets")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("statefulsets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("statefulsets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListDaemonsets(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_daemonsets", mcp.WithDescription(`
Retrieve the paginated daemonsets list. The response will include:
1. items: An array of daemonsets objects containing:
the item actual is daemonsets resource in kubernetes.
2. totalItems: The total number of daemonsets in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of daemonsets displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of daemonsets to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project daemonsets")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("daemonsets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("daemonsets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListJobs(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_jobs", mcp.WithDescription(`
Retrieve the paginated jobs list. The response will include:
1. items: An array of jobs objects containing:
the item actual is jobs resource in kubernetes.
2. totalItems: The total number of jobs in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of jobs displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of jobs to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project jobs")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("jobs").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("jobs").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListCronJobs(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_cronjobs", mcp.WithDescription(`
Retrieve the paginated cronjobs list. The response will include:
1. items: An array of cronjobs objects containing:
the item actual is cronjobs resource in kubernetes.
2. totalItems: The total number of cronjobs in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of cronjobs displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of cronjobs to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project cronjobs")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("cronjobs").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("cronjobs").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListPods(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_pods", mcp.WithDescription(`
Retrieve the paginated pods list. The response will include:
1. items: An array of pods objects containing:
the item actual is pods resource in kubernetes.
2. totalItems: The total number of pods in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of pods displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of pods to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project pods")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("pods").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("pods").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListServices(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_services", mcp.WithDescription(`
Retrieve the paginated services list. The response will include:
1. items: An array of services objects containing:
the item actual is services resource in kubernetes.
2. totalItems: The total number of services in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of services displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of services to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project services")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("services").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("services").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListIngresses(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_ingresses", mcp.WithDescription(`
Retrieve the paginated ingresses list. The response will include:
1. items: An array of ingresses objects containing:
the item actual is ingresses resource in kubernetes.
2. totalItems: The total number of ingresses in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of ingresses displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of ingresses to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project ingresses")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("ingresses").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("ingresses").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListSecrets(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_secrets", mcp.WithDescription(`
Retrieve the paginated secrets list. The response will include:
1. items: An array of secrets objects containing:
the item actual is secrets resource in kubernetes.
2. totalItems: The total number of secrets in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of secrets displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of secrets to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project secrets")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("secrets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("secrets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListConfigmaps(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_configmaps", mcp.WithDescription(`
Retrieve the paginated configmaps list. The response will include:
1. items: An array of configmaps objects containing:
the item actual is configmaps resource in kubernetes.
2. totalItems: The total number of configmaps in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of configmaps displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of configmaps to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project configmaps")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("configmaps").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("configmaps").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListServiceAccounts(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_serviceaccounts", mcp.WithDescription(`
Retrieve the paginated serviceaccounts list. The response will include:
1. items: An array of serviceaccounts objects containing:
the item actual is serviceaccounts resource in kubernetes.
2. totalItems: The total number of serviceaccounts in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of serviceaccounts displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of serviceaccounts to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project serviceaccounts")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("serviceaccounts").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("serviceaccounts").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListCustomResourceDefinitions(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_customresourcedefinitions", mcp.WithDescription(`
Retrieve the paginated customresourcedefinitions list. The response will include:
1. items: An array of customresourcedefinitions objects containing:
the item actual is customresourcedefinitions resource in kubernetes.
2. totalItems: The total number of customresourcedefinitions in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of customresourcedefinitions displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of customresourcedefinitions to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("customresourcedefinitions").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func ListPersistentVolumeClaims(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_persistentvolumeclaims", mcp.WithDescription(`
Retrieve the paginated persistentvolumeclaims list. The response will include:
1. items: An array of persistentvolumeclaims objects containing:
the item actual is persistentvolumeclaims resource in kubernetes.
2. totalItems: The total number of persistentvolumeclaims in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of persistentvolumeclaims displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of persistentvolumeclaims to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project persistentvolumeclaims")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := ""
			if reqProject, ok := request.Params.Arguments["project"].(string); ok && reqProject != "" {
				project = reqProject
			}
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("persistentvolumeclaims").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("persistentvolumeclaims").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func ListPersistentVolumes(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_persistentvolumes", mcp.WithDescription(`
Retrieve the paginated persistentvolumes list. The response will include:
1. items: An array of persistentvolumes objects containing:
the item actual is persistentvolumes resource in kubernetes.
2. totalItems: The total number of persistentvolumes in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of persistentvolumes displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of persistentvolumes to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("persistentvolumes").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func ListStorageClasses(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_storageclasses", mcp.WithDescription(`
Retrieve the paginated storageclasses list. The response will include:
1. items: An array of storageclasses objects containing:
the item actual is storageclasses resource in kubernetes.
2. totalItems: The total number of storageclasses in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of storageclasses displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of storageclasses to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(storagev1.SchemeGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("storageclasses").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
