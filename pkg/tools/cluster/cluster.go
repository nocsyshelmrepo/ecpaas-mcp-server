package cluster

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
)

func ListClusters(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_clusters", mcp.WithDescription(`
Retrieve the paginated clusters list. The response will include:
1. items: An array of clusters objects containing:
  - clusterName: Maps to metadata.name
  - specific metadata.labels fields indicate:
   - cluster-role.kubesphere.io/host: the cluster role is host.
   - cluster-role.kubesphere.io/edge: the cluster is edge cluster which deploy in edge node.
   - label.cluster.kubesphere.io/xhn72r: the xhn72r of label is the tags of cluster.
2. totalItems: The total number of clusters in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of clusters displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of clusters to display. Default is "+constants.DefPage)),
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
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "resources.kubesphere.io", Version: "v1alpha3"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(clusterv1alpha1.ResourcesPluralCluster).Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetCluster(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_cluster", mcp.WithDescription(`
Get cluster information. The response will contain:
- clusterName: Maps metadata.name
- specific metadata.labels fields indicate:
 - cluster-role.kubesphere.io/host: the cluster role is host.
 - cluster-role.kubesphere.io/edge: the cluster is edge cluster which deploy in edge node.
 - label.cluster.kubesphere.io/xhn72r: the xhn72r of label_id is the tags of cluster.
`),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "resources.kubesphere.io", Version: "v1alpha3"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(clusterv1alpha1.ResourcesPluralCluster).Name(cluster).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteCluster(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_cluster", mcp.WithDescription(`Delete the specified cluster by name.`),
			mcp.WithString("cluster", mcp.Description("the given clusterName to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "resources.kubesphere.io", Version: "v1alpha3"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource(clusterv1alpha1.ResourcesPluralCluster).Name(cluster).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Cluster '%s' was deleted successfully.", cluster)), nil
		},
	}
}

// func GetClusterTags(ksconfig *kubesphere.KSConfig) server.ServerTool {
// 	return server.ServerTool{
// 		Tool: mcp.NewTool("get_cluster_tags", mcp.WithDescription(`
// Retrieve the paginated cluster tags map. The response will include:
// - map key is label key
// - map value is a value array.
//  - value: is the actual value of key.
//  - id: is the label_id which show in cluster information.
// `)),
// 		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 			// deal http request
// 			client, err := ksconfig.RestClient(clusterv1alpha1.SchemeGroupVersion, "")
// 			if err != nil {
// 				return nil, err
// 			}
// 			data, err := client.Get().Resource("labels").Do(ctx).Raw()
// 			if err != nil {
// 				return nil, err
// 			}

// 			return mcp.NewToolResultText(string(data)), nil
// 		},
// 	}
// }

func ListClusterMembers(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_cluster_members", mcp.WithDescription(`
Retrieve the paginated project members list. The response will include:
1. items: An array of project member objects in project containing:
  - username: Maps to metadata.name
  - specific metadata.annotations fields indicate:
   - iam.kubesphere.io/clusterrole: the cluster role which this user belong to.
2. totalItems: The total number of project members in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of project members displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of project members to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			// limit := constants.DefLimit
			// if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
			// 	limit = fmt.Sprintf("%d", reqLimit)
			// }
			// page := constants.DefPage
			// if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
			// 	page = fmt.Sprintf("%d", reqPage)
			// }
			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("clustermembers").Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetClustermember(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_cluster_member", mcp.WithDescription(`
Get a specific cluster member information by name. The response will include:
- workspaceMemberName: Maps to metadata.name
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
			mcp.WithString("clusterMemberName", mcp.Description("the given clusterMemberName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			clusterMemberName := request.Params.Arguments["clusterMemberName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("clustermembers").Name(clusterMemberName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteClusterMember(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_cluster_member", mcp.WithDescription(`Delete a specific cluster member by name.`),
			mcp.WithString("clusterMemberName", mcp.Description("the given clusterMemberName to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			clusterMemberName := request.Params.Arguments["clusterMemberName"].(string)

			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource("clustermembers").Name(clusterMemberName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Cluster Member '%s' was deleted successfully.", clusterMemberName)), nil
		},
	}
}
