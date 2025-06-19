package youtube

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	// DefaultRedirectURI is the default redirect URI for OAuth2 authentication
	DefaultRedirectURI = "https://localhost:8080"
)

var config *oauth2.Config

func Init(secretfile string) error {
	var err error
	if secretfile == "" {
		return fmt.Errorf("secret file must be provided")
	}

	secrets, err := os.ReadFile(secretfile)
	if err != nil {
		return fmt.Errorf("failed to read secret file: %w", err)
	}

	config, err = google.ConfigFromJSON(secrets, youtube.YoutubeUploadScope, youtube.YoutubeReadonlyScope)

	if err != nil {
		return fmt.Errorf("failed to create OAuth2 config: %w", err)
	}
	return nil
}

func Config() (*oauth2.Config, error) {
	if config == nil {
		return nil, fmt.Errorf("YouTube OAuth2 config is not initialized, call Init() first")
	}
	return config, nil
}

// Service creates a new YouTube service using the provided OAuth2 token.
func Service(ctx context.Context, token *oauth2.Token) (*youtube.Service, error) {
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}

	config, err := Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 config: %w", err)
	}

	client := config.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return service, nil
}
