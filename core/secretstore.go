package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// fChannelsKeyFileName holds the AES-256 key used to encrypt the channels
// cache file at rest. It is a separate file so that copying the cache file
// alone (e.g. an accidental backup, sync, or git add) does not also leak a
// usable refresh token.
const fChannelsKeyFileName = ".youtube_uploader_channels_cache.key"

const encryptionKeySize = 32 // AES-256

// loadOrCreateEncryptionKey returns the AES-256 key for this working
// directory, creating and persisting one with 0600 permissions if it does
// not already exist.
func loadOrCreateEncryptionKey(workingDir string) ([]byte, error) {
	keyPath := filepath.Join(workingDir, fChannelsKeyFileName)

	key, err := os.ReadFile(keyPath)
	if err == nil {
		if len(key) != encryptionKeySize {
			return nil, fmt.Errorf("encryption key file %s is corrupt: expected %d bytes, got %d", keyPath, encryptionKeySize, len(key))
		}
		return key, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read encryption key file: %w", err)
	}

	key = make([]byte, encryptionKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to persist encryption key file: %w", err)
	}

	return key, nil
}

// encryptBytes encrypts plaintext with AES-256-GCM, returning nonce||ciphertext.
func encryptBytes(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decryptBytes decrypts data produced by encryptBytes.
func decryptBytes(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}
