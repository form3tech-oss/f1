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
	Name              string
	Description       string
	New               Constructor
	Flags             *pflag.FlagSet
	IgnoreCommonFlags bool
}

type Constructor func(*pflag.FlagSet, trace.Tracer) (*Trigger, error)

type Trigger struct {
	Trigger     WorkTriggerer
	DryRun      RateFunction
	Description string
	Duration    time.Duration
	Options     Options
}

type Options struct {
	MaxDuration     time.Duration
	Concurrency     int
	Verbose         bool
	VerboseFail     bool
	MaxIterations   int32
	MaxFailures     int
	MaxFailuresRate int
	IgnoreDropped   bool
	Scenario        string
}

type Rates struct {
	IterationDuration time.Duration
	Rate              RateFunction
	Duration          time.Duration
}
