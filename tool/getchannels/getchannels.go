package getchannels

import (
	"context"
	"encoding/json"

	"github.com/anwerj/youtube-uploader-mcp/youtube"
	"github.com/mark3labs/mcp-go/mcp"
)

type GetChannelsTool struct {
}

func (t *GetChannelsTool) Name() string {
	return "channels"
}

func (t *GetChannelsTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Get the list of channels that the user has access to."+
			"Use this get list of authenticated channels if user want refresh token or upload video."+
			"It is recommended not to use this tool too often, as it may confuse user."+
			"Once user has chosen a channel, keep it in memory for future use."),
	)
}

func (t *GetChannelsTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channels, err := youtube.ReadChannels(false)
	if err != nil {
		return mcp.NewToolResultError("Failed to get channels: " + err.Error()), nil
	}
	// remove token from channels
	for _, channel := range channels {
		channel.Mask()
	}

	bytes, err := json.Marshal(channels)
	if err != nil {
		return mcp.NewToolResultError("Failed to marshal channels: " + err.Error()), nil
	}

	return mcp.NewToolResultText(string(bytes)), nil
}
