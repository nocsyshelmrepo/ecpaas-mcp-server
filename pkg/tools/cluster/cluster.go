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
