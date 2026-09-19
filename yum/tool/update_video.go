package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/youtube/v3"
)

type UpdateVideoTool struct {
	Core *core.Core
}

func (t *UpdateVideoTool) Name() string {
	return "update_video"
}

func (t *UpdateVideoTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Configure or update an existing YouTube video's metadata (add to playlist, upload subtitles, upload custom thumbnail, schedule publish time)"),
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
		mcp.WithString("publish_at",
			mcp.Description("Optional RFC3339 timestamp (e.g. 2026-09-20T18:00:00+05:30) to schedule "+
				"an already-uploaded video for release. The video stays private until that time."),
		),
		mcp.WithBoolean("made_for_kids",
			mcp.Description("Optional. Whether to mark the video as made for kids (updates status.selfDeclaredMadeForKids). Omit to leave the current setting unchanged."),
		),
	)
}

type UpdateVideoResult struct {
	VideoID           string            `json:"video_id"`
	PlaylistStatus    string            `json:"playlist_status,omitempty"`
	SubtitlesStatus   string            `json:"subtitles_status,omitempty"`
	ThumbnailStatus   string            `json:"thumbnail_status,omitempty"`
	ScheduleStatus    string            `json:"schedule_status,omitempty"`
	MadeForKidsStatus string            `json:"made_for_kids_status,omitempty"`
	Errors            map[string]string `json:"errors,omitempty"`
}

func boolArg(request mcp.CallToolRequest, key string) (*bool, error) {
	val, ok := request.GetArguments()[key]
	if !ok {
		return nil, nil
	}
	switch v := val.(type) {
	case bool:
		return &v, nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("%s must be a boolean", key)
		}
		return &b, nil
	default:
		return nil, fmt.Errorf("%s must be a boolean", key)
	}
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
	publishAt := request.GetString("publish_at", "")
	madeForKids, err := boolArg(request, "made_for_kids")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if playlistID == "" && subtitlePath == "" && thumbnailPath == "" && publishAt == "" && madeForKids == nil {
		return mcp.NewToolResultError(
			"at least one update parameter (playlist_id, subtitle_path, thumbnail_path, publish_at, or made_for_kids) must be provided"), nil
	}

	if publishAt != "" {
		when, err := time.Parse(time.RFC3339, publishAt)
		if err != nil {
			return mcp.NewToolResultError("publish_at must be a valid RFC3339 timestamp: " + err.Error()), nil
		}
		if !when.After(time.Now()) {
			return mcp.NewToolResultError("publish_at must be a future timestamp"), nil
		}
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

	// 3. Pre-flight: if scheduling and/or made_for_kids is requested, fetch the
	// video and build the combined status once, before any mutating action
	// runs, so an invalid request does not leave playlist/subtitle/thumbnail
	// changes applied.
	var newStatus *youtube.VideoStatus
	if publishAt != "" || madeForKids != nil {
		video, err := t.Core.GetVideo(ctx, videoId, channel.Token)
		if err != nil {
			return mcp.NewToolResultError("failed to verify video before updating status: " + err.Error()), nil
		}
		if video.Status == nil {
			return mcp.NewToolResultError(fmt.Sprintf("video %s returned no status", videoId)), nil
		}
		opts := core.StatusUpdateOptions{MadeForKids: madeForKids}
		if publishAt != "" {
			opts.PublishAt = &publishAt
		}
		newStatus, err = core.BuildStatusUpdate(video.Status, opts)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	// 4. Process actions
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

	if publishAt != "" || madeForKids != nil {
		if err := t.Core.UpdateVideoStatus(ctx, videoId, newStatus, channel.Token); err != nil {
			if publishAt != "" {
				result.ScheduleStatus = "failed"
				result.Errors["schedule"] = err.Error()
			}
			if madeForKids != nil {
				result.MadeForKidsStatus = "failed"
				result.Errors["made_for_kids"] = err.Error()
			}
		} else {
			if publishAt != "" {
				result.ScheduleStatus = "success"
			}
			if madeForKids != nil {
				result.MadeForKidsStatus = "success"
			}
			t.Core.InvalidateVideoCatalog(channelId)
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
	if publishAt != "" {
		requestedCount++
		if result.ScheduleStatus == "failed" {
			failedCount++
		}
	}
	if madeForKids != nil {
		requestedCount++
		if result.MadeForKidsStatus == "failed" {
			failedCount++
		}
	}

	if failedCount == requestedCount && requestedCount > 0 {
		return mcp.NewToolResultError(string(bytes)), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}
