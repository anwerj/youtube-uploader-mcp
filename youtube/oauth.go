package youtube

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

func AuthURL(redirectURI string) (string, error) {
	// Implement the logic to start the YouTube client with the provided secret file and redirect URI
	// This is a placeholder function and should be replaced with actual implementation
	if redirectURI == "" {
		return "", fmt.Errorf("redirect URI must be provided")
	}

	config, err := Config()
	if err != nil {
		return "", fmt.Errorf("failed to get OAuth2 config: %w", err)
	}
	config.RedirectURL = redirectURI

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return authURL, nil
}

func GetAccessToken(code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, fmt.Errorf("code must be provided")
	}

	config, err := Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 config: %w", err)
	}
	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

func RefreshAccessToken(token *oauth2.Token) (*oauth2.Token, error) {
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}
	config, err := Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 config: %w", err)
	}
	newToken, err := config.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	return newToken, nil
}
