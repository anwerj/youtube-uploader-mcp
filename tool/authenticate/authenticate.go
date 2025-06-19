package authenticate

import (
	"context"

	"github.com/anwerj/youtube-uploader-mcp/tool"
	"github.com/anwerj/youtube-uploader-mcp/youtube"
	"github.com/mark3labs/mcp-go/mcp"
)

type AuthenticateTool struct {
	tool.Tool
}

func (t *AuthenticateTool) Name() string {
	return "authenticate"
}

func (t *AuthenticateTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Authenticate tools starts the OAuth2 flow for YouTube API."+
			"It returns an authentication URL that the user needs to open in their browser to complete the authentication process."),
		mcp.WithString("redirect_uri",
			mcp.Description("Redirect URI for OAuth2 authentication, default is "+youtube.DefaultRedirectURI),
		),
	)
}

func (t *AuthenticateTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Implement the authentication logic here
	// For now, we will just return a dummy response
	redirectURI := request.GetString("redirect_uri", youtube.DefaultRedirectURI)
	if redirectURI == "" {
		return mcp.NewToolResultError("redirect_uri is required"), nil
	}

	authUrl, err := youtube.AuthURL(redirectURI)
	if err != nil {
		return mcp.NewToolResultError("Failed to get authentication URL: " + err.Error()), nil
	}

	return mcp.NewToolResultText(authUrl), nil
}
