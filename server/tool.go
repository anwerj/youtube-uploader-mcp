package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

type Tool interface {
	Name() string
	Define(ctx context.Context) mcp.Tool
	Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
