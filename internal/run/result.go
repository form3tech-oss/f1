package run

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

type Result struct {
	startTime                    time.Time
	failedIterationDurations     metrics.DurationPercentileMap
	templates                    *templates.Templates
	successfulIterationDurations metrics.DurationPercentileMap
	LogFile                      string
	errors                       []error
	runOptions                   options.RunOptions
	FailedIterationCount         uint64
	DroppedIterationCount        uint64
	recentSuccessfulIterations   uint64
	recentDuration               time.Duration
	SuccessfulIterationCount     uint64
	TestDuration                 time.Duration
	mu                           sync.RWMutex
}

func NewResult(runOptions options.RunOptions, templates *templates.Templates) Result {
	return Result{
		runOptions: runOptions,
		templates:  templates,
	}
}

func (r *Result) SetMetrics(
	result metrics.ResultType,
	count uint64,
	quantiles []*io_prometheus_client.Quantile,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.recentDuration = 1 * time.Second
	switch result {
	case metrics.SucessResult:
		r.recentSuccessfulIterations = count - r.SuccessfulIterationCount
		r.SuccessfulIterationCount = count
		r.successfulIterationDurations = parseQuantiles(quantiles)
		return
	case metrics.FailedResult:
		r.FailedIterationCount = count
		r.failedIterationDurations = parseQuantiles(quantiles)
		return
	case metrics.DroppedResult:
		r.DroppedIterationCount = count
		return
	case metrics.UnknownResult:
	}
}

func (r *Result) ClearProgressMetrics() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recentSuccessfulIterations = 0
	r.successfulIterationDurations = map[float64]time.Duration{}
}

func (r *Result) IncrementMetrics(
	duration time.Duration,
	result metrics.ResultType,
	count uint64,
	quantiles []*io_prometheus_client.Quantile,
) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recentDuration = duration
	switch result {
	case metrics.SucessResult:
		r.recentSuccessfulIterations = count
		r.SuccessfulIterationCount += count
		r.successfulIterationDurations = parseQuantiles(quantiles)
		return
	case metrics.FailedResult:
		r.FailedIterationCount += count
		r.failedIterationDurations = parseQuantiles(quantiles)
		return
	case metrics.DroppedResult:
		r.DroppedIterationCount += count
		return
	case metrics.UnknownResult:
	}
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

func parseQuantiles(quantiles []*io_prometheus_client.Quantile) metrics.DurationPercentileMap {
	m := make(metrics.DurationPercentileMap)
	for _, quantile := range quantiles {
		m[quantile.GetQuantile()] = time.Duration(quantile.GetQuantile())
	}
	return m
}

func (r *Result) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Result(templates.ResultData{
		SuccessfulIterationCount:     r.SuccessfulIterationCount,
		DroppedIterationCount:        r.DroppedIterationCount,
		FailedIterationCount:         r.FailedIterationCount,
		SuccessfulIterationDurations: r.successfulIterationDurations,
		Duration:                     r.duration(),
		FailedIterationDurations:     r.failedIterationDurations,
		Error:                        r.Error(),
		Failed:                       r.Failed(),
		LogFile:                      r.LogFile,
		Iterations:                   r.iterations(),
		IterationsStarted:            r.iterationsStarted(),
	})
}

func (r *Result) Failed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	opts := r.runOptions

	return r.Error() != nil ||
		(!opts.IgnoreDropped && r.DroppedIterationCount > 0) ||
		(opts.MaxFailures == 0 && opts.MaxFailuresRate == 0 && r.FailedIterationCount > 0) ||
		(opts.MaxFailures > 0 && r.FailedIterationCount > opts.MaxFailures) ||
		(opts.MaxFailuresRate > 0 && (r.FailedIterationCount*100/r.iterations() > uint64(opts.MaxFailuresRate)))
}

func (r *Result) Progress() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.templates.Progress(templates.ProgressData{
		SuccessfulIterationDurations: r.successfulIterationDurations,
		Duration:                     r.duration(),
		RecentSuccessfulIterations:   r.recentSuccessfulIterations,
		RecentDuration:               r.recentDuration,
		FailedIterationCount:         r.FailedIterationCount,
		DroppedIterationCount:        r.DroppedIterationCount,
		SuccessfulIterationCount:     r.SuccessfulIterationCount,
	})
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

func (r *Result) iterations() uint64 {
	return r.FailedIterationCount + r.SuccessfulIterationCount + r.DroppedIterationCount
}

func (r *Result) iterationsStarted() uint64 {
	return r.FailedIterationCount + r.SuccessfulIterationCount
}

func (r *Result) duration() time.Duration {
	if r.startTime.IsZero() {
		return 0
	}

	return time.Since(r.startTime)
}
