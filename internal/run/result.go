package run

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/run/views"
)

// Result is shared by the run and the progress runner, which run in separate goroutines.
//
// mu guards the fields copied into resultState, and TestDuration. The setters take it to
// write them, and state takes it to copy them. Every other method works on that copy
// without holding mu, so no method ever takes mu while it already holds it. TestDuration
// is set once the run has finished and is read after that, so it isn't copied.
type Result struct {
	startTime     time.Time
	progressStats *progress.Stats
	views         *views.Views
	LogFilePath   string
	errors        []error
	runOptions    options.RunOptions
	snapshot      progress.Snapshot
	TestDuration  time.Duration
	mu            sync.RWMutex
}

func NewResult(
	runOptions options.RunOptions,
	views *views.Views,
	progressStats *progress.Stats,
) *Result {
	return &Result{
		runOptions:    runOptions,
		views:         views,
		progressStats: progressStats,
	}
}

func (r *Result) SnapshotProgress(period time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.snapshot = r.progressStats.Snapshot(period)
}

func (r *Result) GetTotals() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.snapshot = r.progressStats.Total()
}

func (r *Result) Snapshot() progress.Snapshot {
	return r.state().snapshot
}

func (r *Result) AddError(err error) *Result {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors = append(r.errors, err)
	return r
}

func (r *Result) Error() error {
	return r.state().err()
}

func (r *Result) Summary() *views.ViewContext[views.ResultData] {
	s := r.state()

	return r.views.Result(views.ResultData{
		SuccessfulIterationCount:     s.snapshot.SuccessfulIterationDurations.Count,
		DroppedIterationCount:        s.snapshot.DroppedIterationCount,
		FailedIterationCount:         s.snapshot.FailedIterationDurations.Count,
		SuccessfulIterationDurations: s.snapshot.SuccessfulIterationDurations,
		Duration:                     s.duration(),
		FailedIterationDurations:     s.snapshot.FailedIterationDurations,
		Error:                        s.err(),
		Failed:                       s.failed(r.runOptions),
		LogFilePath:                  r.LogFilePath,
		Iterations:                   s.snapshot.Iterations(),
		IterationsStarted:            s.snapshot.IterationsStarted(),
	})
}

func (r *Result) Failed() bool {
	return r.state().failed(r.runOptions)
}

func (r *Result) Progress() *views.ViewContext[views.ProgressData] {
	s := r.state()

	return r.views.Progress(views.ProgressData{
		Duration:                              s.duration(),
		SuccessfulIterationDurationsForPeriod: s.snapshot.SuccessfulIterationDurationsForPeriod,
		Period:                                s.snapshot.Period,
		FailedIterationCount:                  s.snapshot.FailedIterationDurations.Count,
		DroppedIterationCount:                 s.snapshot.DroppedIterationCount,
		SuccessfulIterationCount:              s.snapshot.SuccessfulIterationDurations.Count,
	})
}

func (r *Result) HasDroppedIterations() bool {
	return r.state().snapshot.DroppedIterationCount > 0
}

func (r *Result) Setup() *views.ViewContext[views.SetupData] {
	return r.views.Setup(views.SetupData{
		Error: r.Error(),
	})
}

func (r *Result) Teardown() *views.ViewContext[views.TeardownData] {
	return r.views.Teardown(views.TeardownData{
		Error: r.Error(),
	})
}

func (r *Result) MaxDurationElapsed() *views.ViewContext[views.TimeoutData] {
	return r.views.Timeout(views.TimeoutData{
		Duration: r.state().duration(),
	})
}

func (r *Result) Interrupted() *views.ViewContext[views.InterruptData] {
	return r.views.Interrupt(views.InterruptData{
		Duration: r.state().duration(),
	})
}

func (r *Result) RecordStarted() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.startTime = time.Now()
}

func (r *Result) RecordTestFinished() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.TestDuration = time.Since(r.startTime)
}

func (r *Result) MaxIterationsReached() *views.ViewContext[views.MaxIterationsReachedData] {
	return r.views.MaxIterationsReached(views.MaxIterationsReachedData{
		Duration: r.state().duration(),
	})
}

// state returns a copy of the fields guarded by mu. It is the only method that takes mu to
// read them.
func (r *Result) state() resultState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// AddError only appends, so the errors up to the copied length never change and the
	// copy can share them.
	return resultState{
		snapshot:  r.snapshot,
		errors:    r.errors,
		startTime: r.startTime,
	}
}

// resultState is a copy of the fields of Result that are guarded by mu.
type resultState struct {
	snapshot  progress.Snapshot
	startTime time.Time
	errors    []error
}

func (s resultState) err() error {
	if s.errors == nil {
		return nil
	}

	if len(s.errors) == 1 {
		return s.errors[0]
	}

	errorStrings := make([]string, len(s.errors))
	for i := range len(s.errors) {
		errorStrings[i] = fmt.Sprintf("Error %d: %s", i, s.errors[i].Error())
	}

	return errors.New(strings.Join(errorStrings, "; "))
}

func (s resultState) failed(opts options.RunOptions) bool {
	return s.err() != nil ||
		(!opts.IgnoreDropped && s.snapshot.DroppedIterationCount > 0) ||
		(opts.MaxFailures == 0 && opts.MaxFailuresRate == 0 && s.snapshot.FailedIterationDurations.Count > 0) ||
		(opts.MaxFailures > 0 && s.snapshot.FailedIterationDurations.Count > opts.MaxFailures) ||
		(opts.MaxFailuresRate > 0 && (s.snapshot.FailedIterationsRate() > uint64(opts.MaxFailuresRate)))
}

func (s resultState) duration() time.Duration {
	if s.startTime.IsZero() {
		return 0
	}

	return time.Since(s.startTime)
}
