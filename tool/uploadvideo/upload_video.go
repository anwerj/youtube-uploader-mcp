package uploadvideo

import (
	"context"
	"strings"

	"github.com/anwerj/youtube-uploader-mcp/tool"
	"github.com/anwerj/youtube-uploader-mcp/youtube"
	"github.com/mark3labs/mcp-go/mcp"
)

type UploadVideoTool struct {
	tool.Tool
}

func (t *UploadVideoTool) Name() string {
	return "upload_video"
}

func (t *UploadVideoTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Upload a video to YouTube, taking advantages of AI to generate descriptions, title and tags"),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the video file"),
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
	)
}

func (t *UploadVideoTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath := request.GetString("file_path", "")
	description := request.GetString("description", "")
	title := request.GetString("title", "")
	tags := request.GetString("tags", "")
	categoryID := request.GetString("category_id", "")
	if filePath == "" || description == "" || title == "" || tags == "" || categoryID == "" {
		return mcp.NewToolResultError("all fields are required: file_path, description, title, tags, category_id"), nil
	}

	token, err := youtube.ReadToken()
	if err != nil {
		return mcp.NewToolResultError("Failed to load token: " + err.Error()), nil
	}

	video := &youtube.Video{
		Path:        filePath,
		Title:       title,
		Description: description,
		Tags:        strings.Split(tags, ","), // Assuming tags are comma-separated, you might want to split them
		CategoryID:  categoryID,
	}
	id, err := video.Upload(ctx, token) // Assuming you have a way to get the OAuth2 token
	if err != nil {
		return mcp.NewToolResultError("Failed to upload video: " + err.Error()), nil
	}

	return mcp.NewToolResultText("Video uploaded successfully with ID:" + id), nil
}
