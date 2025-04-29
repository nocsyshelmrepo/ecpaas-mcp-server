package userrole

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pkg/errors"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
)

func ListRoles(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_roles", mcp.WithDescription(`
Retrieve the paginated roles list. The response will contain:
1. items: An array containing globalRole data where:
  - roleName: Maps to metadata.name
  - aggregationRoleTemplates: includes permission templates defined by KubeSphere, where each roleTemplate encapsulates multiple rules.
  - rules: defines actual access permissions to Kubernetes resources based on Kubernetes-native definitions.
2. totalItems: The total number of roles in KubeSphere.
`),
			mcp.WithNumber("limit", mcp.Description("Number of roles displayed at once. Default is "+constants.DefLimit)),
			mcp.WithNumber("page", mcp.Description("Page number of roles to display. Default is "+constants.DefPage)),
			mcp.WithString("level", mcp.Description("role level. it's four level: global (platform-level), cluster (cluster-level), workspace (workspace-level), namespace (project-level)"), mcp.Required()),
			mcp.WithString("cluster", mcp.Description("the given clusterName which role belong to. require when level is cluster, workspace, namespace.")),
			mcp.WithString("workspace", mcp.Description("role in which workspace. require when level is workspace.")),
			mcp.WithString("project", mcp.Description("role in which project. require when level is namespace.")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			level := ""
			if reqLevel, ok := request.Params.Arguments["level"].(string); ok &&
				(reqLevel == constants.PlatformLevel || reqLevel == constants.ClusterLevel || reqLevel == constants.WorkspaceLevel || reqLevel == constants.ProjectLevel) {
				level = reqLevel
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
			switch level {
			case constants.PlatformLevel:
				client, err := ksconfig.RestClient(iamv1beta1.SchemeGroupVersion, "")
				if err != nil {
					return nil, err
				}
				data, err := client.Get().Resource(iamv1beta1.ResourcesPluralGlobalRole).
					Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			case constants.ClusterLevel:
				cluster, ok := request.Params.Arguments["cluster"].(string)
				if !ok || cluster == "" {
					return nil, errors.Errorf("cluster is not allow empty when level is cluster")
				}
				client, err := ksconfig.RestClient(iamv1beta1.SchemeGroupVersion, cluster)
				if err != nil {
					return nil, err
				}
				data, err := client.Get().Resource(iamv1beta1.ResourcesPluralClusterRole).
					Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			case constants.WorkspaceLevel:
				cluster, ok := request.Params.Arguments["cluster"].(string)
				if !ok || cluster == "" {
					return nil, errors.Errorf("cluster is not allow empty when level is workspace")
				}
				workspace, ok := request.Params.Arguments["workspace"].(string)
				if !ok || workspace == "" {
					return nil, errors.Errorf("workspace is not allow empty when level is workspace")
				}
				client, err := ksconfig.RestClient(iamv1beta1.SchemeGroupVersion, cluster)
				if err != nil {
					return nil, err
				}
				data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).SubResource(iamv1beta1.ResourcesPluralWorkspaceRole).
					Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			case constants.ProjectLevel:
				cluster, ok := request.Params.Arguments["cluster"].(string)
				if !ok || cluster == "" {
					return nil, errors.Errorf("cluster is not allow empty when level is namespace")
				}
				project, ok := request.Params.Arguments["project"].(string)
				if !ok || project == "" {
					return nil, errors.Errorf("project is not allow empty when level is namespace")
				}
				client, err := ksconfig.RestClient(iamv1beta1.SchemeGroupVersion, cluster)
				if err != nil {
					return nil, err
				}
				data, err := client.Get().Namespace(project).Resource(iamv1beta1.ResourcesPluralRole).
					Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				return nil, errors.Errorf("unsupport level. it's should be one of %s", strings.Join([]string{constants.PlatformLevel, constants.ClusterLevel, constants.WorkspaceLevel, constants.ProjectLevel}, ","))
			}
		},
	}
}

func GetRole(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("get_role", mcp.WithDescription(`
Get role information by roleName. The response will contain:
- roleName: Maps to metadata.name
- aggregationRoleTemplates: includes permission templates defined by KubeSphere, where each roleTemplate encapsulates multiple rules.
- rules: defines actual access permissions to Kubernetes resources based on Kubernetes-native definitions.		
`),
			mcp.WithString("level", mcp.Description("role level. it's three level: global (platform-level), workspace (workspace-level), namespace (project-level)"), mcp.Required()),
			mcp.WithString("workspace", mcp.Description("role in which workspace. require when level is workspace.")),
			mcp.WithString("project", mcp.Description("role in which project. require when level is namespace.")),
			mcp.WithString("rolename", mcp.Description("the given rolename"), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			level := ""
			if reqLevel, ok := request.Params.Arguments["level"].(string); ok &&
				(reqLevel == constants.PlatformLevel || reqLevel == constants.WorkspaceLevel || reqLevel == constants.ProjectLevel) {
				level = reqLevel
			}
			rolename := request.Params.Arguments["rolename"].(string)
			// deal http request
			client, err := ksconfig.RestClient(iamv1beta1.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			switch level {
			case constants.PlatformLevel:
				data, err := client.Get().Resource(iamv1beta1.ResourcesPluralGlobalRole).Name(rolename).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			case constants.WorkspaceLevel:
				workspace, ok := request.Params.Arguments["workspace"].(string)
				if !ok || workspace == "" {
					return nil, errors.Errorf("workspace is not allow empty when level is workspace")
				}
				data, err := client.Get().Resource(tenantv1beta1.ResourcePluralWorkspace).Name(workspace).SubResource(iamv1beta1.ResourcesPluralWorkspaceRole, rolename).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			case constants.ProjectLevel:
				project, ok := request.Params.Arguments["project"].(string)
				if !ok || project == "" {
					return nil, errors.Errorf("project is not allow empty when level is namespace")
				}
				data, err := client.Get().Namespace(project).Resource(iamv1beta1.ResourcesPluralRole).Name(rolename).Do(ctx).Raw()
				if err != nil {
					return nil, err
				}

				return mcp.NewToolResultText(string(data)), nil
			default:
				return nil, errors.Errorf("unsupport level. it's should be one of %s", strings.Join([]string{constants.PlatformLevel, constants.WorkspaceLevel, constants.ProjectLevel}, ","))
			}
		},
	}
}

func ListPermissions(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("list_permissions", mcp.WithDescription(`
Retrieve all permissions defined by KubeSphere. The response will contain:
items: An array of globalRole data where:
  - permissionName: Maps to metadata.name.
  - displayName: Maps to spec.displayName (a user-friendly, intuitive name for the permission).
  - description: Maps to spec.description (a detailed explanation of the permission).
  - rules: define Kubernetes RBAC-based access permissions to resources.
  - specific metadata.annotations fields indicate:
  - iam.kubesphere.io/scope: the permission level. typically includes three scopes: global (platform-level access), workspace (workspace-level access), namespace (project-level access)
`),
			mcp.WithString("level", mcp.Description("permission level. Default is all")),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// deal request params
			label := []string{"kubesphere.io/managed=true"}
			if reqLevel, ok := request.Params.Arguments["level"].(string); ok &&
				(reqLevel == "global" || reqLevel == "workspace" || reqLevel == "namespace") {
				label = append(label, fmt.Sprintf("iam.kubesphere.io/scope=%s", reqLevel))
			}
			// deal http request
			client, err := ksconfig.RestClient(iamv1beta1.SchemeGroupVersion, "")
			if err != nil {
				return nil, err
			}
			data, err := client.Get().Resource(iamv1beta1.ResourcesPluralRoleTemplate).
				Param("labelSelector", strings.Join(label, ",")).
				Do(ctx).Raw()
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
