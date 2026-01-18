package main

import (
	"context"
	"flag"
	"os"
	"os/user"

	"github.com/anwerj/youtube-uploader-mcp/logn"
	"github.com/anwerj/youtube-uploader-mcp/yum"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

var clientSecretFile = flag.String("client_secret_file", "",
	"Path to the client secret file for OAuth2 authentication.")
var workingDir = flag.String("working_dir", "",
	"Working directory to store the OAuth2 tokens and logs.")

func main() {
	// expect a flag to be passed for the client secret file
	// if not provided, it will use the default "./client_secrets.json"
	flag.Parse()
	if *clientSecretFile == "" {
		logn.Errorf("client_secret_file is required, please provide it using the -client_secret_file flag")
		os.Exit(1)
	}
	if *workingDir == "" {
		current, err := user.Current()
		if err != nil {
			logn.Errorf("Failed to get current user: %v\n", err)
			os.Exit(1)
		}
		*workingDir = current.HomeDir
	}

	s, err := yum.Build(context.Background(), *clientSecretFile, *workingDir)
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
