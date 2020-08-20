package options

import (
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/logging"
)

type RunOptions struct {
	RunName             string
	ScenarioName        string
	MaxDuration         time.Duration
	Concurrency         int
	Env                 map[string]string
	Verbose             bool
	VerboseFail         bool
	MaxIterations       int32
	RegisterLogHookFunc logging.RegisterLogHookFunc
	IgnoreDropped       bool
}
