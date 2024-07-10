package options

import (
	"time"
)

type RunOptions struct {
	Scenario        string
	MaxDuration     time.Duration
	Concurrency     int
	MaxIterations   uint64
	MaxFailures     uint64
	MaxFailuresRate int
	Verbose         bool
	IgnoreDropped   bool
}

func (o *RunOptions) LogToFile() bool {
	return !o.Verbose
}
