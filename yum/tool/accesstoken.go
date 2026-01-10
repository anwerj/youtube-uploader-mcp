package tool

import (
	"context"
	"encoding/json"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type AccessTokenTool struct {
	Core *core.Core
}

func (t *AccessTokenTool) Name() string {
	return "accesstoken"
}

func (t *AccessTokenTool) Define(ctx context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("AccessToken tool retrieves the access token for YouTube API authentication."+
			"It requires code to fetch the access token."+
			"User must first authenticate using the authenticate tool to get the code."+
			"It will return channel details ex: ID. Use this ID in further calls to refresh token or upload video"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("The code part received on redirected URL, Only send the code part, not the full URL."),
		),
	)
}

func (t *AccessTokenTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code := request.GetString("code", "")
	if code == "" {
		return mcp.NewToolResultError("code is required"), nil
	}

	token, err := t.Core.GetAccessToken(code)
	if err != nil {
		return mcp.NewToolResultError("Failed to get access token: " + err.Error()), nil
	}
	// Once we have received the access token, lets fetch the channel for the token
	channel, err := t.Core.GetChannelForToken(token)
	if err != nil {
		return mcp.NewToolResultError("Failed to get channel: " + err.Error()), nil
	}

	// save the access token or use it as needed
	err = t.Core.SaveChannel(channel)

	if err != nil {
		return mcp.NewToolResultError("Failed to save access token: " + err.Error()), nil
	}
	// First mask the secrets in channel
	channel.Mask()

	bytes, err := json.Marshal(channel)
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal channels: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}
