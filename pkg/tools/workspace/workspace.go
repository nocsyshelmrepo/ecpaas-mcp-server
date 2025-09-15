package workspace

import (
	"bytes"
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/runtime/schema"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
)

func ListWorkspaces(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_workspaces", mcp.WithDescription(`
Retrieve the paginated workspace list. The response will include:
1. items: An array of workspace objects containing:
   - workspace: Maps to metadata.name.
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
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "tenant.kubesphere.io", Version: "v1alpha3"}, "")
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
Get workspace information by the name of workspace. The response will contain:
- workspace: Maps to metadata.name.
- administrator: Maps to spec.template.spec.manager. indicates the workspace's administrator.
- clusters: Maps to spec.placement.clusterSelector. specifies the clusters to which the workspace is assigned.		
`),
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "tenant.kubesphere.io", Version: "v1alpha3"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspaceTemplate).Name(request.Params.Arguments["workspace"].(string)).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteWorkspace(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_workspace", mcp.WithDescription(`Delete a specified workspace by name.`),
			mcp.WithString("workspace", mcp.Description("the given workspaceName to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspace"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "tenant.kubesphere.io", Version: "v1alpha3"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource(tenantv1beta1.ResourcePluralWorkspaceTemplate).Name(workspace).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("Workspace '%s' was deleted successfully.", workspace)), nil
		},
	}
}

func ListWorkspaceMembers(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_workspace_members", mcp.WithDescription(`
Retrieve the paginated workspace members list by the name of workspace. The response will include:
1. items: An array of workspace members objects containing:
  - username: Maps to metadata.name.
  - specific metadata.annotations fields indicate:
    - iam.kubesphere.io/workspacerole: The workspace's role assigned to this user.
2. totalItems: The total number of workspace members in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of workspace members displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of workspace members to display. Default is "+constants.DefPage)),
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspace"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
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

func GetWorkspaceMember(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_workspace_member", mcp.WithDescription(`
Get a specific workspace member by the name of workspace and workspace member. The response will include:
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
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
			mcp.WithString("workspaceMemberName", mcp.Description("the given workspaceMemberName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspace"].(string)
			workspaceMemberName := request.Params.Arguments["workspaceMemberName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).Suffix("workspacemembers", workspaceMemberName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteWorkspaceMember(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_workspace_member", mcp.WithDescription(`Delete a specified workspaceMember by the name of workspace and workspace member.`),
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
			mcp.WithString("workspaceMemberName", mcp.Description("the workspaceMember name to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspace"].(string)
			workspaceMemberName := request.Params.Arguments["workspaceMemberName"].(string)

			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).Suffix("workspacemembers", workspaceMemberName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("WorkspaceMember '%s' in worsapce '%s' was deleted successfully.", workspaceMemberName, workspace)), nil
		},
	}
}

func GetWorkspaceQuotas(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_workspace_quota", mcp.WithDescription(`
Get workspace's quotas by the name of workspace. The response will contain:
- name: Maps to metadata.name, the same as workspaceName.
- quota: the details quota information.
- specific metadata.labels fields indicate:
 - kubesphere.io/workspace: which workspace belong to.
if workspace_quota is not set. will response not found.
`),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			workspace := request.Params.Arguments["workspace"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Namespace(project).Resource("resourcequotas").Param("workspace", workspace).Do(ctx).Raw()
			if err != nil && !bytes.Contains(data, []byte("not found")) {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func ListProjectMembers(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_project_members", mcp.WithDescription(`
Retrieve the paginated project members list. The response will include:
1. items: An array of project member objects in project containing:
  - username: Maps to metadata.name
  - specific metadata.annotations fields indicate:
   - iam.kubesphere.io/role: the project role which this user belong to.
2. totalItems: The total number of project members in KubeSphere.		
`),
			mcp.WithNumber("limit", mcp.Description("Number of project members displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of project members to display. Default is "+constants.DefPage)),
			mcp.WithString("cluster", mcp.Description("the given clusterName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			cluster := request.Params.Arguments["cluster"].(string)
			project := request.Params.Arguments["project"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, cluster)
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Namespace(project).Resource("members").
				Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
