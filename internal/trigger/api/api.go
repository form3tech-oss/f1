package api

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trace"
)

type (
	WorkTriggerer func(doWork chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions)
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

type Constructor func(*pflag.FlagSet, trace.Tracer) (*Trigger, error)

type Trigger struct {
	Trigger     WorkTriggerer
	DryRun      RateFunction
	Description string
	Options     Options
	Duration    time.Duration
}

type Options struct {
	Scenario        string
	MaxDuration     time.Duration
	Concurrency     int
	MaxFailures     int
	MaxFailuresRate int
	MaxIterations   int32
	Verbose         bool
	VerboseFail     bool
	IgnoreDropped   bool
}

type Rates struct {
	Rate              RateFunction
	IterationDuration time.Duration
	Duration          time.Duration
}
