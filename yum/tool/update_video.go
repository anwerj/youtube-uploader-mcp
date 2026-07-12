package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type UpdateVideoTool struct {
	Core *core.Core
}

func (t *UpdateVideoTool) Name() string {
	return "update_video"
}

func (t *UpdateVideoTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Configure or update an existing YouTube video's metadata (add to playlist, upload subtitles, upload custom thumbnail)"),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Channel ID associated with the video"),
		),
		mcp.WithString("video_id",
			mcp.Required(),
			mcp.Description("The ID of the video to update"),
		),
		mcp.WithString("playlist_id",
			mcp.Description("Optional playlist ID to which the video should be added"),
		),
		mcp.WithString("subtitle_path",
			mcp.Description("Optional path to a subtitle file (e.g., .srt or .vtt) to attach to the video"),
		),
		mcp.WithString("subtitle_language",
			mcp.Description("Language code of the subtitle track (ISO 639-1). Default is en (English)."),
		),
		mcp.WithString("thumbnail_path",
			mcp.Description("Optional path to an image file to use as the video's custom thumbnail (max 2MB)"),
		),
	)
}

type UpdateVideoResult struct {
	VideoID          string            `json:"video_id"`
	PlaylistStatus   string            `json:"playlist_status,omitempty"`
	SubtitlesStatus  string            `json:"subtitles_status,omitempty"`
	ThumbnailStatus  string            `json:"thumbnail_status,omitempty"`
	Errors           map[string]string `json:"errors,omitempty"`
}

func (t *UpdateVideoTool) Handle(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	channelId := request.GetString("channel_id", "")
	videoId := request.GetString("video_id", "")
	if channelId == "" || videoId == "" {
		return mcp.NewToolResultError("channel_id and video_id are required"), nil
	}

	playlistID := request.GetString("playlist_id", "")
	subtitlePath := request.GetString("subtitle_path", "")
	subtitleLanguage := request.GetString("subtitle_language", "en")
	thumbnailPath := request.GetString("thumbnail_path", "")

	if playlistID == "" && subtitlePath == "" && thumbnailPath == "" {
		return mcp.NewToolResultError("at least one update parameter (playlist_id, subtitle_path, or thumbnail_path) must be provided"), nil
	}

	// 1. Fail-fast validation of local files before authentication and API calls
	if subtitlePath != "" {
		info, err := os.Stat(subtitlePath)
		if err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultError(fmt.Sprintf("subtitle file not found: %s", subtitlePath)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("failed to read subtitle file: %s", err.Error())), nil
		}
		if info.IsDir() {
			return mcp.NewToolResultError(fmt.Sprintf("subtitle path is a directory: %s", subtitlePath)), nil
		}
	}

	if thumbnailPath != "" {
		info, err := os.Stat(thumbnailPath)
		if err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultError(fmt.Sprintf("thumbnail file not found: %s", thumbnailPath)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("failed to read thumbnail file: %s", err.Error())), nil
		}
		if info.IsDir() {
			return mcp.NewToolResultError(fmt.Sprintf("thumbnail path is a directory: %s", thumbnailPath)), nil
		}
		// Custom thumbnail file size limit is 2MB on YouTube
		if info.Size() > 2*1024*1024 {
			return mcp.NewToolResultError(fmt.Sprintf("thumbnail file size (%.2f MB) exceeds the YouTube limit of 2MB", float64(info.Size())/(1024*1024))), nil
		}
	}

	// 2. Token authorization
	channel, err := t.Core.GetChannelByID(channelId)
	if err != nil {
		return mcp.NewToolResultError("Failed to load token: " + err.Error()), nil
	}

	if channel == nil || channel.Token == nil {
		return mcp.NewToolResultError("channel or token is nil, please authenticate first"), nil
	}

	if channel.Token.Expiry.IsZero() || channel.Token.AccessToken == "" || channel.Token.RefreshToken == "" {
		return mcp.NewToolResultError("channel token is expired or malformed, please authenticate again"), nil
	}

	// Check if token is expiring (within 2 minutes) and refresh it if necessary
	now := time.Now().In(channel.Token.Expiry.Location())
	if channel.Token.Expiry.Before(now.Add(2 * time.Minute)) {
		newToken, err := t.Core.RefreshAccessToken(channel.Token)
		if err != nil {
			return mcp.NewToolResultError("token was expiring, failed to refresh: " + err.Error()), nil
		}
		channel.Token = newToken
		err = t.Core.SaveChannel(channel)
		if err != nil {
			return mcp.NewToolResultError("token was expiring, failed to save refreshed token: " + err.Error()), nil
		}
	}

	// 3. Process actions
	result := UpdateVideoResult{
		VideoID: videoId,
		Errors:  make(map[string]string),
	}

	if playlistID != "" {
		if err := t.Core.AddVideoToPlaylist(ctx, playlistID, videoId, channel.Token); err != nil {
			result.PlaylistStatus = "failed"
			result.Errors["playlist"] = err.Error()
		} else {
			result.PlaylistStatus = "success"
		}
	}

	if subtitlePath != "" {
		if err := t.Core.AddSubtitles(ctx, videoId, subtitlePath, subtitleLanguage, channel.Token); err != nil {
			result.SubtitlesStatus = "failed"
			result.Errors["subtitles"] = err.Error()
		} else {
			result.SubtitlesStatus = "success"
		}
	}

	if thumbnailPath != "" {
		if err := t.Core.SetThumbnail(ctx, videoId, thumbnailPath, channel.Token); err != nil {
			result.ThumbnailStatus = "failed"
			result.Errors["thumbnail"] = err.Error()
		} else {
			result.ThumbnailStatus = "success"
		}
	}

	if len(result.Errors) == 0 {
		result.Errors = nil
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal results: " + err.Error()), nil
	}

	// If everything failed, mark the tool result as error, but return the details.
	// We check how many of the requested items failed.
	requestedCount := 0
	failedCount := 0
	if playlistID != "" {
		requestedCount++
		if result.PlaylistStatus == "failed" {
			failedCount++
		}
	}
	if subtitlePath != "" {
		requestedCount++
		if result.SubtitlesStatus == "failed" {
			failedCount++
		}
	}
	if thumbnailPath != "" {
		requestedCount++
		if result.ThumbnailStatus == "failed" {
			failedCount++
		}
	}

	if failedCount == requestedCount && requestedCount > 0 {
		return mcp.NewToolResultError(string(bytes)), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}
