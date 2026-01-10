package server

import (
	"context"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/anwerj/youtube-uploader-mcp/hook"
	"github.com/anwerj/youtube-uploader-mcp/logn"
	"github.com/anwerj/youtube-uploader-mcp/server/tool"
	"github.com/mark3labs/mcp-go/server"
)

func Build(ctx context.Context, clientSecretFile string) (*server.MCPServer, error) {
	c := core.NewCore(clientSecretFile)

	s := server.NewMCPServer(
		"Youtube Uploader MCP",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithHooks(hook.New().Define()),
		server.WithLogging(),
	)

	tools := []Tool{
		&tool.AuthenticateTool{Core: c},
		&tool.AccessTokenTool{Core: c},
		&tool.GetChannelsTool{Core: c},
		&tool.RefreshTokenTool{Core: c},
		&tool.UploadVideoTool{Core: c},
	}
	for _, t := range tools {
		logn.Debugf("Registering tool: %s\n", t.Name())
		// Define the tool and add it to the server
		s.AddTool(t.Define(ctx), t.Handle)
	}

	return s, nil
}
