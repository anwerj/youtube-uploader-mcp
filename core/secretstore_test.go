package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateEncryptionKeyGeneratesAndPersists(t *testing.T) {
	dir := t.TempDir()

	key, err := loadOrCreateEncryptionKey(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key) != encryptionKeySize {
		t.Fatalf("key length: got %d, want %d", len(key), encryptionKeySize)
	}

	keyPath := filepath.Join(dir, fChannelsKeyFileName)
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("expected key file to exist: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("key file permissions: got %o, want 0600", perm)
	}

	again, err := loadOrCreateEncryptionKey(dir)
	if err != nil {
		t.Fatalf("unexpected error on reload: %v", err)
	}
	if !bytes.Equal(key, again) {
		t.Fatalf("expected the same key to be reused across calls")
	}
}

func TestLoadOrCreateEncryptionKeyRejectsCorruptFile(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, fChannelsKeyFileName)
	if err := os.WriteFile(keyPath, []byte("too-short"), 0600); err != nil {
		t.Fatalf("failed to seed corrupt key file: %v", err)
	}

	if _, err := loadOrCreateEncryptionKey(dir); err == nil {
		t.Fatalf("expected error for corrupt key file, got nil")
	}
}

func TestEncryptDecryptBytesRoundTrip(t *testing.T) {
	key, err := loadOrCreateEncryptionKey(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plaintext := []byte(`{"refresh_token":"super-secret-value"}`)
	ciphertext, err := encryptBytes(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if bytes.Contains(ciphertext, []byte("super-secret-value")) {
		t.Fatalf("ciphertext leaks plaintext: %q", ciphertext)
	}

	got, err := decryptBytes(key, ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("round trip mismatch: got %q, want %q", got, plaintext)
	}
}

func TestDecryptBytesFailsWithWrongKey(t *testing.T) {
	key, err := loadOrCreateEncryptionKey(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	otherKey, err := loadOrCreateEncryptionKey(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ciphertext, err := encryptBytes(key, []byte("some secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := decryptBytes(otherKey, ciphertext); err == nil {
		t.Fatalf("expected decryption with the wrong key to fail")
	}
}

func TestDecryptBytesRejectsTruncatedCiphertext(t *testing.T) {
	key, err := loadOrCreateEncryptionKey(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := decryptBytes(key, []byte("short")); err == nil {
		t.Fatalf("expected error for truncated ciphertext, got nil")
	}
}
