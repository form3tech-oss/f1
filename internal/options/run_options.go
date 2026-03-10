package options

import (
	"time"
)

type RunOptions struct {
	Scenario                 string
	MaxDuration              time.Duration
	Concurrency              int
	MaxIterations            uint64
	MaxFailures              uint64
	MaxFailuresRate          int
	Verbose                  bool
	IgnoreDropped            bool
	WaitForCompletionTimeout time.Duration
}

type RunOption func(*RunOptions)

func WithScenario(s string) RunOption {
	return func(o *RunOptions) { o.Scenario = s }
}

func WithMaxDuration(d time.Duration) RunOption {
	return func(o *RunOptions) { o.MaxDuration = d }
}

func WithConcurrency(n int) RunOption {
	return func(o *RunOptions) { o.Concurrency = n }
}

func WithMaxIterations(n uint64) RunOption {
	return func(o *RunOptions) { o.MaxIterations = n }
}

func WithMaxFailures(n uint64) RunOption {
	return func(o *RunOptions) { o.MaxFailures = n }
}

func WithMaxFailuresRate(n int) RunOption {
	return func(o *RunOptions) { o.MaxFailuresRate = n }
}

func WithVerbose(v bool) RunOption {
	return func(o *RunOptions) { o.Verbose = v }
}

func WithIgnoreDropped(v bool) RunOption {
	return func(o *RunOptions) { o.IgnoreDropped = v }
}

func WithWaitForCompletionTimeout(d time.Duration) RunOption {
	return func(o *RunOptions) { o.WaitForCompletionTimeout = d }
}

func DefaultRunOptions() RunOptions {
	return RunOptions{
		MaxDuration:              time.Second,
		Concurrency:              100,
		WaitForCompletionTimeout: 10 * time.Second,
	}
}

func (o *RunOptions) LogToFile() bool {
	return !o.Verbose
}
