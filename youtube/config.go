package youtube

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func Config(secretfile string) (*oauth2.Config, error) {
	if secretfile == "" {
		return nil, fmt.Errorf("secret file must be provided")
	}

	secrets, err := os.ReadFile(secretfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(secrets, "https://www.googleapis.com/auth/youtube.upload")
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 config: %w", err)
	}

	return config, nil
}
