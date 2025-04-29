package workspace

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	applicationv2 "kubesphere.io/api/application/v2"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
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
			client, err := ksconfig.RestClient(applicationv2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).SubResource("repos").
				Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
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
			mcp.WithString("project", mcp.Description("the given projectName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
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
			client, err := ksconfig.RestClient(applicationv2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Namespace(project).Resource("applications").
				Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
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
Retrieve the paginated application list by applicationName and project. The response will include:
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
`),
			mcp.WithString("project", mcp.Description("the given project"), mcp.Required()),
			mcp.WithString("applicationName", mcp.Description("the given applicationName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			project := request.Params.Arguments["project"].(string)
			applicationName := request.Params.Arguments["applicationName"].(string)
			// deal http request
			client, err := ksconfig.RestClient(applicationv2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Namespace(project).Resource("applications").Name(applicationName).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetApplicationVersion(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_application_version", mcp.WithDescription(`
Retrieve the paginated application list by applicationName and project. The response will include:
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
			mcp.WithString("applicationName", mcp.Description("the given applicationName"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			applicationName := request.Params.Arguments["applicationName"].(string)
			limit := constants.DefLimit
			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
				limit = fmt.Sprintf("%d", reqLimit)
			}
			page := constants.DefPage
			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
				page = fmt.Sprintf("%d", reqPage)
			}
			// deal http request
			client, err := ksconfig.RestClient(applicationv2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource("apps").Name(applicationName).SubResource("versions").
				Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
