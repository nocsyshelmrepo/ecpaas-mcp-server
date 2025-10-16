package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"kubesphere.io/ks-mcp-server/pkg/constants"
	"kubesphere.io/ks-mcp-server/pkg/kubesphere"
)

func ApplyYAMLFile(ksconfig *kubesphere.KSConfig) server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("apply_yaml_file", mcp.WithDescription(`
Applies a Kubernetes resource defined in YAML format using server-side apply.
This is similar to 'kubectl apply --server-side --force-conflicts'.
The YAML content must contain a single, valid Kubernetes resource.
`),
			mcp.WithString("yaml_content", mcp.Description("The full YAML content of the Kubernetes resource to apply."), mcp.Required()),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// 1. Validate and get YAML content argument
			yamlContent, ok := request.Params.Arguments["yaml_content"].(string)
			if !ok || yamlContent == "" {
				return nil, fmt.Errorf("required argument 'yaml_content' is missing or not a string")
			}

			// 2. Decode YAML into Unstructured object
			// This allows us to handle any Kubernetes resource type.
			obj := &unstructured.Unstructured{}
			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(yamlContent), 4096)
			if err := decoder.Decode(obj); err != nil {
				return nil, fmt.Errorf("failed to decode YAML content: %w", err)
			}

			// 3. Extract GVK for RestClient and Resource path
			apiVersion := obj.GetAPIVersion()
			kind := obj.GetKind()
			name := obj.GetName()

			if apiVersion == "" || kind == "" || name == "" {
				return nil, fmt.Errorf("missing apiVersion, kind, or name in the provided YAML")
			}

			gv, err := schema.ParseGroupVersion(apiVersion)
			if err != nil {
				return nil, fmt.Errorf("invalid apiVersion '%s': %w", apiVersion, err)
			}

			// Convert kind to plural resource name (a simplification, but works for most core resources)
			// For a robust solution, you'd need a DiscoveryClient to get the plural.
			resourcePlural := strings.ToLower(kind) + "s" // e.g., "Deployment" -> "deployments"

			// 4. Get the KubeSphere RestClient
			// The cluster parameter is usually empty ("") for accessing resources in the host cluster.
			client, err := ksconfig.RestClient(gv, "")
			if err != nil {
				return nil, fmt.Errorf("failed to get RestClient for GroupVersion '%s': %w", gv.String(), err)
			}

			// 5. Convert Unstructured object back to JSON for the request body
			jsonBody, err := json.Marshal(obj)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal object to JSON: %w", err)
			}

			// 6. Perform Server-Side Apply request (PATCH method is often used for SSA)
			// The Content-Type header and query parameters are crucial for Server-Side Apply.
			req := client.Patch(types.ApplyPatchType).
				Resource(resourcePlural).
				Name(name).
				Body(jsonBody).
				SetHeader("Content-Type", constants.ApplyContentType). // Set the required content type for SSA
				Param("fieldManager", "ks-mcp-server").                // Mandatory SSA parameter
				Param("force", "true")                                 // Allows overwriting conflicts

			// Handle namespaced resources
			namespace := obj.GetNamespace()
			if namespace != "" {
				req = req.Namespace(namespace)
			}

			data, err := req.Do(ctx).Raw()
			if err != nil {
				return nil, fmt.Errorf("failed to execute Server-Side Apply: %w", err)
			}

			// 7. Return the result
			return mcp.NewToolResultText(string(data)), nil
		},
	}
}
