package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func testChannel(id string) *Channel {
	return &Channel{
		ID:        id,
		Name:      "Test Channel",
		CustomURL: "@test",
		Token: &oauth2.Token{
			AccessToken:  "access-secret-value",
			RefreshToken: "refresh-secret-value",
			Expiry:       time.Now().Add(time.Hour),
		},
	}
}

func TestSaveChannelEncryptsTokensAtRest(t *testing.T) {
	c := &Core{workingDir: t.TempDir()}

	if err := c.SaveChannel(testChannel("chan-1")); err != nil {
		t.Fatalf("SaveChannel: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(c.workingDir, fChannelsFileName))
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	if bytes.Contains(raw, []byte("refresh-secret-value")) {
		t.Fatalf("refresh token stored in plaintext on disk: %q", raw)
	}
	if bytes.Contains(raw, []byte("access-secret-value")) {
		t.Fatalf("access token stored in plaintext on disk: %q", raw)
	}
}

func TestSaveChannelThenReadChannelsRoundTrip(t *testing.T) {
	c := &Core{workingDir: t.TempDir()}

	want := testChannel("chan-1")
	if err := c.SaveChannel(want); err != nil {
		t.Fatalf("SaveChannel: %v", err)
	}

	channels, err := c.ReadChannels(false)
	if err != nil {
		t.Fatalf("ReadChannels: %v", err)
	}

	got, ok := channels["chan-1"]
	if !ok {
		t.Fatalf("expected channel chan-1 to be present")
	}
	if got.Token.RefreshToken != want.Token.RefreshToken {
		t.Fatalf("refresh token: got %q, want %q", got.Token.RefreshToken, want.Token.RefreshToken)
	}
	if got.Token.AccessToken != want.Token.AccessToken {
		t.Fatalf("access token: got %q, want %q", got.Token.AccessToken, want.Token.AccessToken)
	}
	if got.Name != want.Name {
		t.Fatalf("name: got %q, want %q", got.Name, want.Name)
	}
}

func TestReadChannelsMissingFileIgnoreError(t *testing.T) {
	c := &Core{workingDir: t.TempDir()}

	channels, err := c.ReadChannels(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 0 {
		t.Fatalf("expected empty channels map, got %d entries", len(channels))
	}
}

func TestReadChannelsMissingFilePropagatesError(t *testing.T) {
	c := &Core{workingDir: t.TempDir()}

	if _, err := c.ReadChannels(false); err == nil {
		t.Fatalf("expected error when cache file is missing and ignoreError is false")
	}
}

func TestReadChannelsRejectsLegacyPlaintextCacheWhenErrorsNotIgnored(t *testing.T) {
	c := &Core{workingDir: t.TempDir()}

	legacyPlaintext := `{"chan-1":{"id":"chan-1","name":"Legacy","token":{"refresh_token":"legacy-secret"}}}`
	path := filepath.Join(c.workingDir, fChannelsFileName)
	if err := os.WriteFile(path, []byte(legacyPlaintext), 0600); err != nil {
		t.Fatalf("failed to seed legacy cache file: %v", err)
	}

	if _, err := c.ReadChannels(false); err == nil {
		t.Fatalf("expected error reading legacy plaintext cache as ciphertext")
	}
}

func TestSaveChannelPreservesExistingChannels(t *testing.T) {
	c := &Core{workingDir: t.TempDir()}

	if err := c.SaveChannel(testChannel("chan-1")); err != nil {
		t.Fatalf("SaveChannel chan-1: %v", err)
	}
	if err := c.SaveChannel(testChannel("chan-2")); err != nil {
		t.Fatalf("SaveChannel chan-2: %v", err)
	}

	channels, err := c.ReadChannels(false)
	if err != nil {
		t.Fatalf("ReadChannels: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
}
