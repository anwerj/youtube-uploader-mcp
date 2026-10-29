package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
)

// fRedirectURIFileName stores the redirect URI used by the latest AuthURL call so
// that a later accesstoken call - possibly served by another process - can send
// the identical value during the code exchange.
const fRedirectURIFileName = ".youtube_uploader_redirect_uri"

func (c *Core) persistRedirectURI(uri string) {
	if c.workingDir == "" || uri == "" {
		return
	}
	// Best effort only: the in-memory value still covers the single-process case.
	_ = os.WriteFile(filepath.Join(c.workingDir, fRedirectURIFileName), []byte(uri), 0o600)
}

// resolveRedirectURI returns the redirect URI that must accompany the code
// exchange. Google rejects the exchange unless it matches the one used to build
// the authorization URL, so prefer, in order: the in-memory value from the last
// AuthURL call, the value persisted by a previous process, the redirect URI
// registered in the client secret, and finally DefaultRedirectURI.
func (c *Core) resolveRedirectURI() string {
	c.redirectURIMu.Lock()
	uri := c.redirectURI
	c.redirectURIMu.Unlock()
	if uri != "" {
		return uri
	}
	if c.workingDir != "" {
		if b, err := os.ReadFile(filepath.Join(c.workingDir, fRedirectURIFileName)); err == nil {
			if v := strings.TrimSpace(string(b)); v != "" {
				return v
			}
		}
	}
	if c.config != nil && c.config.RedirectURL != "" {
		return c.config.RedirectURL
	}
	return DefaultRedirectURI
}

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

	// Remember the redirect URI so the code exchange in GetAccessToken can send
	// the same value. Google requires the two to match exactly. It is also
	// persisted to disk because authenticate and accesstoken may be served by
	// different server processes.
	c.redirectURIMu.Lock()
	c.redirectURI = redirectURI
	c.redirectURIMu.Unlock()
	c.persistRedirectURI(redirectURI)

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

	// Exchange the code with the same redirect_uri used in AuthURL; omitting it
	// makes Google reject the exchange with invalid_grant.
	redirectURI := c.resolveRedirectURI()

	config := *c.config
	config.RedirectURL = redirectURI

	token, err := config.Exchange(context.TODO(), code)
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
