package youtube

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

	config, err = google.ConfigFromJSON(secrets, "https://www.googleapis.com/auth/youtube.upload")
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
