package tool

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/anwerj/youtube-uploader-mcp/tracker"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	uploadEarlyReturn = 30 * time.Second
	uploadHardTimeout = 10 * time.Minute
)

type UploadVideoTool struct {
	Core    *core.Core
	Tracker *tracker.Tracker
}

func (t *UploadVideoTool) Name() string {
	return "upload_video"
}

func (t *UploadVideoTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Upload a video to YouTube. "),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the video file"),
		),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Channel ID to upload the video to, if not provided, Agent should call tool channels to get the list of channels and ask the user to select one"),
		),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Description of the video, if not provided, Agent should generate a description based on the video content"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Title of the video, if not provided, Agent should generate a title based on the video description"),
		),
		mcp.WithString("tags",
			mcp.Required(),
			mcp.Description("Tags for the video, if not provided, Agent should generate tags based on the video description"),
		),
		mcp.WithString("category_id",
			mcp.Required(),
			mcp.Description("Category ID for the video, if not provided, Agent should generate a category based on the video description"),
		),
		mcp.WithString("video_language",
			mcp.Description("Optional language of the video's audio/content (ISO 639-1). If not set, YouTube's default or auto-detection is used."),
		),
		mcp.WithString("status",
			mcp.Description("status of video, could be any of unlisted, public, private. Default is private"),
		),
		mcp.WithString("publish_at",
			mcp.Description("The date and time when the video is scheduled to publish. It can be set only if the privacy status of the video is private. The value is specified in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sZ)."),
		),
		mcp.WithBoolean("made_for_kids",
			mcp.Description("Whether the video is made exclusively for kids. Default is false"),
		),
	)
}

func (t *UploadVideoTool) Handle(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	filePath := request.GetString("file_path", "")
	description := request.GetString("description", "")
	title := request.GetString("title", "")
	tags := request.GetString("tags", "")
	categoryID := request.GetString("category_id", "")
	if filePath == "" || description == "" || title == "" || tags == "" || categoryID == "" {
		return mcp.NewToolResultError(
			"all fields are required: file_path, description, title, tags, category_id"), nil
	}
	channelId := request.GetString("channel_id", "")
	if channelId == "" {
		return mcp.NewToolResultError("channel_id is required to upload video"), nil
	}
	status := request.GetString("status", "private")
	if status != "public" && status != "private" && status != "unlisted" {
		return mcp.NewToolResultError("status must be one of: public, private, unlisted"), nil
	}
	madeForKids := request.GetBool("made_for_kids", false)
	videoLanguage := request.GetString("video_language", "")

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError("video file path does not exist: " + filePath), nil
		}
		return mcp.NewToolResultError("failed to check video file path: " + err.Error()), nil
	}
	if info.IsDir() {
		return mcp.NewToolResultError("video file path is a directory: " + filePath), nil
	}

	channel, err := t.Core.GetChannelByID(channelId)
	if err != nil {
		return mcp.NewToolResultError("Failed to load token: " + err.Error()), nil
	}

	if channel == nil ||
		channel.Token == nil {
		return mcp.NewToolResultError("channel or token is nil, please authenticate first"), nil
	}

	if channel.Token.Expiry.IsZero() ||
		channel.Token.AccessToken == "" ||
		channel.Token.RefreshToken == "" {
		return mcp.NewToolResultError(
			"channel token is expired or malformed, please start authenticate"), nil
	}

	now := time.Now().In(channel.Token.Expiry.Location())
	if channel.Token.Expiry.Before(now.Add(2 * time.Minute)) {
		newToken, err := t.Core.RefreshAccessToken(channel.Token)
		if err != nil {
			return mcp.NewToolResultError(
				"token was expiring, Failed to refresh token: " + err.Error()), nil
		}
		channel.Token = newToken
		err = t.Core.SaveChannel(channel)
		if err != nil {
			return mcp.NewToolResultError(
				"token was expiring, Failed to save refreshed token: " + err.Error()), nil
		}
	}

	video := core.Video{
		Path:          filePath,
		Title:         title,
		Description:   description,
		Tags:          strings.Split(tags, ","),
		CategoryID:    categoryID,
		Language:      videoLanguage,
		PrivacyStatus: status,
		MadeForKids:   madeForKids,
		PublishAt:     request.GetString("publish_at", ""),
	}

	jobKey := tracker.UploadJobKey(channelId, filePath)
	token := channel.Token

	return tracker.RunTool(t.Tracker, t.Name(), jobKey, uploadEarlyReturn, uploadHardTimeout,
		func(ctx context.Context) (core.Video, error) {
			id, err := t.Core.UploadVideo(ctx, &video, token)
			if err != nil {
				return core.Video{}, err
			}
			t.Core.InvalidateVideoCatalog(channelId)
			video.ID = id
			return video, nil
		},
		func(v core.Video) ([]byte, error) { return json.Marshal(v) },
		tracker.MCPRunOptions{
			DuplicateError: "an upload for this channel_id and file_path is already in progress; use verify_upload to check status",
			RunningMessage: "Upload is continuing in the background. Call verify_upload with the same channel_id and file_path until status is success or failed.",
			PendingFields: map[string]string{
				"channel_id": channelId,
				"file_path":  filePath,
				"job_id":     jobKey,
			},
		},
	)
}
