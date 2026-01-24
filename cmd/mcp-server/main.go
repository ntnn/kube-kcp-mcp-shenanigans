package main

import (
	"context"
	"flag"
	"kube-kcp-mcp-shenanigans/tools"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	kindprovider "sigs.k8s.io/multicluster-runtime/providers/kind"
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalf("error: %v", err)
	}
}

var (
	fHost = flag.String("host", "127.0.0.1", "host to listen on")
	fPort = flag.String("port", "8080", "port to listen on")
)

func run(ctx context.Context) error {
	server := mcp.NewServer(&mcp.Implementation{Name: "mcp-shenanigans", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, tools.PingTool, tools.Ping)
	// mcp.AddTool(server, tools.KubeconfigTool, tools.Kubeconfig)

	mcrKindProvider := kindprovider.New(kindprovider.Options{})
	mcrKindTool := tools.NewMCRTool(ctx, "kind", mcrKindProvider)

	listClustersTool, listClustersHandler := mcrKindTool.ListClustersTool()
	mcp.AddTool(server, listClustersTool, listClustersHandler)

	getClusterTool, getClusterHandler := mcrKindTool.GetClusterTool()
	mcp.AddTool(server, getClusterTool, getClusterHandler)

	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	s := &http.Server{
		Addr:    *fHost + ":" + *fPort,
		Handler: handler,
	}
	go func() {
		<-ctx.Done()
		ctxShutDown, cancel := context.WithTimeout(context.TODO(), time.Minute)
		defer cancel()
		if err := s.Shutdown(ctxShutDown); err != nil {
			log.Printf("MCP server shutdown error: %v", err)
		}
	}()

	log.Printf("MCP server listening on %s", s.Addr)
	return s.ListenAndServe()
}
