package tracker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRunToolSyncSuccessReturnsFnResultUntouched(t *testing.T) {
	tr := New()
	want := mcp.NewToolResultText("hello")
	opts := MCPRunOptions{
		ToolName:    "test",
		KeyFields:   map[string]string{"a": "1"},
		EarlyReturn: time.Second,
		HardTimeout: time.Minute,
	}

	got, err := RunTool(tr, opts, func(ctx context.Context) (*mcp.CallToolResult, error) {
		return want, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected fn's CallToolResult to be returned untouched (no job_key injected)")
	}
}

func TestRunToolAsyncRunningIncludesJobKey(t *testing.T) {
	tr := New()
	opts := MCPRunOptions{
		ToolName:    "test",
		KeyFields:   map[string]string{"channel_id": "c1", "file_path": "f1"},
		EarlyReturn: 5 * time.Millisecond,
		HardTimeout: time.Minute,
	}
	block := make(chan struct{})

	result, err := RunTool(tr, opts, func(ctx context.Context) (*mcp.CallToolResult, error) {
		<-block
		return mcp.NewToolResultText("done"), nil
	})
	close(block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(text.Text), &body); err != nil {
		t.Fatalf("failed to unmarshal running body: %v", err)
	}
	if body["status"] != "running" {
		t.Fatalf("expected status=running, got %q", body["status"])
	}
	if want := BuildKeyFromFields(opts.KeyFields); body["job_key"] != want {
		t.Fatalf("expected job_key %q, got %q", want, body["job_key"])
	}
}

func TestRunToolDuplicateIncludesJobKey(t *testing.T) {
	tr := New()
	keyFields := map[string]string{"channel_id": "c1", "file_path": "f1"}
	block := make(chan struct{})

	go func() {
		_, _ = RunTool(tr, MCPRunOptions{
			ToolName:    "test",
			KeyFields:   keyFields,
			EarlyReturn: time.Minute,
			HardTimeout: time.Minute,
		}, func(ctx context.Context) (*mcp.CallToolResult, error) {
			<-block
			return mcp.NewToolResultText("done"), nil
		})
	}()
	time.Sleep(20 * time.Millisecond) // let the job register as running

	result, err := RunTool(tr, MCPRunOptions{
		ToolName:    "test",
		KeyFields:   keyFields,
		EarlyReturn: time.Millisecond,
		HardTimeout: time.Minute,
	}, func(ctx context.Context) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("should not run"), nil
	})
	close(block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(text.Text), &body); err != nil {
		t.Fatalf("failed to unmarshal duplicate body: %v", err)
	}
	if want := BuildKeyFromFields(keyFields); body["job_key"] != want {
		t.Fatalf("expected job_key %q on duplicate response, got %q", want, body["job_key"])
	}
}

func TestRunToolSyncErrorIsWrappedAsErrorResult(t *testing.T) {
	tr := New()
	opts := MCPRunOptions{
		ToolName:    "test",
		KeyFields:   map[string]string{"a": "1"},
		EarlyReturn: time.Second,
		HardTimeout: time.Minute,
	}

	result, err := RunTool(tr, opts, func(ctx context.Context) (*mcp.CallToolResult, error) {
		return nil, errors.New("boom")
	})
	if err != nil {
		t.Fatalf("unexpected transport-level error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected an error result")
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok || text.Text != "boom" {
		t.Fatalf("expected error text %q, got %+v", "boom", result.Content)
	}
}

func TestRunToolHardTimeout(t *testing.T) {
	tr := New()
	keyFields := map[string]string{"a": "1"}

	_, err := RunTool(tr, MCPRunOptions{
		ToolName:    "test",
		KeyFields:   keyFields,
		EarlyReturn: 5 * time.Millisecond,
		HardTimeout: 20 * time.Millisecond,
	}, func(ctx context.Context) (*mcp.CallToolResult, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(80 * time.Millisecond) // let the hard timeout fire and finish() record it

	jobKey := BuildKeyFromFields(keyFields)
	job, ok := Get(tr, jobKey)
	if !ok {
		t.Fatalf("expected job to be present in registry")
	}
	if job.Status != StatusFailed {
		t.Fatalf("expected status=failed after hard timeout, got %q", job.Status)
	}
	if job.Error == "" {
		t.Fatalf("expected a timeout error message to be recorded")
	}
}
