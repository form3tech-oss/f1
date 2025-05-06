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
	Trigger     WorkTriggerer
	Rate        RateFunction
	Description string
	Options     Options
	Duration    time.Duration
}

type Options struct {
	Scenario        string
	MaxDuration     time.Duration
	Concurrency     int
	MaxIterations   uint64
	MaxFailures     uint64
	MaxFailuresRate int
	Verbose         bool
	VerboseFail     bool
	IgnoreDropped   bool
}

type Rates struct {
	IterationRate     RateFunction
	IterationDuration time.Duration
	Duration          time.Duration
	Rate              RateFunction
}
