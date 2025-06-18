package main

import (
	"context"
	"fmt"

	"github.com/anwerj/youtube-uploader-mcp/hook"
	"github.com/anwerj/youtube-uploader-mcp/logn"
	"github.com/anwerj/youtube-uploader-mcp/tool"
	"github.com/anwerj/youtube-uploader-mcp/tool/accesstoken"
	"github.com/anwerj/youtube-uploader-mcp/tool/authenticate"
	"github.com/anwerj/youtube-uploader-mcp/tool/refreshtoken"
	"github.com/anwerj/youtube-uploader-mcp/tool/uploadvideo"
	"github.com/mark3labs/mcp-go/server"
)

func main() {

	s := server.NewMCPServer(
		"Youtube Uploader MCP",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithHooks(hook.New().Define()),
		server.WithLogging(),
	)

	ctx := context.Background()

	tools := []tool.Tool{
		&authenticate.AuthenticateTool{},
		&accesstoken.AccessTokenTool{},
		&refreshtoken.RefreshTokenTool{},
		&uploadvideo.UploadVideoTool{},
	}
	for _, t := range tools {
		logn.Info("Registering tool: %s\n", t.Name())
		// Define the tool and add it to the server
		s.AddTool(t.Define(ctx), t.Handle)
	}

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
