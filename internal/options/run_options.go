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
	MaxFailures         int
	MaxFailuresRate     int
	MaxIterations       int32
	Verbose             bool
	VerboseFail         bool
	IgnoreDropped       bool
}
