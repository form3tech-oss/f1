package testing

import (
	"runtime"
	"runtime/debug"
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
	Logger        *log.Logger
	failed        int64
	Require       *require.Assertions
	Scenario      string
	teardownStack []func()
}

func NewT(iter, scenarioName string) (*T, func()) {
	t := &T{
		Iteration:     iter,
		Logger:        log.WithField("i", iter).WithField("scenario", scenarioName).Logger,
		Scenario:      scenarioName,
		teardownStack: []func(){},
	}
	t.Require = require.New(t)
	return t, t.teardown
}

func (t *T) Name() string {
	return t.Scenario
}

func (t *T) FailNow() {
	atomic.StoreInt64(&t.failed, int64(1))
	runtime.Goexit()
}

func (t *T) Fail() {
	atomic.StoreInt64(&t.failed, int64(1))
}

func (t *T) Errorf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.Fail()
}

func (t *T) Error(err error) {
	t.Logf("%s failed due to %s", t.Iteration, err.Error())
	t.Fail()
}

func (t *T) Fatalf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.FailNow()
}

func (t *T) Fatal(err error) {
	t.Logf("%s failed due to %s", t.Iteration, err.Error())
	t.FailNow()
}

func (t *T) Log(args ...interface{}) {
	t.Logger.Error(args...)
}

func (t *T) Logf(format string, args ...interface{}) {
	t.Logger.Errorf(format, args...)
}

func (t *T) Failed() bool {
	return atomic.LoadInt64(&t.failed) == int64(1)
}

// Time records a metric for the duration of the given function
func (t *T) Time(stageName string, f func()) {
	start := time.Now()
	defer recordTime(t, stageName, start)
	f()
}

func recordTime(t *T, stageName string, start time.Time) {
	metrics.Instance().Record(metrics.IterationResult, t.Scenario, stageName, metrics.Result(t.Failed()), time.Since(start).Nanoseconds())
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

func checkResults(t *T, done chan<- struct{}) {
	r := recover()
	if r != nil {
		err, isError := r.(error)
		if isError {
			t.Error(err)
			debug.PrintStack()
		} else {
			t.Errorf("panic in %s: %v", t.Iteration, err)
		}
	}
	close(done)
}
