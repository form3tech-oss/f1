package testing

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/internal/metrics"

	"github.com/stretchr/testify/require"
)

// T is a type passed to Scenario functions to manage test state and support formatted test logs.
// A test ends when its Scenario function returns or calls any of the methods FailNow, Fatal, Fatalf.
// Those methods must be called only from the goroutine running the Scenario function.
// The other reporting methods, such as the variations of Log and Error, may be called simultaneously from multiple goroutines.
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

// Name returns the name of the running Scenario.
func (t *T) Name() string {
	return t.Scenario
}

// FailNow marks the function as having failed and stops its execution by calling runtime.Goexit (which then runs all deferred calls in the current goroutine).
// Execution will continue at the next Scenario iteration. FailNow must be called from the goroutine running the Scenario,
// not from other goroutines created during the Scenario. Calling FailNow does not stop those other goroutines.
func (t *T) FailNow() {
	atomic.StoreInt64(&t.failed, int64(1))
	runtime.Goexit()
}

// Fail marks the function as having failed but continues execution.
func (t *T) Fail() {
	atomic.StoreInt64(&t.failed, int64(1))
}

// Errorf is equivalent to Logf followed by Fail.
func (t *T) Errorf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.Fail()
}

// Error is equivalent to Log followed by Fail.
func (t *T) Error(err error) {
	t.Logf("%s failed due to %s", t.Iteration, err.Error())
	t.Fail()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (t *T) Fatalf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.FailNow()
}

// Fatal is equivalent to Log followed by FailNow.
func (t *T) Fatal(err error) {
	t.Logf("%s failed due to %s", t.Iteration, err.Error())
	t.FailNow()
}

// Log formats its arguments using default formatting, analogous to Println, and records the text in the error log.
// The text will be printed only if f1 is running in verbose mode.
func (t *T) Log(args ...interface{}) {
	t.Logger.Error(args...)
}

// Logf formats its arguments according to the format, analogous to Printf, and records the text in the error log.
// A final newline is added if not provided. The text will be printed only if f1 is running in verbose mode.
func (t *T) Logf(format string, args ...interface{}) {
	t.Logger.Errorf(format, args...)
}

// Failed reports whether the function has failed.
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

// Cleanup registers a function to be called when the scenario or the iteration completes.
// Cleanup functions will be called in last added, first called order.
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
