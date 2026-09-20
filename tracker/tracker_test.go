package tracker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunSyncSuccess(t *testing.T) {
	tr := New()
	out := Run(tr, "test", "k1", time.Second, time.Minute, func(ctx context.Context) (int, error) {
		return 42, nil
	})
	if out.Kind != SyncSuccess || out.Result != 42 {
		t.Fatalf("got kind=%v result=%d", out.Kind, out.Result)
	}
	job, ok := Get[int](tr, "k1")
	if !ok || job.Status != StatusSuccess || job.Result != 42 {
		t.Fatalf("job: ok=%v %+v", ok, job)
	}
}

func TestRunAsyncThenSuccess(t *testing.T) {
	tr := New()
	start := make(chan struct{})
	out := Run(tr, "test", "k2", 5*time.Millisecond, time.Minute, func(ctx context.Context) (string, error) {
		close(start)
		time.Sleep(20 * time.Millisecond)
		return "done", nil
	})
	if out.Kind != AsyncRunning {
		t.Fatalf("expected async, got %v", out.Kind)
	}
	<-start
	time.Sleep(30 * time.Millisecond)
	job, ok := Get[string](tr, "k2")
	if !ok || job.Status != StatusSuccess || job.Result != "done" {
		t.Fatalf("job: ok=%v %+v", ok, job)
	}
}

func TestRunDuplicate(t *testing.T) {
	tr := New()
	block := make(chan struct{})
	Run(tr, "test", "k3", time.Minute, time.Minute, func(ctx context.Context) (int, error) {
		<-block
		return 1, nil
	})
	time.Sleep(10 * time.Millisecond)
	out := Run(tr, "test", "k3", time.Millisecond, time.Minute, func(ctx context.Context) (int, error) {
		return 2, nil
	})
	if out.Kind != Duplicate {
		t.Fatalf("expected duplicate, got %v", out.Kind)
	}
	close(block)
}

func TestRunHardTimeout(t *testing.T) {
	tr := New()
	out := Run(tr, "test", "k4", 5*time.Millisecond, 20*time.Millisecond, func(ctx context.Context) (int, error) {
		time.Sleep(50 * time.Millisecond)
		return 0, nil
	})
	if out.Kind != AsyncRunning {
		t.Fatalf("expected async, got %v", out.Kind)
	}
	time.Sleep(80 * time.Millisecond)
	job, ok := Get[int](tr, "k4")
	if !ok || job.Status != StatusFailed {
		t.Fatalf("job: ok=%v status=%s err=%q", ok, job.Status, job.Error)
	}
}

func TestRunSyncError(t *testing.T) {
	tr := New()
	want := errors.New("boom")
	out := Run(tr, "test", "k5", time.Second, time.Minute, func(ctx context.Context) (int, error) {
		return 0, want
	})
	if out.Kind != SyncError {
		t.Fatalf("got %v", out.Kind)
	}
	job, _ := Get[int](tr, "k5")
	if job.Status != StatusFailed {
		t.Fatalf("status %s", job.Status)
	}
}
