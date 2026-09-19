package core

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

const videoCatalogTTL = 60 * time.Second

type videoCatalogEntry struct {
	uploadsPlaylistID string
	videos            []VideoSummary
	fetchedAt         time.Time
	truncated         bool
}

type Core struct {
	config     *oauth2.Config
	workingDir string

	catalogMu sync.RWMutex
	catalog   map[string]videoCatalogEntry
}

func NewCore(clientSecretFile string) *Core {
	core := &Core{}
	return core
}

func (c *Core) WithSecretFile(secretfile string) error {
	var err error
	if secretfile == "" {
		return fmt.Errorf("secret file must be provided")
	}

	secrets, err := os.ReadFile(secretfile)
	if err != nil {
		return fmt.Errorf("failed to read secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(
		secrets,
		youtube.YoutubeUploadScope,
		youtube.YoutubeForceSslScope,
		youtube.YoutubeReadonlyScope,
	)
	if err != nil {
		return fmt.Errorf("failed to create OAuth2 config: %w", err)
	}
	c.config = config

	return nil
}

func (c *Core) WithWorkingDir(dir string) error {
	c.workingDir = dir
	return nil
}
