package testing

import (
	"runtime"
	"sync/atomic"
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
	failed      int64
	Require     *require.Assertions
	Environment map[string]string
	Scenario    string
}

func NewT(env map[string]string, vu, iter string, scenarioName string) *T {
	t := &T{
		VirtualUser: vu,
		Iteration:   iter,
		Log:         log.WithField("u", vu).WithField("i", iter).WithField("scenario", scenarioName).Logger,
		Environment: env,
		Scenario:    scenarioName,
	}
	t.Require = require.New(t)
	return t
}

func (t *T) Errorf(format string, args ...interface{}) {
	atomic.StoreInt64(&t.failed, int64(1))
	t.Log.Errorf(format, args...)
}

func (t *T) FailNow() {
	atomic.StoreInt64(&t.failed, int64(1))
	t.Log.Errorf("test failed and stopped")
	runtime.Goexit()
}

func (t *T) Fail() {
	atomic.StoreInt64(&t.failed, int64(1))
	t.Log.Errorf("test failed")
}

func (t *T) FailWithError(err error) {
	atomic.StoreInt64(&t.failed, int64(1))
	t.Log.WithError(err).Errorf("test failed due to %s", err.Error())
}

func (t *T) HasFailed() bool {
	return atomic.LoadInt64(&t.failed) == int64(1)
}

// Time records a metric for the duration of the given function
func (t *T) Time(stageName string, f func()) {
	start := time.Now()
	defer recordTime(t, stageName, start)
	f()
}

func recordTime(t *T, stageName string, start time.Time) {
	metrics.Instance().Record(metrics.IterationResult, t.Scenario, stageName, metrics.Result(t.HasFailed()), time.Since(start).Nanoseconds())
}
