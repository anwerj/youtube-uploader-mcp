package core

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

// AuthURL generates the authorization URL for YouTube OAuth2 authentication.
func (c *Core) AuthURL(redirectURI string) (string, error) {
	if redirectURI == "" {
		return "", fmt.Errorf("redirect URI must be provided")
	}

	if c.config == nil {
		return "", fmt.Errorf("OAuth2 config is not initialized in Core")
	}
	// Copy config to avoid race conditions if we were modifying it, though here we just set RedirectURL
	// But c.config is a pointer. Modifying it is bad if concurrent.
	// Better to create a shallow copy? oauth2.Config is a struct.
	config := *c.config
	config.RedirectURL = redirectURI

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return authURL, nil
}

// GetAccessToken exchanges the authorization code for an access token.
func (c *Core) GetAccessToken(code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, fmt.Errorf("code must be provided")
	}

	if c.config == nil {
		return nil, fmt.Errorf("OAuth2 config is not initialized in Core")
	}

	token, err := c.config.Exchange(context.TODO(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// RefreshAccessToken refreshes the access token using the provided token.
func (c *Core) RefreshAccessToken(token *oauth2.Token) (*oauth2.Token, error) {
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}
	if c.config == nil {
		return nil, fmt.Errorf("OAuth2 config is not initialized in Core")
	}

	// TokenSource will automatically refresh the token if needed
	newToken, err := c.config.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	return newToken, nil
}
