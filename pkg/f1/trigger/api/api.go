package api

import (
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/options"

	"github.com/spf13/pflag"
)

type WorkTriggerer func(doWork chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions)
type RateFunction func(time.Time) int

type Parameter struct {
	Name        string
	Short       string
	Description string
	Default     string
}

type Builder struct {
	Name        string
	Description string
	New         Constructor
	Flags       *pflag.FlagSet
}

type Constructor func(*pflag.FlagSet) (*Trigger, error)

type Trigger struct {
	Trigger     WorkTriggerer
	DryRun      RateFunction
	Description string
	Duration    time.Duration
}
