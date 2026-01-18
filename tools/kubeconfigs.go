package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeconfigInput struct {
}

type KubeconfigOutput struct {
	Kubeconfig string `json:"kubeconfig"`
}

func Kubeconfig(ctx context.Context, req *mcp.CallToolRequest, input KubeconfigInput) (*mcp.CallToolResult, KubeconfigOutput, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	apiConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, KubeconfigOutput{}, err
	}
	d, err := json.Marshal(apiConfig)
	if err != nil {
		return nil, KubeconfigOutput{}, err
	}
	return nil, KubeconfigOutput{Kubeconfig: string(d)}, nil
}

var KubeconfigTool = &mcp.Tool{
	Name:        "kubeconfig",
	Description: "Retrieves the kubeconfig for the current context.",
}
