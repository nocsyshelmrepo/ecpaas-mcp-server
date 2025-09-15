package workspace

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"

	//openpitrixv1 "kubesphere.io/kubesphere/v3/pkg/kapis/openpitrix/v1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
)

func ListApplicationRepos(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_application_repos", mcp.WithDescription(`
Retrieve the paginated application repository list by workspaceName, it's set by workspace. The response will include:
1. items: An array of application objects in workspace containing:
  - specific metadata.labels fields indicate:
   - kubesphere.io/workspace: which workspace belong to.
2. totalItems: The total number of application repositories in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of application repositories displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of application repositories to display. Default is "+constants.DefPage)),
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
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "openpitrix.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).SubResource("repos").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func ListApplications(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_applications", mcp.WithDescription(`
Retrieve the paginated application list by project, it install by project. The response will include:
1. items: An array of application objects in project containing:
  - applicationName: Maps to metadata.name
  - specific metadata.labels fields indicate:
   - kubesphere.io/cluster: which cluster belong to.
   - kubesphere.io/workspace: which workspace belong to.
   - kubesphere.io/namespace: which project belong to.
  - specific metadata.annotations fields indicate:
   - application.kubesphere.io/app-displayName: application which installed in the application repository
   - application.kubesphere.io/app-name: format [app-repository]-[app-originalName]
   - application.kubesphere.io/app-originalName: this application original from.
   - application.kubesphere.io/app-version: this application version which installed.
2. totalItems: The total number of applications in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of applications displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of applications to display. Default is "+constants.DefPage)),
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspace"].(string)
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
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "openpitrix.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).Suffix("namespaces", project, "applications").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetApplication(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_application", mcp.WithDescription(`
Get the application information by list_applications' cluster_id, project and workspace. The response will include:
- applicationClusterId: Maps to metadata.name
- specific metadata.labels fields indicate:
 - kubesphere.io/cluster: which cluster belong to.
 - kubesphere.io/workspace: which workspace belong to.
 - kubesphere.io/namespace: which project belong to.
- specific metadata.annotations fields indicate:
 - application.kubesphere.io/app-displayName: application which installed in the application repository
 - application.kubesphere.io/app-name: format [app-repository]-[app-originalName]
 - application.kubesphere.io/app-originalName: this application original from.
 - application.kubesphere.io/app-version: this application version which installed.
`),
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given project"), mcp.Required()),
			mcp.WithString("applicationClusterId", mcp.Description("the given applicationClusterId"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspace"].(string)
			project := request.Params.Arguments["project"].(string)
			applicationClusterId := request.Params.Arguments["applicationClusterId"].(string)
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "openpitrix.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).Suffix("namespaces", project, "applications", applicationClusterId).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteApplication(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_application", mcp.WithDescription(`Delete a specified application by list_applications' cluster_id, project and workspace.`),
			mcp.WithString("workspace", mcp.Description("the given workspaceName"), mcp.Required()),
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
			mcp.WithString("applicationClusterId", mcp.Description("the given applicationClusterId to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			workspace := request.Params.Arguments["workspace"].(string)
			project := request.Params.Arguments["project"].(string)
			applicationClusterId := request.Params.Arguments["applicationClusterId"].(string)

			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "openpitrix.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).Suffix("namespaces", project, "applications", applicationClusterId).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("ApplicationClusterId '%s' in project '%s' workspace '%s' was deleted successfully.", applicationClusterId, project, workspace)), nil
		},
	}
}

func GetApplicationVersion(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_application_version", mcp.WithDescription(`
Retrieve the paginated application versions list by list_applications' cluster_id and project. The response will include:
1. items: An array of application versions objects containing:
- applicationName: Maps to metadata.name
- specific metadata.labels fields indicate:
 - kubesphere.io/app-id: the application id.
 - kubesphere.io/repo-name: the application repository which application belong to.
 - kubesphere.io/workspace: which workspace belong to.
2. totalItems: The total number of application versions in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of application versions displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of application versions to display. Default is "+constants.DefPage)),
			mcp.WithString("applicationTemplateAppId", mcp.Description("the given applicationTemplateAppId"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			applicationTemplateAppId := request.Params.Arguments["applicationTemplateAppId"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(schema.GroupVersion{Group: "openpitrix.io", Version: "v1"}, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("apps").Name(applicationTemplateAppId).SubResource("versions").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
