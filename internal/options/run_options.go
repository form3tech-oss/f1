package options

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/logging"
)

type RunOptions struct {
	Scenario            string
	MaxDuration         time.Duration
	Concurrency         int
	Verbose             bool
	VerboseFail         bool
	MaxIterations       int32
	RegisterLogHookFunc logging.RegisterLogHookFunc
	IgnoreDropped       bool
}
