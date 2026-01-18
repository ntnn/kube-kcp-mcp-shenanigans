package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PingInput struct {
	Message string `json:"message" jsonschema:"the message to ping"`
}

type PingOutput struct {
	Response string `json:"response" jsonschema:"the pong response"`
}

func Ping(ctx context.Context, req *mcp.CallToolRequest, input PingInput) (*mcp.CallToolResult, PingOutput, error) {
	return nil, PingOutput{Response: "Pong: " + input.Message}, nil
}

var PingTool = &mcp.Tool{
	Name:        "ping",
	Description: "A simple ping tool that responds with pong.",
}
