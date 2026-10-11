package api

import (
	"context"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

type (
	WorkTriggerer func(ctx context.Context, output *ui.Output, workers *workers.PoolManager, options options.RunOptions)
	RateFunction  func(time.Time) int
)

type Parameter struct {
	Name        string
	Short       string
	Description string
	Default     string
}

type Builder struct {
	New               Constructor
	Flags             *pflag.FlagSet
	Name              string
	Description       string
	IgnoreCommonFlags bool
}

type Constructor func(*pflag.FlagSet) (*Trigger, error)

type Trigger struct {
	Trigger WorkTriggerer
	DryRun  RateFunction
	// Cleanup, if set, is called once the run has stopped waiting for in-flight
	// iterations, before scenario teardown.
	Cleanup     func()
	Description string
	Options     Options
	// Start, if set, is when the trigger's schedule starts. Otherwise it starts with the run.
	Start time.Time
	// Duration, if above 0, is how long the trigger's schedule lasts from its start.
	Duration time.Duration
}

// End returns when the trigger's schedule ends for a run that starts at runStart, or the
// zero time if the schedule has no end.
func (t *Trigger) End(runStart time.Time) time.Time {
	if t.Duration <= 0 {
		return time.Time{}
	}
	if t.Start.IsZero() {
		return runStart.Add(t.Duration)
	}
	return t.Start.Add(t.Duration)
}

// Deadline returns when a run that starts at runStart ends: when the trigger's schedule
// does, or after maxDuration, whichever comes first. maxDuration counts from the
// schedule's start when that is still to come, so waiting for it doesn't use up the run.
func (t *Trigger) Deadline(runStart time.Time, maxDuration time.Duration) time.Time {
	deadline := runStart.Add(maxDuration)
	if t.Start.After(runStart) {
		deadline = t.Start.Add(maxDuration)
	}
	if end := t.End(runStart); !end.IsZero() && end.Before(deadline) {
		return end
	}
	return deadline
}

type Options struct {
	Scenario                 string
	MaxDuration              time.Duration
	Concurrency              int
	MaxIterations            uint64
	MaxFailures              uint64
	MaxFailuresRate          int
	Verbose                  bool
	VerboseFail              bool
	IgnoreDropped            bool
	WaitForCompletionTimeout time.Duration
}

type Rates struct {
	Rate              RateFunction
	IterationDuration time.Duration
	Duration          time.Duration
}
