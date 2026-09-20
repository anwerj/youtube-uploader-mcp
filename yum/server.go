package yum

import (
	"context"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/anwerj/youtube-uploader-mcp/hook"
	"github.com/anwerj/youtube-uploader-mcp/logn"
	"github.com/anwerj/youtube-uploader-mcp/tracker"
	"github.com/anwerj/youtube-uploader-mcp/yum/tool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Build(ctx context.Context, clientSecretFile string, workingDir string) (*server.MCPServer, error) {
	c := core.NewCore(clientSecretFile)
	if err := c.WithSecretFile(clientSecretFile); err != nil {
		return nil, err
	}
	if err := c.WithWorkingDir(workingDir); err != nil {
		return nil, err
	}

	var mcpServer *server.MCPServer
	hooks := hook.New().Define()
	// Nudge MCP clients that support tools.listChanged to re-fetch tools/list after connect.
	hooks.AddAfterInitialize(func(ctx context.Context, _ any, _ *mcp.InitializeRequest, _ *mcp.InitializeResult) {
		if mcpServer == nil {
			return
		}
		if err := mcpServer.SendNotificationToClient(ctx, mcp.MethodNotificationToolsListChanged, nil); err != nil {
			logn.Debugf("tools/list_changed notification: %v\n", err)
		}
	})

	mcpServer = server.NewMCPServer(
		"Youtube Uploader MCP",
		Version,
		server.WithToolCapabilities(true),
		server.WithHooks(hooks),
		server.WithLogging(),
	)
	s := mcpServer

	tr := tracker.New()

	tools := []Tool{
		&tool.AuthenticateTool{Core: c},
		&tool.AccessTokenTool{Core: c},
		&tool.GetChannelsTool{Core: c},
		&tool.RefreshTokenTool{Core: c},
		&tool.UploadVideoTool{Core: c, Tracker: tr},
		&tool.UpdateVideoTool{Core: c},
		&tool.ListVideosTool{Core: c},
		&tool.VerifyUploadTool{Tracker: tr},
	}
	for _, t := range tools {
		logn.Debugf("Registering tool: %s\n", t.Name())
		// Define the tool and add it to the server
		s.AddTool(t.Define(ctx), t.Handle)
	}

	return s, nil
}
