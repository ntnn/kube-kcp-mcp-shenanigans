package tools

import (
	"context"
	"fmt"

	kcpclientset "github.com/kcp-dev/kcp/sdk/client/clientset/versioned"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// ListWorkspacesInput defines the input parameters for listing workspaces
type ListWorkspacesInput struct {
	Kubeconfig string `json:"kubeconfig,omitempty"`
}

// WorkspaceInfo contains information about a single workspace
type WorkspaceInfo struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	Phase       string `json:"phase"`
	URL         string `json:"url,omitempty"`
}

// ListWorkspacesOutput defines the output structure for listing workspaces
type ListWorkspacesOutput struct {
	Workspaces []WorkspaceInfo `json:"workspaces"`
	Count      int             `json:"count"`
}

// ListWorkspaces retrieves all workspaces from a kcp instance
func ListWorkspaces(ctx context.Context, req *mcp.CallToolRequest, input ListWorkspacesInput) (*mcp.CallToolResult, ListWorkspacesOutput, error) {
	restConfig,

		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if input.KubeconfigPath != "" {
		loadingRules.ExplicitPath = input.KubeconfigPath
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if input.Context != "" {
		configOverrides.CurrentContext = input.Context
	}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, ListWorkspacesOutput{}, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	kcpClient, err := kcpclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, ListWorkspacesOutput{}, fmt.Errorf("failed to create kcp client: %w", err)
	}

	workspaceList, err := kcpClient.TenancyV1alpha1().Workspaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, ListWorkspacesOutput{}, fmt.Errorf("failed to list workspaces: %w", err)
	}

	workspaces := make(WorkspaceInfo, 0, len(workspaceList.Items))
	for _, ws := range workspaceList.Items {
		info := WorkspaceInfo{
			Name:        ws.Name,
			ClusterName: ws.Spec.Cluster,
			Phase:       string(ws.Status.Phase),
			URL:         ws.Spec.URL,
		}

		workspaces = append(workspaces, info)
	}

	output := ListWorkspacesOutput{
		Workspaces: workspaces,
		Count:      len(workspaces),
	}

	return nil, output, nil
}

var ListWorkspacesTool = &mcp.Tool{
	Name:        "list-workspaces",
	Description: "Lists all workspaces in a kcp instance. Accepts optional kubeconfig path and context parameters.",
}
