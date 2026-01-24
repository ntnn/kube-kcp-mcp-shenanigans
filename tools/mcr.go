package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/multicluster-runtime/pkg/multicluster"
)

func init() {
	// prevent warnings from controller-runtime about loggers not being
	// configured
	log.SetLogger(zap.New(zap.UseDevMode(true)))
}

type MCRTool struct {
	name     string
	provider multicluster.Provider

	clustersLock sync.RWMutex
	clusters     map[string]cluster.Cluster
}

func NewMCRTool(ctx context.Context, name string, provider multicluster.Provider) *MCRTool {
	tool := &MCRTool{
		name:     name,
		clusters: make(map[string]cluster.Cluster),
		provider: provider,
	}

	if runnable, ok := provider.(multicluster.ProviderRunnable); ok {
		go func() {
			if err := runnable.Start(ctx, tool); err != nil {
				// TODO handle error properly
				panic(err)
			}
		}()
	}

	return tool
}

func (tool *MCRTool) Engage(ctx context.Context, clusterName string, cl cluster.Cluster) error {
	tool.clustersLock.Lock()
	defer tool.clustersLock.Unlock()
	tool.clusters[clusterName] = cl
	return nil
}

type MCRListClustersIn struct{}
type MCRListClustersOut struct {
	ClusterNames []string `json:"clusterNames" jsonschema:"the names of the available clusters"`
}

func (tool *MCRTool) ListClustersTool() (*mcp.Tool, mcp.ToolHandlerFor[MCRListClustersIn, MCRListClustersOut]) {
	toolDef := &mcp.Tool{
		Name:        "list-clusters-" + tool.name,
		Description: "Lists the names of all available clusters in the multicluster provider.",
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input MCRListClustersIn) (*mcp.CallToolResult, MCRListClustersOut, error) {
		tool.clustersLock.RLock()
		defer tool.clustersLock.RUnlock()
		return nil, MCRListClustersOut{ClusterNames: slices.Sorted(maps.Keys(tool.clusters))}, nil
	}

	return toolDef, handler
}

type MCRGetClusterIn struct {
	ClusterName string `json:"clusterName" jsonschema:"the name of the cluster to retrieve"`
}

type MCRGetClusterOut struct {
	JSONConfig string `json:"jsonConfig" jsonschema:"the config of the requested cluster in JSON format"`
	YAMLConfig string `json:"yamlConfig" jsonschema:"the config of the requested cluster in YAML format"`
}

func (tool *MCRTool) GetClusterTool() (*mcp.Tool, mcp.ToolHandlerFor[MCRGetClusterIn, MCRGetClusterOut]) {
	toolDef := &mcp.Tool{
		Name:        "get-cluster-" + tool.name,
		Description: "Retrieves the kubeconfig for a specified cluster in the multicluster provider.",
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input MCRGetClusterIn) (*mcp.CallToolResult, MCRGetClusterOut, error) {
		tool.clustersLock.RLock()
		defer tool.clustersLock.RUnlock()

		cl, exists := tool.clusters[input.ClusterName]
		if !exists {
			return nil, MCRGetClusterOut{}, fmt.Errorf("cluster %s not found", input.ClusterName)
		}

		apiConfig := restToKubeconfig(cl.GetConfig())

		jsonMarshalled, err := json.Marshal(apiConfig)
		if err != nil {
			return nil, MCRGetClusterOut{}, fmt.Errorf("failed to json marshal REST config for cluster %s: %w", input.ClusterName, err)
		}

		yamlMarshalled, err := json.MarshalIndent(apiConfig, "", "  ")
		if err != nil {
			return nil, MCRGetClusterOut{}, fmt.Errorf("failed to yaml marshal REST config for cluster %s: %w", input.ClusterName, err)
		}

		out := MCRGetClusterOut{
			JSONConfig: string(jsonMarshalled),
			YAMLConfig: string(yamlMarshalled),
		}

		return nil, out, nil
	}

	return toolDef, handler
}

func restToKubeconfig(config *rest.Config) clientcmdapi.Config {
	return clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters: map[string]*clientcmdapi.Cluster{
			"default": {
				Server:                   config.Host,
				CertificateAuthorityData: config.CAData,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"default": {
				Cluster:   "default",
				Namespace: "default",
				AuthInfo:  "default",
			},
		},
		CurrentContext: "default",
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"default": {
				Token:                 config.BearerToken,
				Username:              config.Username,
				Password:              config.Password,
				ClientCertificateData: config.CertData,
				ClientKeyData:         config.KeyData,
			},
		},
	}
}
