package main

import (
	"context"
	"flag"
	"os"

	"github.com/anwerj/youtube-uploader-mcp/logn"
	"github.com/anwerj/youtube-uploader-mcp/server"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

var clientSecretFile = flag.String("client_secret_file", "",
	"Path to the client secret file for OAuth2 authentication.")

func main() {
	// expect a flag to be passed for the client secret file
	// if not provided, it will use the default "./client_secrets.json"
	flag.Parse()
	if *clientSecretFile == "" {
		logn.Errorf("client_secret_file is required, please provide it using the -client_secret_file flag")
		os.Exit(1)
	}

	s, err := server.Build(context.Background(), *clientSecretFile)
	if err != nil {
		logn.Errorf("Failed to build server: %v\n", err)
		os.Exit(1)
	}

	// Start the stdio server
	if err := mcpgo.ServeStdio(s); err != nil {
		logn.Errorf("Server error: %v\n", err)
		os.Exit(1)
	}
}
