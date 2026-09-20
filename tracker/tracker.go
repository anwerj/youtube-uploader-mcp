package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// Status is the lifecycle state of a tracked job.
type Status string

const (
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

// Job is a snapshot of tracked work. Result is only populated on success;
// Error is only populated on failure.
type Job struct {
	Tool      string
	Status    Status
	Result    *mcp.CallToolResult
	Error     string
	StartedAt time.Time
	UpdatedAt time.Time
}

// Tracker holds in-memory job state. Only this package stores job results.
type Tracker struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// New returns an empty Tracker.
func New() *Tracker {
	return &Tracker{
		jobs: make(map[string]*Job),
	}
}

// MCPRunOptions configures a tracked MCP tool run: which tool is running,
// which fields identify the job (used to derive the job key), and the
// early-return/hard-timeout durations.
type MCPRunOptions struct {
	ToolName    string
	KeyFields   map[string]string // used to derive the job key, sorted by field name
	EarlyReturn time.Duration
	HardTimeout time.Duration
}

// RunTool acts as an extension of a normal Tool.Handle: once control is
// handed to the tracker, fn is fully responsible for producing its own
// CallToolResult. The job key is computed internally from opts.KeyFields
// (sorted by field name for determinism) and is only surfaced back to the
// caller in the running/duplicate responses, so an agent can poll
// check_job_status with it. A synchronous success returns fn's own
// CallToolResult untouched.
func RunTool(
	tr *Tracker,
	opts MCPRunOptions,
	fn func(context.Context) (*mcp.CallToolResult, error),
) (*mcp.CallToolResult, error) {
	jobKey := BuildKeyFromFields(opts.KeyFields)

	tr.mu.Lock()
	if existing, ok := tr.jobs[jobKey]; ok && existing.Status == StatusRunning {
		tr.mu.Unlock()
		return runningResult(jobKey)
	}
	now := time.Now()
	tr.jobs[jobKey] = &Job{
		Tool:      opts.ToolName,
		Status:    StatusRunning,
		StartedAt: now,
		UpdatedAt: now,
	}
	tr.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), opts.HardTimeout)

	type done struct {
		result *mcp.CallToolResult
		err    error
	}
	ch := make(chan done, 1)

	go func() {
		defer cancel()
		result, err := fn(ctx)
		if ctx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("operation still in progress after %v", opts.HardTimeout)
		}
		ch <- done{result: result, err: err}
	}()

	finish := func(d done) {
		tr.mu.Lock()
		job := tr.jobs[jobKey]
		if job != nil {
			job.UpdatedAt = time.Now()
			if d.err != nil {
				job.Status = StatusFailed
				job.Error = d.err.Error()
			} else {
				job.Status = StatusSuccess
				job.Result = d.result
			}
		}
		tr.mu.Unlock()
	}

	select {
	case d := <-ch:
		finish(d)
		if d.err != nil {
			return mcp.NewToolResultError(d.err.Error()), nil
		}
		return d.result, nil
	case <-time.After(opts.EarlyReturn):
		go func() {
			finish(<-ch)
		}()
		return runningResult(jobKey)
	}
}

func runningResult(jobKey string) (*mcp.CallToolResult, error) {
	body, err := json.Marshal(map[string]string{
		"status":  "running",
		"job_key": jobKey,
		"message": "job is still running, please check in sometime with the same job_key",
	})
	if err != nil {
		return mcp.NewToolResultError("failed to marshal pending status: " + err.Error()), nil
	}
	return mcp.NewToolResultText(string(body)), nil
}

// Get returns a snapshot of the job for jobKey.
func Get(t *Tracker, jobKey string) (Job, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	job, ok := t.jobs[jobKey]
	if !ok {
		return Job{}, false
	}
	return *job, true
}
