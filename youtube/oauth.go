package youtube

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func AuthURL(secretfile string, redirectURI string) (string, error) {
	// Implement the logic to start the YouTube client with the provided secret file and redirect URI
	// This is a placeholder function and should be replaced with actual implementation
	if secretfile == "" || redirectURI == "" {
		return "", fmt.Errorf("secret file and redirect URI must be provided")
	}

	secrets, err := os.ReadFile(secretfile)
	if err != nil {
		return "", fmt.Errorf("failed to read secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(secrets, "https://www.googleapis.com/auth/youtube.upload")
	if err != nil {
		return "", fmt.Errorf("failed to create OAuth2 config: %w", err)
	}
	config.RedirectURL = redirectURI

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return authURL, nil
}

func GetAccessToken(secretfile string, code string) (*oauth2.Token, error) {
	if secretfile == "" || code == "" {
		return nil, fmt.Errorf("secret file and code must be provided")
	}

	secrets, err := os.ReadFile(secretfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(secrets, "https://www.googleapis.com/auth/youtube.upload")
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 config: %w", err)
	}

	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

func RefreshAccessToken(secretfile string, token *oauth2.Token) (*oauth2.Token, error) {
	if secretfile == "" || token == nil {
		return nil, fmt.Errorf("secret file and token must be provided")
	}

	secrets, err := os.ReadFile(secretfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(secrets, "https://www.googleapis.com/auth/youtube.upload")
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth2 config: %w", err)
	}

	newToken, err := config.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	return newToken, nil
}
