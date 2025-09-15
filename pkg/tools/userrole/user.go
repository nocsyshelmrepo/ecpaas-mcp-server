package userrole

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
)

func ListUsers(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_users", mcp.WithDescription(`
Retrieve the paginated user list. The response will contain:
1. items: An array containing user data where:
  - username: Maps to metadata.name
  - specific metadata.annotations fields indicate:
    - iam.kubesphere.io/globalrole: The user's assigned cluster role
    - iam.kubesphere.io/granted-clusters: Clusters assigned to the user
2. totalItems: The total number of users in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of users displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of users to display. Default is "+constants.DefPage)),
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
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(iamv1alpha2.ResourcesPluralUser).
				Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func GetUser(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_user", mcp.WithDescription(`
Get user information by username. The response will contain:
- username: Maps to metadata.name
- specific metadata.annotations fields indicate:
  - iam.kubesphere.io/globalrole: The user's assigned platform role
  - iam.kubesphere.io/granted-clusters: Clusters assigned to the user
`),
			mcp.WithString("userName", mcp.Description("the given username"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(iamv1alpha2.ResourcesPluralUser).SubResource(request.Params.Arguments["userName"].(string)).Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}

func DeleteUser(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_user", mcp.WithDescription(`Delete a specified user by name.`),
			mcp.WithString("userName", mcp.Description("the given user name to delete"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			userName := request.Params.Arguments["userName"].(string)

			// deal http request
			client, err := ksconfig.RestClient(iamv1alpha2.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			err = client.Delete().Resource(iamv1alpha2.ResourcesPluralUser).SubResource(userName).Do(ctx).Error()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(fmt.Sprintf("User '%s' was deleted successfully.", userName)), nil
		},
	}
}
