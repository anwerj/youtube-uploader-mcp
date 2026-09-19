package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/googleapi"
)

const listVideosMaxCap = 200

type ListVideosTool struct {
	Core *core.Core
}

func (t *ListVideosTool) Name() string {
	return "list_videos"
}

func (t *ListVideosTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("List videos uploaded to a YouTube channel (public, unlisted, and private). "+
			"Optional query filters title and description locally. Use offset and max to page results."),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Channel ID to list videos for. Call channels to list authenticated channels if needed."),
		),
		mcp.WithString("query",
			mcp.Description("Optional case-insensitive substring match on video title and description"),
		),
		mcp.WithString("privacy_status",
			mcp.Description("Optional filter: public, unlisted, or private"),
		),
		mcp.WithString("order_by",
			mcp.Description("Sort field. Only date is supported. Default is date"),
		),
		mcp.WithString("direction",
			mcp.Description("Sort direction: desc (newest first, default) or asc (oldest first)"),
		),
		mcp.WithNumber("max",
			mcp.Description("Maximum videos to return. Default 25, hard cap 200"),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of matched videos to skip for paging. Default 0"),
		),
	)
}

func (t *ListVideosTool) Handle(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	channelID := request.GetString("channel_id", "")
	if channelID == "" {
		return mcp.NewToolResultError("channel_id is required to list videos"), nil
	}

	privacyStatus := request.GetString("privacy_status", "")
	if privacyStatus != "" &&
		privacyStatus != "public" && privacyStatus != "private" && privacyStatus != "unlisted" {
		return mcp.NewToolResultError("privacy_status must be one of: public, private, unlisted"), nil
	}

	orderBy := request.GetString("order_by", "date")
	if orderBy == "" {
		orderBy = "date"
	}
	if orderBy != "date" {
		return mcp.NewToolResultError("order_by must be date"), nil
	}

	direction := request.GetString("direction", "desc")
	if direction == "" {
		direction = "desc"
	}
	if direction != "desc" && direction != "asc" {
		return mcp.NewToolResultError("direction must be one of: desc, asc"), nil
	}

	max := request.GetInt("max", 25)
	if max <= 0 {
		max = 25
	}
	if max > listVideosMaxCap {
		max = listVideosMaxCap
	}

	offset := request.GetInt("offset", 0)
	if offset < 0 {
		offset = 0
	}

	channel, err := t.Core.GetChannelByID(channelID)
	if err != nil {
		return mcp.NewToolResultError(listVideosChannelError(err)), nil
	}

	if channel == nil || channel.Token == nil {
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

	result, err := t.Core.ListVideos(ctx, channel, core.ListVideosOptions{
		Query:         request.GetString("query", ""),
		PrivacyStatus: privacyStatus,
		OrderBy:       orderBy,
		Direction:     direction,
		Max:           max,
		Offset:        offset,
	})
	if err != nil {
		return mcp.NewToolResultError(formatListVideosAPIError(err)), nil
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError("failed to marshal list videos result: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}

func listVideosChannelError(err error) string {
	msg := err.Error()
	if msg == "" {
		return "Failed to load channel: unknown error"
	}
	return fmt.Sprintf(
		"Failed to load channel: %s. Call the channels tool to pick an authenticated channel, or authenticate if none exist.",
		msg,
	)
}

func formatListVideosAPIError(err error) string {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		for _, item := range apiErr.Errors {
			switch item.Reason {
			case "playlistItemsNotAccessible":
				return "Cannot access this channel's uploads playlist. Re-authenticate as the channel owner using authenticate and accesstoken."
			case "playlistNotFound":
				return "Uploads playlist not found for this channel. Verify channel_id with the channels tool."
			case "insufficientPermissions", "forbidden":
				return "Saved OAuth token lacks required permissions. Delete .youtube_uploader_channels_cache in your working directory and run authenticate again to grant youtube.readonly and related scopes."
			}
		}
		switch apiErr.Code {
		case 401:
			return "YouTube API rejected the access token. Delete .youtube_uploader_channels_cache and run authenticate again."
		case 403:
			return "Access denied listing uploads. Delete .youtube_uploader_channels_cache and run authenticate again, or verify you own this channel."
		case 404:
			return "Uploads playlist not found for this channel. Verify channel_id with the channels tool."
		}
	}

	return "Failed to list videos: " + err.Error()
}
