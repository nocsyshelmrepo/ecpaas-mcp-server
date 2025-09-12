package cluster

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func GetNode(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_node", mcp.WithDescription(`
Get the specified node by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("nodeName", mcp.Description("the given nodeName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			nodeName := request.Params.Arguments["nodeName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(v1alpha3.ResourcesGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("nodes").Name(nodeName).Do(ctx).Raw()
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

func GetProject(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_project", mcp.WithDescription(`
Get the specified project by name. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/namespaces/%s", project)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetDeployment(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_deployment", mcp.WithDescription(`
Get the specified deployment by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("deploymentName", mcp.Description("the given deploymentName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			deploymentName := request.Params.Arguments["deploymentName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", project, deploymentName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetStatefulset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_statefulset", mcp.WithDescription(`
Get a specific statefulset by name and project. The response will include:
- statefulsetName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("statefulsetName", mcp.Description("the given statefulsetName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			statefulsetName := request.Params.Arguments["statefulsetName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/apps/v1/namespaces/%s/statefulsets/%s", project, statefulsetName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetDaemonset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_daemonset", mcp.WithDescription(`
Get the specified daemonset by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("daemonsetName", mcp.Description("the given daemonsetName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			daemonsetName := request.Params.Arguments["daemonsetName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/apps/v1/namespaces/%s/daemonsets/%s", project, daemonsetName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetJob(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_job", mcp.WithDescription(`
Get the specified job by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("jobName", mcp.Description("the given jobName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			jobName := request.Params.Arguments["jobName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "batch", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/batch/v1/namespaces/%s/jobs/%s", project, jobName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "batch", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/batch/v1/cronjobs")
			data, err := client.Get().RequestURI(uri).Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetCronjob(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_cronjob", mcp.WithDescription(`
Get the specified cronjob in the given project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("cronjobName", mcp.Description("the given cronjobName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			cronjobName := request.Params.Arguments["cronjobName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "batch", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/batch/v1/namespaces/%s/cronjobs/%s", project, cronjobName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetPod(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_pod", mcp.WithDescription(`
Get the specified pod by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("podName", mcp.Description("the given podName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			podName := request.Params.Arguments["podName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s", project, podName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetService(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_service", mcp.WithDescription(`
Get a specific service by name and project. The response will include:
- serviceName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("serviceName", mcp.Description("the given serviceName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			serviceName := request.Params.Arguments["serviceName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/namespaces/%s/services/%s", project, serviceName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetIngress(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_ingress", mcp.WithDescription(`
Get the specified ingress by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("ingressName", mcp.Description("the given ingressName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			ingressName := request.Params.Arguments["ingressName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "networking.k8s.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/networking.k8s.io/v1/namespaces/%s/ingresses/%s", project, ingressName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetSecret(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_secret", mcp.WithDescription(`
Get a specific secret by name and project. The response will include:
- secretName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("secretName", mcp.Description("the given secretName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			secretName := request.Params.Arguments["secretName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/namespaces/%s/secrets/%s", project, secretName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetConfigmap(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_configmap", mcp.WithDescription(`
Get a specific configmap by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("configmapName", mcp.Description("the given configmapName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			configmapName := request.Params.Arguments["configmapName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/namespaces/%s/configmaps/%s", project, configmapName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetServiceaccount(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_serviceaccount", mcp.WithDescription(`
Get a specific serviceaccount by name and project. The response will include:
- serviceaccountName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("serviceaccountName", mcp.Description("the given serviceaccountName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			serviceaccountName := request.Params.Arguments["serviceaccountName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/namespaces/%s/serviceaccounts/%s", project, serviceaccountName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetCustomResourceDefinition(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_customresourcedefinition", mcp.WithDescription(`
Get a specific customresourcedefinition by name. The response will include:
- customresourcedefinitionName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("customresourcedefinitionName", mcp.Description("the given customresourcedefinitionName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			customresourcedefinitionName := request.Params.Arguments["customresourcedefinitionName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apiextensions.k8s.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/apiextensions.k8s.io/v1/customresourcedefinitions/%s", customresourcedefinitionName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
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

func GetPersistentvolumeclaim(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_persistentvolumeclaim", mcp.WithDescription(`
Get the specified persistentvolumeclaim by name and project. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("persistentvolumeclaimName", mcp.Description("the given persistentvolumeclaimName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			persistentvolumeclaimName := request.Params.Arguments["persistentvolumeclaimName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/namespaces/%s/persistentvolumeclaims/%s", project, persistentvolumeclaimName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
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

func GetPersistentvolume(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_persistentvolume", mcp.WithDescription(`
Get the specified persistentvolume by name. The response will include:
- configmapName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("persistentvolumeName", mcp.Description("the given persistentvolumeName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			persistentvolumeName := request.Params.Arguments["persistentvolumeName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/api/v1/persistentvolumes/%s", persistentvolumeName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
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
			mcp.WithString("cluster", mcp.Description("the given clusterName")),
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
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "resources.kubesphere.io", Version: "v1alpha3"}, cluster)
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

func GetStorageclass(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_storageclass", mcp.WithDescription(`
Get a specific storageclass by name. The response will include:
- storageclassName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - app: belong to which app
 - app.kubernetes.io/managed-by: which tool manages the Kubernetes resources.
 - chart: belong to which Helm chart and version.
 - heritage: which tool created the resource
 - release: belong to which Helm release name
- specific metadata.annotations fields indicate:
 - meta.helm.sh/release-name: which Helm release create and manages the kubernetes resource
 - meta.helm.sh/release-namespace: which namespace where the Helm release is installed
`),
			mcp.WithString("storageclassName", mcp.Description("the given storageclassName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			storageclassName := request.Params.Arguments["storageclassName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "storage.k8s.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/storage.k8s.io/v1/storageclasses/%s", storageclassName)
			data, err := client.Get().RequestURI(uri).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func CreateDeployment(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("create_deployment", mcp.WithDescription(`
Create a new deployment in the specified project and cluster.

Required parameters:
- cluster: the cluster name
- project: the Kubernetes project
- manifest: raw deployment manifest in JSON format
`),
			mcp.WithString("project", mcp.Description("the Kubernetes project"), mcp.Required()),
			mcp.WithString("manifest", mcp.Description("deployment spec in JSON"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Extract parameters
			project := request.Params.Arguments["project"].(string)
			rawManifest := request.Params.Arguments["manifest"].(string)

			// Parse manifest string to unstructured object
			unstructuredObj := &unstructured.Unstructured{}
			if err := json.Unmarshal([]byte(rawManifest), &unstructuredObj.Object); err != nil {
				return nil, fmt.Errorf("failed to parse manifest: %v", err)
			}

			// Get Kubernetes client for apps/v1
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments", project)

			// Send POST request to create the deployment
			data, err := client.Post().RequestURI(uri).Body(unstructuredObj).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Deployment created: %s", string(data))), nil
		},
	}
}

func DeleteDeployment(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_deployment", mcp.WithDescription(`
Delete a deployment in the specified project and cluster.

Required parameters:
- project: the Kubernetes project
- deployment: the name of the deployment to delete
`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("deployment", mcp.Description("the deployment name to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Extract parameters
			project := request.Params.Arguments["project"].(string)
			deployment := request.Params.Arguments["deployment"].(string)

			// Get Kubernetes client for the apps/v1 group (standard deployments)
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			// Construct full URI path manually
			uri := fmt.Sprintf("/apis/apps/v1/namespaces/%s/deployments/%s", project, deployment)

			// Send DELETE request to Kubernetes API
			err = client.Delete().RequestURI(uri).Do(ctx).Error()

			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Deployment '%s' in project '%s' deleted successfully.", deployment, project)), nil
		},
	}
}
