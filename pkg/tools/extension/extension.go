package extension

// func ListExtensions(ksconfig *kubesphere.KSConfig) server.ServerTool {
// 	return server.ServerTool{
// 		Tool: mcp.NewTool("list_extensions", mcp.WithDescription(`
// Retrieve the paginated extensions list. The response will include:
// 1. items: An array of extensions objects containing:
//   - extensionName: Maps to metadata.name
//   - extensionDisplayName: Maps to spec.displayName
//   - specific metadata.labels fields indicate:
//    - kubesphere.io/category: the extension's category.
//   - installed state: Maps to status.state.
//   - installed information: Maps to status.clusterSchedulingStatuses. contains clusters which has installed.
// 2. totalItems: The total number of extensions in KubeSphere.
// `),
// 			mcp.WithNumber("limit", mcp.Description("Number of clusters displayed at once. Default is "+constants.DefLimit)),
// 			mcp.WithNumber("page", mcp.Description("Page number of clusters to display. Default is "+constants.DefPage)),
// 		),
// 		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
// 			// deal request params
// 			limit := constants.DefLimit
// 			if reqLimit, ok := request.Params.Arguments["limit"].(int64); ok && reqLimit != 0 {
// 				limit = fmt.Sprintf("%d", reqLimit)
// 			}
// 			page := constants.DefPage
// 			if reqPage, ok := request.Params.Arguments["page"].(int64); ok && reqPage != 0 {
// 				page = fmt.Sprintf("%d", reqPage)
// 			}
// 			// deal http request
// 			client, err := ksconfig.RestClient(corev1alpha1.SchemeGroupVersion, "")
// 			if err != nil {
// 				return nil, err
// 			}
// 			data, err := client.Get().Resource("extensions").Param("sortBy", "createTime").Param("limit", limit).Param("page", page).Do(ctx).Raw()
// 			if err != nil {
// 				return nil, err
// 			}

// 			return mcp.NewToolResultText(string(data)), nil
// 		},
// 	}
// }
