package accesstoken

import (
	"context"

	"github.com/anwerj/youtube-uploader-mcp/tool"
	"github.com/anwerj/youtube-uploader-mcp/youtube"
	"github.com/mark3labs/mcp-go/mcp"
)

type AccessTokenTool struct {
	tool.Tool
}

func (t *AccessTokenTool) Name() string {
	return "accesstoken"
}

func (t *AccessTokenTool) Define(ctx context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("AccessToken tool retrieves the access token for YouTube API authentication."+
			"It requires the client secret file and code to fetch the access token."+
			"User must first authenticate using the authenticate tool to get the code."),
		mcp.WithString("client_secret_file",
			mcp.Required(),
			mcp.Description("Client ID for OAuth2 authentication"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("The code part received on redirected URL, Only send the code part, not the full URL."),
		),
	)
}

func (t *AccessTokenTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Implement the access token retrieval logic here
	// For now, we will just return a dummy response
	clientSecretFile := request.GetString("client_secret_file", "./client_secrets.json")
	code := request.GetString("code", "")
	if clientSecretFile == "" {
		return mcp.NewToolResultError("client_secret_file is required"), nil
	}
	if code == "" {
		return mcp.NewToolResultError("code is required"), nil
	}

	// Here you would typically call a function to get the access token using the client secret file and code
	// For example:
	accessToken, err := youtube.GetAccessToken(clientSecretFile, code)
	if err != nil {
		return mcp.NewToolResultError("Failed to get access token: " + err.Error()), nil
	}
	// save the access token or use it as needed
	err = youtube.SaveToken(accessToken)
	if err != nil {
		return mcp.NewToolResultError("Failed to save access token: " + err.Error()), nil
	}

	return mcp.NewToolResultText("Access token retrieved successfully: " + accessToken.Expiry.GoString()), nil
}
