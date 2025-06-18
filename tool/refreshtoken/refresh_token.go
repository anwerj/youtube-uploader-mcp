package refreshtoken

import (
	"context"

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
			"This tool is useful for maintaining a valid access token without user intervention."),
	)
}

func (t *RefreshTokenTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// first load the token from the file
	token, err := youtube.ReadToken()
	if err != nil {
		return mcp.NewToolResultError("Failed to load token: " + err.Error()), nil
	}
	// Check if the token is already valid
	if token.ExpiresIn == 0 {
		return mcp.NewToolResultText("Invalid access token loaded:. "), nil
	}

	// Here you would typically call a function to refresh the access token using the client secret file
	// For example:
	accessToken, err := youtube.RefreshAccessToken(token)
	if err != nil {
		return mcp.NewToolResultError("Failed to refresh access token: " + err.Error()), nil
	}
	// Save the refreshed access token
	err = youtube.SaveToken(accessToken)
	if err != nil {
		return mcp.NewToolResultError("Failed to save refreshed access token: " + err.Error()), nil
	}

	// For now, we will just return a dummy response
	return mcp.NewToolResultText("Access token refreshed successfully."), nil
}
