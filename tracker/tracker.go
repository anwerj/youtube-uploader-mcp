package tracker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status is the lifecycle state of a tracked job.
type Status string

const (
	StatusRunning Status = "running"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

// Job is a snapshot of tracked work for result type T.
type Job[T any] struct {
	Tool      string    `json:"tool"`
	Status    Status    `json:"status"`
	Result    T         `json:"-"`
	Error     string    `json:"error,omitempty"`
	Message   string    `json:"message,omitempty"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RunKind describes how Run returned to the caller.
type RunKind int

const (
	SyncSuccess  RunKind = 1
	SyncError    RunKind = 2
	AsyncRunning RunKind = 3
	Duplicate    RunKind = 4
)

// RunOutcome is the immediate result of Run.
type RunOutcome[T any] struct {
	Kind   RunKind
	Result T
	Err    error
}

// Tracker holds in-memory job state. Only this package stores job results.
type Tracker struct {
	mu   sync.RWMutex
	jobs map[string]*jobState
}

type jobState struct {
	tool      string
	status    Status
	err       string
	message   string
	startedAt time.Time
	updatedAt time.Time
	result    interface{}
}

// New returns an empty Tracker.
func New() *Tracker {
	return &Tracker{
		jobs: make(map[string]*jobState),
	}
}

// Run executes fn in a goroutine. The caller must not start fn with go itself.
// If fn finishes before earlyReturn, Run returns synchronously.
// Otherwise Run returns AsyncRunning while fn continues until hardTimeout.
func Run[T any](
	t *Tracker,
	toolName string,
	jobKey string,
	earlyReturn time.Duration,
	hardTimeout time.Duration,
	fn func(context.Context) (T, error),
) RunOutcome[T] {
	var zero T
	if t == nil {
		return RunOutcome[T]{Kind: SyncError, Err: fmt.Errorf("tracker is nil")}
	}

	t.mu.Lock()
	if existing, ok := t.jobs[jobKey]; ok && existing.status == StatusRunning {
		t.mu.Unlock()
		return RunOutcome[T]{Kind: Duplicate}
	}
	now := time.Now()
	t.jobs[jobKey] = &jobState{
		tool:      toolName,
		status:    StatusRunning,
		startedAt: now,
		updatedAt: now,
	}
	t.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), hardTimeout)

	type done struct {
		value T
		err   error
	}
	ch := make(chan done, 1)

	go func() {
		defer cancel()
		value, err := fn(ctx)
		if ctx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("operation still in progress after %v", hardTimeout)
		}
		ch <- done{value: value, err: err}
	}()

	finish := func(d done) {
		t.mu.Lock()
		st := t.jobs[jobKey]
		if st == nil {
			t.mu.Unlock()
			return
		}
		st.updatedAt = time.Now()
		if d.err != nil {
			st.status = StatusFailed
			st.err = d.err.Error()
		} else {
			st.status = StatusSuccess
			st.result = d.value
		}
		t.mu.Unlock()
	}

	select {
	case d := <-ch:
		finish(d)
		if d.err != nil {
			return RunOutcome[T]{Kind: SyncError, Err: d.err}
		}
		return RunOutcome[T]{Kind: SyncSuccess, Result: d.value}
	case <-time.After(earlyReturn):
		go func() {
			d := <-ch
			finish(d)
		}()
		return RunOutcome[T]{Kind: AsyncRunning, Result: zero}
	}
}

// Get returns the current job for jobKey and result type T.
func Get[T any](t *Tracker, jobKey string) (Job[T], bool) {
	var zero Job[T]
	if t == nil {
		return zero, false
	}

	t.mu.RLock()
	st, ok := t.jobs[jobKey]
	t.mu.RUnlock()
	if !ok {
		return zero, false
	}

	job := Job[T]{
		Tool:      st.tool,
		Status:    st.status,
		Error:     st.err,
		Message:   st.message,
		StartedAt: st.startedAt,
		UpdatedAt: st.updatedAt,
	}
	if st.status == StatusSuccess && st.result != nil {
		v, ok := st.result.(T)
		if !ok {
			return zero, false
		}
		job.Result = v
	}
	return job, true
}
