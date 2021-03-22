package run

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"text/template"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
)

type RunResult struct {
	mu                           sync.RWMutex
	errors                       []error
	SuccessfulIterationCount     uint64
	SuccessfulIterationDurations DurationPercentileMap
	FailedIterationCount         uint64
	FailedIterationDurations     DurationPercentileMap
	startTime                    time.Time
	TestDuration                 time.Duration
	LogFile                      string
	IgnoreDropped                bool
	DroppedIterationCount        uint64
	RecentSuccessfulIterations   uint64
	RecentDuration               time.Duration
}

func (r *RunResult) SetMetrics(result string, stage string, count uint64, quantiles []*io_prometheus_client.Quantile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if stage != IterationStage {
		return
	}
	r.RecentDuration = 1 * time.Second
	switch result {
	case "success":
		r.RecentSuccessfulIterations = count - r.SuccessfulIterationCount
		r.SuccessfulIterationCount = count
		r.SuccessfulIterationDurations = parseQuantiles(quantiles)
		return
	case "fail":
		r.FailedIterationCount = count
		r.FailedIterationDurations = parseQuantiles(quantiles)
		return
	case "dropped":
		r.DroppedIterationCount = count
		return
	default:
		log.Fatalf("unknown result type %s", result)
	}
}

func (r *RunResult) ClearProgressMetrics() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RecentSuccessfulIterations = 0
	r.SuccessfulIterationDurations = map[float64]time.Duration{}
}

func (r *RunResult) IncrementMetrics(duration time.Duration, result string, stage string, count uint64, quantiles []*io_prometheus_client.Quantile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if stage != IterationStage {
		return
	}
	r.RecentDuration = duration
	switch result {
	case "success":
		r.RecentSuccessfulIterations = count
		r.SuccessfulIterationCount += count
		r.SuccessfulIterationDurations = parseQuantiles(quantiles)
		return
	case "fail":
		r.FailedIterationCount += count
		r.FailedIterationDurations = parseQuantiles(quantiles)
		return
	case "dropped":
		r.DroppedIterationCount += count
		return
	default:
		log.Fatalf("unknown result type %s", result)
	}
}

func (r *RunResult) AddError(err error) *RunResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, err)
	return r
}

func (r *RunResult) Error() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.errors == nil {
		return nil
	}

	if len(r.errors) == 1 {
		return r.errors[0]
	}

	errorStrings := make([]string, len(r.errors))
	for i := 0; i < len(r.errors); i++ {
		errorStrings[i] = fmt.Sprintf("Error %d: %s", i, r.errors[i].Error())
	}

	return fmt.Errorf(strings.Join(errorStrings, "; "))
}

func parseQuantiles(quantiles []*io_prometheus_client.Quantile) DurationPercentileMap {
	m := make(DurationPercentileMap)
	for _, quantile := range quantiles {
		m[*quantile.Quantile] = time.Duration(*quantile.Value)
	}
	return m
}

var (
	resultTemplate = template.Must(template.New("result parse").
			Funcs(templateFunctions).
			Parse(`
{{if .Failed -}}
{red}{bold}{u}Load Test Failed{-}
{{- else -}}
{green}{bold}{u}Load Test Passed{-}
{{- end}}
{{- if .Error}}
{red}Error: {{.Error}}{-}
{{- end}}
{{.IterationsStarted}} iterations started in {{duration .Duration}} ({{rate .Duration .IterationsStarted}}/second)
{{- if .SuccessfulIterationCount}}
{bold}Successful Iterations:{-} {green}{{.SuccessfulIterationCount}} ({{percent .SuccessfulIterationCount .Iterations | printf "%0.2f"}}%%, {{rate .Duration .SuccessfulIterationCount}}/second){-} {{.SuccessfulIterationDurations.String}}
{{- end}}
{{- if .FailedIterationCount}}
{bold}Failed Iterations:{-} {red}{{.FailedIterationCount}} ({{percent .FailedIterationCount .Iterations | printf "%0.2f"}}%%, {{rate .Duration .FailedIterationCount}}){-} {{.FailedIterationDurations.String}}
{{- end}}
{{- if .DroppedIterationCount}}
{bold}Dropped Iterations:{-} {yellow}{{.DroppedIterationCount}} ({{percent .DroppedIterationCount .Iterations | printf "%0.2f"}}%%, {{rate .Duration .DroppedIterationCount}}){-} (consider increasing --concurrency setting)
{{- end}}
{bold}Full logs:{-} {{.LogFile}}
`))
	setup = template.Must(template.New("setup").
		Funcs(templateFunctions).
		Parse(`{cyan}[Setup]{-}    {{if .Error}}{red}✘ {{.Error}}{-}{{else}}{green}✔{-}{{end}}`))

	progress = template.Must(template.New("result parse").
			Funcs(templateFunctions).
			Parse(`{cyan}[{{durationSeconds .Duration | printf "%5s"}}]{-}  {green}✔ {{printf "%5d" .SuccessfulIterationCount}}{-}  {{if .DroppedIterationCount}}{yellow}⦸ {{printf "%5d" .DroppedIterationCount}}{-}  {{end}}{red}✘ {{printf "%5d" .FailedIterationCount}}{-} {light_black}({{rate .RecentDuration .RecentSuccessfulIterations}}/s){-}
{{- with .SuccessfulIterationDurations}}   p(50): {{.Get 0.5}},  p(95): {{.Get 0.95}}, p(100): {{.Get 1.0}}{{end}}`))
	teardown = template.Must(template.New("teardown").
			Funcs(templateFunctions).
			Parse(`{cyan}[Teardown]{-} {{if .Error}}{red}✘ {{.Error}}{-}{{else}}{green}✔{-}{{end}}`))
	timeout = template.Must(template.New("timeout").
		Funcs(templateFunctions).
		Parse(`{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Max Duration Elapsed - waiting for active tests to complete{-}`))
	maxIterationsReached = template.Must(template.New("maxIterationsReached").
				Funcs(templateFunctions).
				Parse(`{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Max Iterations Reached - waiting for active tests to complete{-}`))
	interrupt = template.Must(template.New("interrupt").
			Funcs(templateFunctions).
			Parse(`{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Interrupted - waiting for active tests to complete{-}`))
)

func (r *RunResult) String() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(resultTemplate, r)
}

func (r *RunResult) Failed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Error() != nil || r.FailedIterationCount > 0 || (!r.IgnoreDropped && r.DroppedIterationCount > 0)
}

func (r *RunResult) Progress() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(progress, r)
}

func (r *RunResult) Duration() time.Duration {
	if r.StartTime().IsZero() {
		return 0
	}

	return time.Since(r.StartTime())
}

func (r *RunResult) Setup() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(setup, r)
}

func (r *RunResult) Teardown() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(teardown, r)
}

func (r *RunResult) MaxDurationElapsed() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(timeout, r)
}

func (r *RunResult) Interrupted() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(interrupt, r)
}

func (r *RunResult) Iterations() uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.FailedIterationCount + r.SuccessfulIterationCount + r.DroppedIterationCount
}

func (r *RunResult) IterationsStarted() uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.FailedIterationCount + r.SuccessfulIterationCount
}

func (r *RunResult) StartTime() time.Time {
	return r.startTime
}

func (r *RunResult) RecordStarted() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.startTime = time.Now()
}

func (r *RunResult) RecordTestFinished() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.TestDuration = time.Since(r.startTime)
}

func (r *RunResult) MaxIterationsReached() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return renderTemplate(maxIterationsReached, r)
}
