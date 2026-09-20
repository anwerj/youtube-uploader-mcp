package tool

import (
	"context"

	"github.com/anwerj/youtube-uploader-mcp/tracker"
	"github.com/mark3labs/mcp-go/mcp"
)

// CheckJobStatusTool is a generic status-check tool for any long-running job
// tracked via tracker.RunTool. It is not tied to a specific tool (e.g.
// upload_video) — it only needs the job_key that tool returned.
type CheckJobStatusTool struct {
	Tracker *tracker.Tracker
}

func (t *CheckJobStatusTool) Name() string {
	return "check_job_status"
}

func (t *CheckJobStatusTool) Define(context.Context) mcp.Tool {
	return mcp.NewTool(t.Name(),
		mcp.WithDescription("Check the status of a long-running job (e.g. an upload) using the job_key returned by the tool call that started it."),
		mcp.WithString("job_key",
			mcp.Required(),
			mcp.Description("job_key returned by the original tool call"),
		),
	)
}

func (t *CheckJobStatusTool) Handle(
	_ context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	jobKey := request.GetString("job_key", "")
	if jobKey == "" {
		return mcp.NewToolResultError("job_key is required"), nil
	}

	job, ok := tracker.Get(t.Tracker, jobKey)
	if !ok {
		return mcp.NewToolResultError("job is not present in registry"), nil
	}

	switch job.Status {
	case tracker.StatusRunning:
		return mcp.NewToolResultText("job is still running, please check in sometime with the same job_key"), nil
	case tracker.StatusSuccess:
		return job.Result, nil
	case tracker.StatusFailed:
		return mcp.NewToolResultError(job.Error), nil
	default:
		return mcp.NewToolResultError("unknown job status"), nil
	}
}
