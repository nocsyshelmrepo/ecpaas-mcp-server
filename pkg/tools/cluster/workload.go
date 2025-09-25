package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
	k8sCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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

func DeleteNode(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_node", mcp.WithDescription(`Delete the specified node by cluster and nodeName.`),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("nodeName", mcp.Description("the given nodeName to delete"), mcp.Required()),
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
			err = client.Delete().Resource("nodes").Name(nodeName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Node '%s' was deleted successfully.", nodeName)), nil
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
			data, err := client.Get().Namespace(project).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteProject(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_project", mcp.WithDescription(`Delete the specified project by name.`),
			mcp.WithString("project", mcp.Description("the given project to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Namespace(project).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Project '%s' was deleted successfully.", project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("deployments", deploymentName).Do(ctx).Raw()
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
			// deal request params
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
			data, err := client.Post().Namespace(project).Resource("deployments").Body(unstructuredObj).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Deployment created: %s", string(data))), nil
		},
	}
}

func DeleteDeployment(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_deployment", mcp.WithDescription(`Delete a specified deployment by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("deployment", mcp.Description("the given deployment name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("deployments", deploymentName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Deployment '%s' in project '%s' was deleted successfully.", deploymentName, project)), nil
		},
	}
}

func ScaleDeployment(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("scale_deployment", mcp.WithDescription(`scale a specified deployment by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("deploymentName", mcp.Description("the given deployment name to scale"), mcp.Required()),
			mcp.WithNumber("replicas", mcp.Description("number of replicas to scale"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			deploymentName := request.Params.Arguments["deploymentName"].(string)

			// `replicas` comes as float64, convert to int32
			replicasFloat := request.Params.Arguments["replicas"].(float64)
			replicas := int32(replicasFloat)

			// Create patch body
			patch := fmt.Sprintf(`{"spec":{"replicas":%d}}`, replicas)
			patchBytes := []byte(patch)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Patch(types.MergePatchType).Namespace(project).Resource("deployments").
				Name(deploymentName).Body(patchBytes).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Deployment '%s' in project '%s' was scaled successfully.", deploymentName, project)), nil
		},
	}
}

func RolloutDeployment(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("rollout_deployment", mcp.WithDescription(`Rollout a specified deployment by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("deploymentName", mcp.Description("the given deployment name to rollout"), mcp.Required()),
			mcp.WithString("action", mcp.Description("rollout subcommand, only two subcommands (restart) or (undo)"), mcp.Required()),
			mcp.WithString("templateSpec", mcp.Description("The desired spec.template of the Deployment in JSON format. This typically comes from a historical Replicaset. Required only when initiating a rollout undo operation.")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			deploymentName := request.Params.Arguments["deploymentName"].(string)
			action := ""
			if reqAction, ok := request.Params.Arguments["action"].(string); ok &&
				(reqAction == constants.RolloutRestart || reqAction == constants.RolloutUndo) {
				action = reqAction
			}
			switch action {
			case constants.RolloutRestart:
				// Prepare patch to update annotation "kubectl.kubernetes.io/restartedAt" with current timestamp
				currentTime := time.Now().UTC().Format(time.RFC3339)
				patchBytes := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, currentTime))

				// deal http request
				client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
				if err != nil {
					return nil, err
				}
				err = client.Patch(types.MergePatchType).Namespace(project).Suffix("deployments", deploymentName).Body(patchBytes).Do(ctx).Error()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(fmt.Sprintf("Deployment '%s' in project '%s' was rollout restarted successfully.", deploymentName, project)), nil
			case constants.RolloutUndo:
				templateSpecStr, ok := request.Params.Arguments["templateSpec"].(string)
				if !ok || templateSpecStr == "" {
					return nil, errors.New("missing or invalid parameter: 'templateSpec'")
				}

				// 2. Unmarshal the input template spec string into a v1.PodTemplateSpec object
				var templateSpec k8sCoreV1.PodTemplateSpec
				if err := json.Unmarshal([]byte(templateSpecStr), &templateSpec); err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal templateSpec parameter")
				}

				// 3. Create the JSON Patch payload
				// We are performing a strategic merge patch to update the spec.template
				patchPayload := map[string]interface{}{
					"spec": map[string]interface{}{
						"template": templateSpec,
					},
				}

				patchBytes, err := json.Marshal(patchPayload)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal patch payload")
				}

				// 4. Get the KubeSphere RestClient
				client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
				if err != nil {
					return nil, err
				}

				// 5. Perform the PATCH request
				// The URL path will be like: /apis/apps/v1/namespaces/{namespace}/deployments/{name}
				data, err := client.Patch(types.StrategicMergePatchType).Namespace(project).Resource("deployments").Name(deploymentName).Body(patchBytes).Do(ctx).Raw()

				if err != nil {
					return nil, errors.Wrapf(err, "failed to patch deployment %s/%s", project, deploymentName)
				}

				return mcp.NewToolResultText(fmt.Sprintf("Deployment '%s' successfully rollout undone. Response: %s", deploymentName, string(data))), nil
			default:
				return nil, errors.Errorf("Unsupport action, it should be one of %s", strings.Join([]string{constants.RolloutRestart, constants.RolloutUndo}, ","))
			}

		},
	}
}

