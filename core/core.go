package core

import (
	"os"

	"github.com/anwerj/youtube-uploader-mcp/logn"
	"golang.org/x/oauth2"
)

type Core struct {
	config *oauth2.Config
}

func NewCore(clientSecretFile string) *Core {
	config, err := Init(clientSecretFile)
	if err != nil {
		logn.Errorf("Failed to initialize YouTube client: %v\n", err)
		os.Exit(1)
	}
	return &Core{
		config: config,
	}
}

//------------------ AUTH RELATED ------------------ //

// Auth methods are implemented in oauth.go
