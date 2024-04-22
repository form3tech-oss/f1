package options

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/logging"
)

type RunOptions struct {
	RegisterLogHookFunc logging.RegisterLogHookFunc
	Scenario            string
	MaxDuration         time.Duration
	Concurrency         int
	MaxIterations       uint64
	MaxFailures         uint64
	MaxFailuresRate     int
	Verbose             bool
	VerboseFail         bool
	IgnoreDropped       bool
}
