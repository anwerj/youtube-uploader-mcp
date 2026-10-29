package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// testRegisteredRedirectURI mirrors what Google puts in a downloaded desktop
// client secret.
const testRegisteredRedirectURI = "http://localhost"

// newTestCore returns a Core whose token endpoint is served by the given handler.
// A google "installed" client secret is written to dir and loaded from there.
func newTestCore(t *testing.T, dir, tokenURL string) *Core {
	t.Helper()

	secret := map[string]any{
		"installed": map[string]any{
			"client_id":     "test-client-id",
			"client_secret": "test-client-secret",
			"auth_uri":      "https://accounts.google.com/o/oauth2/auth",
			"token_uri":     tokenURL,
			"redirect_uris": []string{testRegisteredRedirectURI},
		},
	}
	raw, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("marshal secret: %v", err)
	}
	secretFile := filepath.Join(dir, "client_secret.json")
	if err := os.WriteFile(secretFile, raw, 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	core := NewCore("")
	if err := core.WithSecretFile(secretFile); err != nil {
		t.Fatalf("WithSecretFile: %v", err)
	}
	if err := core.WithWorkingDir(dir); err != nil {
		t.Fatalf("WithWorkingDir: %v", err)
	}
	return core
}

// tokenServer asserts the exchange request carries redirect_uri and returns a token.
func tokenServer(t *testing.T, wantRedirectURI string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", got)
		}
		if got := r.PostForm.Get("code"); got != "test-code" {
			t.Errorf("code = %q, want test-code", got)
		}
		// The regression: without redirect_uri Google answers invalid_grant.
		if got := r.PostForm.Get("redirect_uri"); got != wantRedirectURI {
			t.Errorf("redirect_uri = %q, want %q", got, wantRedirectURI)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"test-access-token","token_type":"Bearer",` +
			`"refresh_token":"test-refresh-token","expires_in":3600}`))
	}))
}

func TestAuthURLAndExchangeUseSameRedirectURI(t *testing.T) {
	const redirectURI = "http://localhost:9191"

	dir := t.TempDir()
	srv := tokenServer(t, redirectURI)
	defer srv.Close()

	core := newTestCore(t, dir, srv.URL)

	authURL, err := core.AuthURL(redirectURI)
	if err != nil {
		t.Fatalf("AuthURL: %v", err)
	}
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth url: %v", err)
	}
	if got := parsed.Query().Get("redirect_uri"); got != redirectURI {
		t.Fatalf("auth url redirect_uri = %q, want %q", got, redirectURI)
	}

	token, err := core.GetAccessToken("test-code")
	if err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
	if token.AccessToken != "test-access-token" {
		t.Errorf("access token = %q", token.AccessToken)
	}
	if token.RefreshToken != "test-refresh-token" {
		t.Errorf("refresh token = %q", token.RefreshToken)
	}
	if token.Expiry.IsZero() {
		t.Error("expiry was not parsed")
	}
}

// authenticate and accesstoken can be served by different server processes, so the
// redirect URI chosen during authentication must survive a restart.
func TestExchangeUsesPersistedRedirectURI(t *testing.T) {
	const redirectURI = "http://localhost:9191"

	dir := t.TempDir()
	srv := tokenServer(t, redirectURI)
	defer srv.Close()

	authenticating := newTestCore(t, dir, srv.URL)
	if _, err := authenticating.AuthURL(redirectURI); err != nil {
		t.Fatalf("AuthURL: %v", err)
	}

	// Fresh Core, same working dir: no in-memory redirect URI available.
	exchanging := newTestCore(t, dir, srv.URL)
	if _, err := exchanging.GetAccessToken("test-code"); err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
}

// Without a prior AuthURL call the exchange falls back to the redirect URI that is
// registered in the client secret.
func TestExchangeFallsBackToRegisteredRedirectURI(t *testing.T) {
	dir := t.TempDir()
	srv := tokenServer(t, testRegisteredRedirectURI)
	defer srv.Close()

	// No AuthURL call at all.
	core := newTestCore(t, dir, srv.URL)
	if _, err := core.GetAccessToken("test-code"); err != nil {
		t.Fatalf("GetAccessToken: %v", err)
	}
}
