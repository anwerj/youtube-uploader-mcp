package core

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

type Video struct {
	ID            string   `json:"id"`
	Path          string   `json:"path"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	CategoryID    string   `json:"category_id"`
	PrivacyStatus string   `json:"privacy_status"`
	MadeForKids   bool     `json:"made_for_kids"`
}

func (v *Video) toUpload() (*youtube.Video, error) {
	privacy := "private"
	if v.PrivacyStatus != "" {
		privacy = v.PrivacyStatus
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       v.Title,
			Description: v.Description,
			Tags:        v.Tags,
			CategoryId:  v.CategoryID,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: privacy,
			MadeForKids:   v.MadeForKids,
		},
	}

	return upload, nil
}

func (c *Core) UploadVideo(ctx context.Context, video *Video, token *oauth2.Token) (string, error) {

	// First open the video file and verify it exists
	file, err := os.Open(video.Path)
	if err != nil {
		return "", fmt.Errorf("failed to open video file %s: %w", video.Path, err)
	}
	defer file.Close()

	service, err := c.Service(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube service: %w", err)
	}

	upload, err := video.toUpload()
	if err != nil {
		return "", fmt.Errorf("failed to convert video to upload format: %w", err)
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	resp, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload video: %w", err)
	}

	return resp.Id, nil
}
