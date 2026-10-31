package core

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Service creates a new YouTube service using the provided OAuth2 token.
func (c *Core) Service(ctx context.Context, token *oauth2.Token) (*youtube.Service, error) {
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}

	if c.config == nil {
		return nil, fmt.Errorf("OAuth2 config is not initialized in Core")
	}

	client := c.config.Client(ctx, token)
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return service, nil
}

// DefaultRedirectURI is the redirect URI loaded from the client secret.
func (c *Core) DefaultRedirectURI() string {
	if c == nil || c.config == nil {
		return ""
	}
	return c.config.RedirectURL
}
