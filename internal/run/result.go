package run

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

type Result struct {
	startTime     time.Time
	progressStats *progress.Stats
	templates     *templates.Templates
	LogFile       string
	errors        []error
	runOptions    options.RunOptions
	snapshot      progress.Snapshot
	TestDuration  time.Duration
	mu            sync.RWMutex
}

func NewResult(
	runOptions options.RunOptions,
	templates *templates.Templates,
	progressStats *progress.Stats,
) Result {
	return Result{
		runOptions:    runOptions,
		templates:     templates,
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
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.snapshot
}

func (r *Result) AddError(err error) *Result {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.errors = append(r.errors, err)
	return r
}

func (r *Result) Error() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.errors == nil {
		return nil
	}

	if len(r.errors) == 1 {
		return r.errors[0]
	}

	errorStrings := make([]string, len(r.errors))
	for i := range len(r.errors) {
		errorStrings[i] = fmt.Sprintf("Error %d: %s", i, r.errors[i].Error())
	}

	return errors.New(strings.Join(errorStrings, "; "))
}

func (r *Result) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Result(templates.ResultData{
		SuccessfulIterationCount:     r.snapshot.SuccessfulIterationDurations.Count,
		DroppedIterationCount:        r.snapshot.DroppedIterationCount,
		FailedIterationCount:         r.snapshot.FailedIterationDurations.Count,
		SuccessfulIterationDurations: r.snapshot.SuccessfulIterationDurations,
		Duration:                     r.duration(),
		FailedIterationDurations:     r.snapshot.FailedIterationDurations,
		Error:                        r.Error(),
		Failed:                       r.Failed(),
		LogFile:                      r.LogFile,
		Iterations:                   r.snapshot.Iterations(),
		IterationsStarted:            r.snapshot.IterationsStarted(),
	})
}

func (r *Result) Failed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	opts := r.runOptions

	return r.Error() != nil ||
		(!opts.IgnoreDropped && r.snapshot.DroppedIterationCount > 0) ||
		(opts.MaxFailures == 0 && opts.MaxFailuresRate == 0 && r.snapshot.FailedIterationDurations.Count > 0) ||
		(opts.MaxFailures > 0 && r.snapshot.FailedIterationDurations.Count > opts.MaxFailures) ||
		(opts.MaxFailuresRate > 0 && (r.snapshot.FailedIterationsRate() > uint64(opts.MaxFailuresRate)))
}

func (r *Result) Progress() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Progress(templates.ProgressData{
		Duration:                              r.duration(),
		SuccessfulIterationDurationsForPeriod: r.snapshot.SuccessfulIterationDurationsForPeriod,
		Period:                                r.snapshot.Period,
		FailedIterationCount:                  r.snapshot.FailedIterationDurations.Count,
		DroppedIterationCount:                 r.snapshot.DroppedIterationCount,
		SuccessfulIterationCount:              r.snapshot.SuccessfulIterationDurations.Count,
	})
}

func (r *Result) HasDroppedIterations() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.snapshot.DroppedIterationCount > 0
}

func (r *Result) Setup() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Setup(templates.SetupData{
		Error: r.Error(),
	})
}

func (r *Result) Teardown() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Teardown(templates.TeardownData{
		Error: r.Error(),
	})
}

func (r *Result) MaxDurationElapsed() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Timeout(templates.TimeoutData{
		Duration: r.duration(),
	})
}

func (r *Result) Interrupted() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Interrupt(templates.InterruptData{
		Duration: r.duration(),
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

func (r *Result) MaxIterationsReached() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.MaxIterationsReached(templates.MaxIterationsReachedData{
		Duration: r.duration(),
	})
}

func (r *Result) duration() time.Duration {
	if r.startTime.IsZero() {
		return 0
	}

	return time.Since(r.startTime)
}
