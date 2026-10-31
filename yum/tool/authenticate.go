package tool

import (
	"context"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/mark3labs/mcp-go/mcp"
)

type AuthenticateTool struct {
	Core *core.Core
}

func (t *AuthenticateTool) Name() string {
	return "authenticate"
}

func (t *AuthenticateTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Starts the YouTube OAuth2 flow and returns the authorization URL to open in a browser. "+
			"redirect_uri is optional and defaults to "+core.DefaultRedirectURI+". "+
			"When a custom redirect_uri is supplied, pass that same URI to accesstoken."),
		mcp.WithString("redirect_uri",
			mcp.Description("Optional. When omitted, "+core.DefaultRedirectURI+" is used. "+
				"When a custom URI is supplied, the caller must pass it again to accesstoken."),
		),
	)
}

func (t *AuthenticateTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Implement the authentication logic here
	// For now, we will just return a dummy response
	redirectURI := request.GetString("redirect_uri", core.DefaultRedirectURI)
	if redirectURI == "" {
		return mcp.NewToolResultError("redirect_uri is required"), nil
	}

	authUrl, err := t.Core.AuthURL(redirectURI)
	if err != nil {
		return mcp.NewToolResultError("Failed to get authentication URL: " + err.Error()), nil
	}

	return mcp.NewToolResultText(authUrl), nil
}
