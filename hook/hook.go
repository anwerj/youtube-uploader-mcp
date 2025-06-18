package hook

import (
	"context"

	"github.com/anwerj/youtube-uploader-mcp/logn"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type hook struct {
}

func New() *hook {
	return &hook{}
}

func (h *hook) Define() *server.Hooks {
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		logn.Infof("beforeAny: %s, %v, %v\n", method, id, message)
	})

	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		logn.Infof("afterCallTool: %v, %v, %v\n", id, message, result)
	})

	return hooks
}
