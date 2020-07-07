package testing

import (
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/pkg/f1/metrics"

	"github.com/stretchr/testify/require"
)

type T struct {
	// Identifier of the user for the test
	VirtualUser string
	// Iteration number, "setup" or "teardown"
	Iteration string
	// Logger with user and iteration tags
	Log         *log.Logger
	failed      bool
	Require     *require.Assertions
	Environment map[string]string
	Scenario    string
}

func NewT(env map[string]string, vu, iter string, scenarioName string) *T {
	t := &T{
		VirtualUser: vu,
		Iteration:   iter,
		Log:         log.New().WithField("u", vu).WithField("i", iter).WithField("scenario", scenarioName).Logger,
		Environment: env,
		Scenario:    scenarioName,
	}
	t.Require = require.New(t)
	return t
}

func (t *T) Errorf(format string, args ...interface{}) {
	t.failed = true
	t.Log.Errorf(format, args...)
}

func (t *T) FailNow() {
	t.failed = true
	runtime.Goexit()
}

func (t *T) HasFailed() bool {
	return t.failed
}

// Time records a metric for the duration of the given function
func (t *T) Time(stageName string, f func()) {
	start := time.Now()
	defer recordTime(t, stageName, start)
	f()
}

func recordTime(t *T, stageName string, start time.Time) {
	metrics.Instance().Record(metrics.IterationResult, t.Scenario, stageName, metrics.Result(t.failed), time.Since(start).Nanoseconds())
}