func ListReplicasets(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_replicasets", mcp.WithDescription(`
Retrieve the paginated replicasets list. The response will include:
1. items: An array of replicasets objects containing:
the item actual is replicasets resource in kubernetes.
2. totalItems: The total number of replicasets in KubeSphere.	
`),
			mcp.WithNumber("limit", mcp.Description("Number of replicasets displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of replicasets to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName, if empty will return all project replicasets")),
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
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, cluster)
			if err != nil {
				return nil, err
			}
			switch project {
			case "":
				data, err := client.Get().Resource("replicasets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				data, err := client.Get().Namespace(project).Resource("replicasets").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			}
		},
	}
}

func GetReplicaset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_replicaset", mcp.WithDescription(`
Get a specific replicaset by name and project. The response will include:
- replicasetName: Maps to metadata.name
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
			mcp.WithString("replicasetName", mcp.Description("the given replicasetName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			replicasetName := request.Params.Arguments["replicasetName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Namespace(project).Suffix("replicasets", replicasetName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteReplicaset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_replicaset", mcp.WithDescription(`Delete a specified replicaset by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("replicasetName", mcp.Description("the given replicasetName to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			replicasetName := request.Params.Arguments["replicasetName"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Namespace(project).Suffix("replicasets", replicasetName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Replicaset '%s' in project '%s' was deleted successfully.", replicasetName, project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("statefulsets", statefulsetName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteStatefulset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_statefulset", mcp.WithDescription(`Delete a specified statefulset by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("statefulsetName", mcp.Description("the given statefulsetName to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("statefulsets", statefulsetName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Statefulset '%s' in project '%s' was deleted successfully.", statefulsetName, project)), nil
		},
	}
}

func ScaleStatefulset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("scale_statefulset", mcp.WithDescription(`scale a specified statefulset by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("statefulsetName", mcp.Description("the given statefulset name to scale"), mcp.Required()),
			mcp.WithNumber("replicas", mcp.Description("number of replicas"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			statefulsetName := request.Params.Arguments["statefulsetName"].(string)

			// `replicas` comes as float64, convert to int32
			replicasFloat := request.Params.Arguments["replicas"].(float64)
			replicas := int32(replicasFloat)

			// Create patch body
			patch := fmt.Sprintf(`{"spec":{"replicas":%d}}`, replicas)
			patchBytes := []byte(patch)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Patch(types.MergePatchType).Namespace(project).Resource("statefulsets").
				Name(statefulsetName).Body(patchBytes).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Statefulset '%s' in project '%s' was scaled successfully.", statefulsetName, project)), nil
		},
	}
}

func RolloutStatefulset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("rollout_statefulset", mcp.WithDescription(`Rollout a specified statefulset by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("statefulset", mcp.Description("the given statefulset name to rollout"), mcp.Required()),
			mcp.WithString("action", mcp.Description("rollout subcommand, only two subcommands (restart) or (undo)"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			statefulsetName := request.Params.Arguments["statefulsetName"].(string)
			action := ""
			if reqAction, ok := request.Params.Arguments["action"].(string); ok &&
				(reqAction == constants.RolloutRestart || reqAction == constants.RolloutUndo) {
				action = reqAction
			}
			switch action {
			case constants.RolloutRestart:
				// Prepare patch to update annotation "kubectl.kubernetes.io/restartedAt" with current timestamp
				currentTime := time.Now().UTC().Format(time.RFC3339)
				patchBytes := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, currentTime))

				// deal http request
				client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
				if err != nil {
					return nil, err
				}
				err = client.Patch(types.MergePatchType).Namespace(project).Suffix("statefulsets", statefulsetName).Body(patchBytes).Do(ctx).Error()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(fmt.Sprintf("Statefulset '%s' in project '%s' was rollout restarted successfully.", statefulsetName, project)), nil
			case constants.RolloutUndo:
				// deal http request
				client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
				if err != nil {
					return nil, err
				}
				err = client.Patch(types.MergePatchType).Namespace(project).Suffix("statefulsets", statefulsetName).Do(ctx).Error()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(fmt.Sprintf("Statefulset '%s' in project '%s' was rollout undoed successfully.", statefulsetName, project)), nil
			default:
				return nil, errors.Errorf("Unsupport action, it should be one of %s", strings.Join([]string{constants.RolloutRestart, constants.RolloutUndo}, ","))
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
			data, err := client.Get().Namespace(project).Suffix("daemonsets", daemonsetName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteDaemonset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_daemonset", mcp.WithDescription(`Delete a specified daemonset by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("daemonsetName", mcp.Description("the given daemonsetName to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("daemonsets", daemonsetName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Daemonset '%s' in project '%s' was deleted successfully.", daemonsetName, project)), nil
		},
	}
}

func RolloutDaemonset(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("rollout_daemonset", mcp.WithDescription(`Rollout a specified daemonset by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("daemonset", mcp.Description("the given daemonset name to rollout"), mcp.Required()),
			mcp.WithString("action", mcp.Description("rollout subcommand, only two subcommands (restart) or (undo)"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			daemonsetName := request.Params.Arguments["daemonsetName"].(string)
			action := ""
			if reqAction, ok := request.Params.Arguments["action"].(string); ok &&
				(reqAction == constants.RolloutRestart || reqAction == constants.RolloutUndo) {
				action = reqAction
			}
			switch action {
			case constants.RolloutRestart:
				// Prepare patch to update annotation "kubectl.kubernetes.io/restartedAt" with current timestamp
				currentTime := time.Now().UTC().Format(time.RFC3339)
				patchBytes := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, currentTime))

				// deal http request
				client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
				if err != nil {
					return nil, err
				}
				err = client.Patch(types.MergePatchType).Namespace(project).Suffix("daemonsets", daemonsetName).Body(patchBytes).Do(ctx).Error()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(fmt.Sprintf("Daemonset '%s' in project '%s' was rollout restarted successfully.", daemonsetName, project)), nil
			case constants.RolloutUndo:
				// deal http request
				client, err := ksconfig.RestClient(schema.GroupVersion{Group: "apps", Version: "v1"}, "")
				if err != nil {
					return nil, err
				}
				err = client.Patch(types.MergePatchType).Namespace(project).Suffix("daemonsets", daemonsetName).Do(ctx).Error()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(fmt.Sprintf("Daemonset '%s' in project '%s' was rollout undoed successfully.", daemonsetName, project)), nil
			default:
				return nil, errors.Errorf("Unsupport action, it should be one of %s", strings.Join([]string{constants.RolloutRestart, constants.RolloutUndo}, ","))
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
			data, err := client.Get().Namespace(project).Suffix("jobs", jobName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteJob(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_job", mcp.WithDescription(`Delete a specified job by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("jobName", mcp.Description("the job name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("jobs", jobName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Job '%s' in project '%s' was deleted successfully.", jobName, project)), nil
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
			data, err := client.Get().Resource("cronjobs").Param("sortBy", "name").Param("limit", limit).Param("page", page).Do(ctx).Raw()
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
			data, err := client.Get().Namespace(project).Suffix("cronjobs", cronjobName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteCronjob(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_cronjob", mcp.WithDescription(`Delete a specified cronjob by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("cronjobName", mcp.Description("the cronjob name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("cronjobs", cronjobName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Cronjob '%s' in project '%s' was deleted successfully.", cronjobName, project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("pods", podName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func LogsPod(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("logs_pod", mcp.WithDescription(`show logs of a specified pod by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("podName", mcp.Description("the pod name to logs"), mcp.Required()),
			mcp.WithString("containerName", mcp.Description("container name (optional)")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			podName := request.Params.Arguments["podName"].(string)
			containerName := request.Params.Arguments["containerName"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Namespace(project).Suffix("pods", podName, "log").Param("container", containerName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func ExecPod(restConfig *rest.Config) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("exec_pod",
			mcp.WithDescription(`
Execute a command in the specified pod using Kubernetes exec API.
This tool executes commands in pods and returns the output directly.
Supports both interactive and non-interactive command execution.

Parameters:
- project: The namespace/project name
- podName: The target pod name  
- command: Command to execute (e.g., "/bin/bash", "ls -la")
- stdin: Enable stdin stream (true/false, default: false)
- stdout: Enable stdout stream (true/false, default: true)
- stderr: Enable stderr stream (true/false, default: true)
- tty: Enable TTY for interactive sessions (true/false, default: false)
- containerName: Specific container name (optional)
			`),
			mcp.WithString("project", mcp.Description("the project/namespace name"), mcp.Required()),
			mcp.WithString("podName", mcp.Description("the pod name"), mcp.Required()),
			mcp.WithString("command", mcp.Description("command to execute (e.g., '/bin/bash' or 'ls -la')"), mcp.Required()),
			mcp.WithString("stdin", mcp.Description("enable stdin stream (true/false, default: false)")),
			mcp.WithString("stdout", mcp.Description("enable stdout stream (true/false, default: true)")),
			mcp.WithString("stderr", mcp.Description("enable stderr stream (true/false, default: true)")),
			mcp.WithString("tty", mcp.Description("enable TTY for interactive sessions (true/false, default: false)")),
			mcp.WithString("containerName", mcp.Description("container name (optional)")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse request parameters
			project := request.Params.Arguments["project"].(string)
			podName := request.Params.Arguments["podName"].(string)
			commandStr := request.Params.Arguments["command"].(string)

			// Parse optional parameters with defaults
			stdinStr := getStringParamWithDefault(request.Params.Arguments, "stdin", "false")
			stdoutStr := getStringParamWithDefault(request.Params.Arguments, "stdout", "true")
			stderrStr := getStringParamWithDefault(request.Params.Arguments, "stderr", "true")
			ttyStr := getStringParamWithDefault(request.Params.Arguments, "tty", "false")
			containerName := ""
			if reqContainerName, ok := request.Params.Arguments["containerName"].(string); ok && reqContainerName != "" {
				containerName = reqContainerName
			}

			// Convert string parameters to booleans
			stdin := parseBool(stdinStr)
			stdout := parseBool(stdoutStr)
			stderr := parseBool(stderrStr)
			tty := parseBool(ttyStr)

			// Create Kubernetes client
			k8sClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
			}

			// Parse command string into slice
			command := strings.Fields(commandStr)

			// Prepare exec request
			req := k8sClient.CoreV1().RESTClient().Post().
				Resource("pods").
				Name(podName).
				Namespace(project).
				SubResource("exec").
				VersionedParams(&k8sCoreV1.PodExecOptions{
					Command:   command,
					Container: containerName,
					Stdin:     stdin,
					Stdout:    stdout,
					Stderr:    stderr,
					TTY:       tty,
				}, scheme.ParameterCodec)

			// Modify the URL scheme to use WebSocket
			execURL := req.URL()
			websocketURL := convertToWebSocketURL(execURL, restConfig)

			// Return result as structured content
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Text: fmt.Sprintf("WebSocket URL: %s", websocketURL)},
				},
			}, nil
		},
	}
}

// convertToWebSocketURL modifies the Kubernetes exec URL to use ws/wss
func convertToWebSocketURL(u *url.URL, config *rest.Config) string {
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}
	if config != nil && config.Host != "" {
		clusterURL, err := url.Parse(config.Host)
		if err == nil {
			u.Host = clusterURL.Host
		}
	}
	return u.String()
}

// Helper function to get string parameter with default value
func getStringParamWithDefault(args map[string]interface{}, key, defaultValue string) string {
	if val, exists := args[key]; exists {
		if strVal, ok := val.(string); ok && strVal != "" {
			return strVal
		}
	}
	return defaultValue
}

// Helper function to parse boolean from string
func parseBool(s string) bool {
	return strings.ToLower(s) == "true"
}

func DeletePod(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_pod", mcp.WithDescription(`Delete a specified pod by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("podName", mcp.Description("the pod name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("pods", podName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Pod '%s' in project '%s' was deleted successfully.", podName, project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("services", serviceName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteService(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_service", mcp.WithDescription(`Delete a specified service by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("serviceName", mcp.Description("the service name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("services", serviceName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Service '%s' in project '%s' was deleted successfully.", serviceName, project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("ingresses", ingressName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteIngress(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_ingress", mcp.WithDescription(`Delete a specified ingress by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("ingressName", mcp.Description("the ingress name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("ingresses", ingressName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Ingress '%s' in project '%s' was deleted successfully.", ingressName, project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("secrets", secretName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteSecret(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_secret", mcp.WithDescription(`Delete a specified secret by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("secretName", mcp.Description("the secret name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("secrets", secretName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Secret '%s' in project '%s' was deleted successfully.", secretName, project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("configmaps", configmapName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteConfigmap(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_configmap", mcp.WithDescription(`Delete a specified configmap by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("configmapName", mcp.Description("the configmap name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("configmaps", configmapName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Configmap '%s' in project '%s' was deleted successfully.", configmapName, project)), nil
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
			data, err := client.Get().Namespace(project).Suffix("serviceaccounts", serviceaccountName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteServiceaccount(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_serviceaccount", mcp.WithDescription(`Delete a specified serviceaccount by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("serviceaccountName", mcp.Description("the serviceaccount name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("serviceaccounts", serviceaccountName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Serviceaccount '%s' in project '%s' was deleted successfully.", serviceaccountName, project)), nil
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
			data, err := client.Get().Resource("customresourcedefinitions").Name(customresourcedefinitionName).Do(ctx).Raw()
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
			data, err := client.Get().Namespace(project).Suffix("persistentvolumeclaims", persistentvolumeclaimName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeletePersistentvolumeclaim(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_persistentvolumeclaim", mcp.WithDescription(`Delete a specified persistentvolumeclaim by name and project.`),
			mcp.WithString("project", mcp.Description("the Kubesphere project"), mcp.Required()),
			mcp.WithString("persistentvolumeclaimName", mcp.Description("the persistentvolumeclaim name to delete"), mcp.Required()),
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
			err = client.Delete().Namespace(project).Suffix("persistentvolumeclaims", persistentvolumeclaimName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Persistentvolumeclaim '%s' in project '%s' was deleted successfully.", persistentvolumeclaimName, project)), nil
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
			data, err := client.Get().Resource("persistentvolumes").Name(persistentvolumeName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeletePersistentvolume(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_persistentvolume", mcp.WithDescription(`Delete a specified persistentvolume by name.`),
			mcp.WithString("persistentvolumeName", mcp.Description("the persistentvolume name to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			persistentvolumeName := request.Params.Arguments["persistentvolumeName"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource("persistentvolumes").Name(persistentvolumeName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Persistentvolume '%s' was deleted successfully.", persistentvolumeName)), nil
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
			data, err := client.Get().Resource("storageclasses").Name(storageclassName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteStorageclass(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_storageclass", mcp.WithDescription(`Delete a specified storageclass by name.`),
			mcp.WithString("storageclassName", mcp.Description("the storageclass name to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			storageclassName := request.Params.Arguments["storageclassName"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "storage.k8s.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource("storageclasses").Name(storageclassName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Storageclass '%s' was deleted successfully.", storageclassName)), nil
		},
	}
}
