package refreshtoken

import (
	"context"
	"encoding/json"

	"github.com/anwerj/youtube-uploader-mcp/tool"
	"github.com/anwerj/youtube-uploader-mcp/youtube"
	"github.com/mark3labs/mcp-go/mcp"
)

type RefreshTokenTool struct {
	tool.Tool
}

func (t *RefreshTokenTool) Name() string {
	return "refreshtoken"
}

func (t *RefreshTokenTool) Define(ctx context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Refreshes the OAuth2 access token for YouTube API. If the UploadVideo tool is used "+
			"and it requires to refresh token, Use this tool. "+
			"This tool is useful for maintaining a valid access token without user intervention."+
			"It will return channel details with refreshed token, keep it in memory to upload the videos"),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Channel ID to upload the video to, if not provided, Agent should call tool channels to get the list of channels and ask the user to select one"),
		),
	)
}

func (t *RefreshTokenTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID := request.GetString("channel_id", "")
	if channelID == "" {
		return mcp.NewToolResultError("channel_id is required, use tool getchannel"), nil
	}

	channel, err := youtube.GetChannelByID(channelID)
	if err != nil {
		return mcp.NewToolResultError("Failed to get channel: " + err.Error()), nil
	}

	// first load the token from the file
	token := channel.Token
	if token == nil {
		return mcp.NewToolResultError("Failed to load token: " + channelID), nil
	}
	// Check if the token is already valid
	if token.ExpiresIn == 0 {
		return mcp.NewToolResultText("Invalid access token loaded:. "), nil
	}

	newToken, err := youtube.RefreshAccessToken(token)

	if err != nil {
		return mcp.NewToolResultError("Failed to refresh access token: " + err.Error()), nil
	}
	channel.Token = newToken
	// Save the refreshed access token
	err = youtube.SaveChannel(channel)
	if err != nil {
		return mcp.NewToolResultError("Failed to save refreshed access token: " + err.Error()), nil
	}
	// First mask the secrets in channel
	channel.Mask()

	bytes, err := json.Marshal(channel)
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal channels: " + err.Error()), nil
	}

	// For now, we will just return a dummy response
	return mcp.NewToolResultText(string(bytes)), nil
}
