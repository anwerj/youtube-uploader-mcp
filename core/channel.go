package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const fChannelsFileName = ".youtube_uploader_channels_cache"

type Channel struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	CustomURL string        `json:"customer_url"`
	Token     *oauth2.Token `json:"token"` // OAuth2 token for accessing the YouTube API
}

func (c *Channel) Mask() {
	c.Token.AccessToken = "***"
	c.Token.RefreshToken = "***"
}

type Channels map[string]*Channel

func (c *Core) GetChannelForToken(token *oauth2.Token) (*Channel, error) {
	ctx := context.Background()

	service, err := c.Service(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	call := service.Channels.List([]string{"snippet"}).Mine(true)
	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no channels found for token")
	}

	channel := &Channel{
		ID:        resp.Items[0].Id,
		Name:      resp.Items[0].Snippet.Title,
		CustomURL: resp.Items[0].Snippet.CustomUrl,
		Token:     token,
	}
	return channel, nil
}

func (c *Core) GetChannelByID(id string) (*Channel, error) {
	if id == "" {
		return nil, fmt.Errorf("channel ID must be provided")
	}

	channels, err := c.ReadChannels(true)
	if err != nil {
		return nil, fmt.Errorf("failed to read channels: %s", err.Error())
	}

	channel, exists := channels[id]
	if !exists {
		return nil, fmt.Errorf("channel with ID %s not found", id)
	}

	return channel, nil
}

func (c *Core) SaveChannel(channel *Channel) error {
	if channel == nil || channel.Token == nil {
		return fmt.Errorf("invalid channel: channel or token is nil")
	}

	// First retrieve existing channels
	channels, err := c.ReadChannels(true)
	if err != nil {
		return fmt.Errorf("failed to read existing channels: %s", err.Error())
	}
	channels[channel.ID] = channel

	plaintext, err := json.Marshal(channels)
	if err != nil {
		return fmt.Errorf("failed to save channels: %s", err.Error())
	}

	key, err := loadOrCreateEncryptionKey(c.workingDir)
	if err != nil {
		return fmt.Errorf("failed to load encryption key: %w", err)
	}

	ciphertext, err := encryptBytes(key, plaintext)
	if err != nil {
		return fmt.Errorf("failed to encrypt channels: %w", err)
	}

	fChannelsPath := filepath.Join(c.workingDir, fChannelsFileName)
	if err := os.WriteFile(fChannelsPath, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to save channels: %s", err.Error())
	}

	return nil
}

func (c *Core) ReadChannels(ignoreError bool) (Channels, error) {
	fChannelsPath := filepath.Join(c.workingDir, fChannelsFileName)
	ciphertext, err := os.ReadFile(fChannelsPath)
	if err != nil {
		if os.IsNotExist(err) && ignoreError {
			return make(Channels), nil // Return an empty Channels map if the file does not exist
		}
		return nil, fmt.Errorf("failed to open channels file: %s", err.Error())
	}

	key, err := loadOrCreateEncryptionKey(c.workingDir)
	if err != nil {
		if ignoreError {
			return make(Channels), nil
		}
		return nil, fmt.Errorf("failed to load encryption key: %w", err)
	}

	plaintext, err := decryptBytes(key, ciphertext)
	if err != nil {
		if ignoreError {
			return make(Channels), nil // Legacy plaintext cache or corrupt/foreign file: start fresh.
		}
		return nil, fmt.Errorf("failed to decrypt channels file: %w", err)
	}

	var channels Channels
	if err := json.Unmarshal(plaintext, &channels); err != nil {
		if ignoreError {
			return make(Channels), nil // Return an empty Channels map if the file does not exist
		}
		return nil, fmt.Errorf("failed to decode channels from file: %s", err.Error())
	}

	return channels, nil
}
