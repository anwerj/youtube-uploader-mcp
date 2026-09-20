package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// MCPRunOptions customizes the async-running and duplicate-job tool responses.
type MCPRunOptions struct {
	DuplicateError string
	RunningMessage string
	PendingFields  map[string]string
}

// RunTool runs fn through Run and maps the outcome to an MCP CallToolResult.
// Long-running MCP handlers should use this instead of reimplementing select/MCP JSON.
func RunTool[T any](
	tr *Tracker,
	toolName string,
	jobKey string,
	earlyReturn time.Duration,
	hardTimeout time.Duration,
	fn func(context.Context) (T, error),
	marshal func(T) ([]byte, error),
	opts MCPRunOptions,
) (*mcp.CallToolResult, error) {
	out := Run(tr, toolName, jobKey, earlyReturn, hardTimeout, fn)

	switch out.Kind {
	case SyncSuccess:
		bytes, err := marshal(out.Result)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal result: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(bytes)), nil
	case SyncError:
		return mcp.NewToolResultError(out.Err.Error()), nil
	case Duplicate:
		msg := opts.DuplicateError
		if msg == "" {
			msg = "a job with this key is already in progress"
		}
		return mcp.NewToolResultError(msg), nil
	case AsyncRunning:
		body := map[string]string{
			"status": "running",
		}
		for k, v := range opts.PendingFields {
			body[k] = v
		}
		if body["message"] == "" {
			body["message"] = opts.RunningMessage
		}
		if body["message"] == "" {
			body["message"] = "Operation is continuing in the background. Poll the tool-specific verify endpoint until status is success or failed."
		}
		bytes, err := json.Marshal(body)
		if err != nil {
			return mcp.NewToolResultError("failed to marshal pending status: " + err.Error()), nil
		}
		return mcp.NewToolResultText(string(bytes)), nil
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unexpected tracker outcome: %d", out.Kind)), nil
	}
}
