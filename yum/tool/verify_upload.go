package tool

import (
	"context"
	"encoding/json"
	"time"

	"github.com/anwerj/youtube-uploader-mcp/core"
	"github.com/anwerj/youtube-uploader-mcp/tracker"
	"github.com/mark3labs/mcp-go/mcp"
)

type VerifyUploadTool struct {
	Tracker *tracker.Tracker
}

func (t *VerifyUploadTool) Name() string {
	return "verify_upload"
}

func (t *VerifyUploadTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Use if upload_video returns status running."),
		mcp.WithString("channel_id",
			mcp.Required(),
			mcp.Description("Channel ID used for the upload"),
		),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the video file that was uploaded"),
		),
	)
}

func (t *VerifyUploadTool) Handle(
	_ context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	channelID := request.GetString("channel_id", "")
	filePath := request.GetString("file_path", "")
	if channelID == "" || filePath == "" {
		return mcp.NewToolResultError("channel_id and file_path are required"), nil
	}

	jobKey := tracker.UploadJobKey(channelID, filePath)
	job, ok := tracker.Get[core.Video](t.Tracker, jobKey)
	if !ok {
		return mcp.NewToolResultError("no upload record found for this channel_id and file_path"), nil
	}

	switch job.Status {
	case tracker.StatusSuccess:
		bytes, err := json.Marshal(job.Result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal the video: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(bytes)), nil
	case tracker.StatusFailed:
		body, err := json.Marshal(map[string]string{
			"status": "failed",
			"error":  job.Error,
		})
		if err != nil {
			return mcp.NewToolResultError("failed to marshal status: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	case tracker.StatusRunning:
		body, err := json.Marshal(map[string]interface{}{
			"status":     "running",
			"tool":       job.Tool,
			"started_at": job.StartedAt.Format(time.RFC3339),
			"message":    "Upload is still in progress.",
		})
		if err != nil {
			return mcp.NewToolResultError("failed to marshal status: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(body)), nil
	default:
		return mcp.NewToolResultError("unknown upload status"), nil
	}
}
