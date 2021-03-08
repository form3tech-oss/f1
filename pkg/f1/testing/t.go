package testing

import (
	"runtime"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/internal/metrics"

	"github.com/stretchr/testify/require"
)

type T struct {
	// "iteration " + iteration number or "setup"
	Iteration string
	// Logger with user and iteration tags
	Log           *log.Logger
	failed        int64
	Require       *require.Assertions
	Scenario      string
	teardownStack []func()
}

func NewT(iter, scenarioName string) *T {
	t := &T{
		Iteration:     iter,
		Log:           log.WithField("i", iter).WithField("scenario", scenarioName).Logger,
		Scenario:      scenarioName,
		teardownStack: []func(){},
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
	t.Log.Errorf("%s failed and stopped", t.Iteration)
	runtime.Goexit()
}

func (t *T) Fail() {
	atomic.StoreInt64(&t.failed, int64(1))
	t.Log.Errorf("%s failed", t.Iteration)
}

func (t *T) FailWithError(err error) {
	atomic.StoreInt64(&t.failed, int64(1))
	t.Log.WithError(err).Errorf("%s failed due to %s", t.Iteration, err.Error())
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

// Cleanup registers a teardown function to be called when T has completed
func (t *T) Cleanup(f func()) {
	t.teardownStack = append(t.teardownStack, f)
}

func (t *T) teardown() {
	for i := len(t.teardownStack) - 1; i >= 0; i-- {
		done := make(chan struct{})
		go func() {
			defer checkResults(t, done)
			t.teardownStack[i]()
		}()
		<-done
	}
}
