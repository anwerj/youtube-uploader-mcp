package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/anwerj/youtube-uploader-mcp/hook"
	"github.com/anwerj/youtube-uploader-mcp/logn"
	"github.com/anwerj/youtube-uploader-mcp/tool"
	"github.com/anwerj/youtube-uploader-mcp/tool/accesstoken"
	"github.com/anwerj/youtube-uploader-mcp/tool/authenticate"
	"github.com/anwerj/youtube-uploader-mcp/tool/refreshtoken"
	"github.com/anwerj/youtube-uploader-mcp/tool/uploadvideo"
	"github.com/anwerj/youtube-uploader-mcp/youtube"
	"github.com/mark3labs/mcp-go/server"
)

var clientSecretFile = flag.String("client_secret_file", "",
	"Path to the client secret file for OAuth2 authentication.")

func main() {
	// expect a flag to be passed for the client secret file
	// if not provided, it will use the default "./client_secrets.json"
	flag.Parse()
	if *clientSecretFile == "" {
		fmt.Println("client_secret_file is required, please provide it using the -client_secret_file flag")
		return
	}
	err := youtube.Init(*clientSecretFile)
	if err != nil {
		fmt.Printf("Failed to initialize YouTube client: %v\n", err)
		return
	}

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
		logn.Infof("Registering tool: %s\n", t.Name())
		// Define the tool and add it to the server
		s.AddTool(t.Define(ctx), t.Handle)
	}

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
