package run

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

type Result struct {
	mu        sync.RWMutex
	errors    []error
	startTime time.Time
	templates *templates.Templates

	SuccessfulIterationCount     uint64
	SuccessfulIterationDurations DurationPercentileMap
	FailedIterationCount         uint64
	FailedIterationDurations     DurationPercentileMap
	MaxFailedIterations          uint64
	MaxFailedIterationsRate      int
	TestDuration                 time.Duration
	IgnoreDropped                bool
	DroppedIterationCount        uint64
	RecentSuccessfulIterations   uint64
	RecentDuration               time.Duration
	LogFile                      string
}

func (r *Result) SetMetrics(
	result metrics.ResultType,
	count uint64,
	quantiles []*io_prometheus_client.Quantile,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.RecentDuration = 1 * time.Second
	switch result {
	case metrics.SucessResult:
		r.RecentSuccessfulIterations = count - r.SuccessfulIterationCount
		r.SuccessfulIterationCount = count
		r.SuccessfulIterationDurations = parseQuantiles(quantiles)
		return
	case metrics.FailedResult:
		r.FailedIterationCount = count
		r.FailedIterationDurations = parseQuantiles(quantiles)
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
	r.RecentSuccessfulIterations = 0
	r.SuccessfulIterationDurations = map[float64]time.Duration{}
}

func (r *Result) IncrementMetrics(
	duration time.Duration,
	result metrics.ResultType,
	count uint64,
	quantiles []*io_prometheus_client.Quantile,
) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RecentDuration = duration
	switch result {
	case metrics.SucessResult:
		r.RecentSuccessfulIterations = count
		r.SuccessfulIterationCount += count
		r.SuccessfulIterationDurations = parseQuantiles(quantiles)
		return
	case metrics.FailedResult:
		r.FailedIterationCount += count
		r.FailedIterationDurations = parseQuantiles(quantiles)
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

func parseQuantiles(quantiles []*io_prometheus_client.Quantile) DurationPercentileMap {
	m := make(DurationPercentileMap)
	for _, quantile := range quantiles {
		m[quantile.GetQuantile()] = time.Duration(quantile.GetQuantile())
	}
	return m
}

func (r *Result) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(r.templates.Result, r)
}

func (r *Result) Failed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Error() != nil ||
		(!r.IgnoreDropped && r.DroppedIterationCount > 0) ||
		(r.MaxFailedIterations == 0 && r.MaxFailedIterationsRate == 0 && r.FailedIterationCount > 0) ||
		(r.MaxFailedIterations > 0 && r.FailedIterationCount > r.MaxFailedIterations) ||
		(r.MaxFailedIterationsRate > 0 && (r.FailedIterationCount*100/r.Iterations() > uint64(r.MaxFailedIterationsRate)))
}

func (r *Result) Progress() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(r.templates.Progress, r)
}

func (r *Result) Duration() time.Duration {
	if r.StartTime().IsZero() {
		return 0
	}

	return time.Since(r.StartTime())
}

func (r *Result) Setup() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(r.templates.Setup, r)
}

func (r *Result) Teardown() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(r.templates.Teardown, r)
}

func (r *Result) MaxDurationElapsed() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(r.templates.Timeout, r)
}

func (r *Result) Interrupted() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(r.templates.Interrupt, r)
}

func (r *Result) Iterations() uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.FailedIterationCount + r.SuccessfulIterationCount + r.DroppedIterationCount
}

func (r *Result) IterationsStarted() uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.FailedIterationCount + r.SuccessfulIterationCount
}

func (r *Result) StartTime() time.Time {
	return r.startTime
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
	return renderTemplate(r.templates.MaxIterationsReached, r)
}
