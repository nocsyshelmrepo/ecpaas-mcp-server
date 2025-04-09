package workspace

import (
	"bytes"
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
)

func ListWorkspaces(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_workspaces", mcp.WithDescription(`
Retrieve the paginated workspace list. The response will include:
1. items: An array of workspace objects containing:
   - workspaceName: Maps to metadata.name.
   - administrator: Maps to spec.template.spec.manager. indicates the workspace's administrator.
   - clusters: Maps to spec.placement.clusterSelector. specifies the clusters to which the workspace is assigned.
2. totalItems: The total number of workspaces in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of workspaces displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of workspaces to display. Default is "+constants.DefPage)),
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
			client, err := ksconfig.RestClient(tenantv1beta1.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspaceTemplate).
				Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetWorkspace(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_workspace", mcp.WithDescription(`
Get workspace information by workspaceName. The response will contain:
- workspaceName: Maps to metadata.name.
- administrator: Maps to spec.template.spec.manager. indicates the workspace's administrator.
- clusters: Maps to spec.placement.clusterSelector. specifies the clusters to which the workspace is assigned.		
`),
			mcp.WithString("workspaceName", mcp.Description("the given workspaceName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal http request
			client, err := ksconfig.RestClient(tenantv1beta1.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspaceTemplate).Name(request.Params.Arguments["workspaceName"].(string)).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func ListWorkspaceMembers(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_workspace_members", mcp.WithDescription(`
Retrieve the paginated workspace members list by workspaceName. The response will include:
1. items: An array of workspace members objects containing:
  - username: Maps to metadata.name.
  - specific metadata.annotations fields indicate:
    - iam.kubesphere.io/workspacerole: The workspace's role assigned to this user.
2. totalItems: The total number of workspace members in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of workspace members displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of workspace members to display. Default is "+constants.DefPage)),
			mcp.WithString("workspaceName", mcp.Description("the given workspaceName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspaceName"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(iamv1beta1.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).SubResource("workspacemembers").
				Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetWorkspaceQuotas(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_workspace_quota", mcp.WithDescription(`
Get workspace's quotas by workspaceName. The response will contain:
- name: Maps to metadata.name, the same as workspaceName.
- quota: the details quota information.
- specific metadata.labels fields indicate:
 - kubesphere.io/workspace: which workspace belong to.
if workspace_quota is not set. will response not found.
`),
			mcp.WithString("workspaceName", mcp.Description("the given workspaceName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspaceName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(tenantv1beta1.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).SubResource("resourcequotas", workspace).Do(ctx).Raw()
			if err != nil && !bytes.Contains(data, []byte("not found")) {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
