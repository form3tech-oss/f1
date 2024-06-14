package testing

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/metrics"
)

var errFailNow = errors.New("FailNow")

// T is a type passed to Scenario functions to manage test state and support formatted test logs. A
// test ends when its Scenario function returns or calls any of the methods FailNow, Fatal, Fatalf.
// Those methods must be called only from the goroutine running the Scenario function. The other
// reporting methods, such as the variations of Log and Error, may be called simultaneously from
// multiple goroutines.
type T struct {
	logrusLogger   *logrus.Logger
	logger         *slog.Logger
	require        *require.Assertions
	Iteration      string // iteration number or "setup"
	Scenario       string
	teardownStack  []func()
	failed         atomic.Bool
	teardownFailed atomic.Bool
	tearingDown    bool
}

type TOption func(*T)

// WithLogrusLogger will be removed in future versions, needed for backwards compatibility
//
// Deprecated: Will be removed in future versions.
func WithLogrusLogger(logrusLogger *logrus.Logger) TOption {
	return func(t *T) {
		t.logrusLogger = logrusLogger
	}
}

func WithLogger(logger *slog.Logger) TOption {
	return func(t *T) {
		t.logger = logger
	}
}

func WithIteration(iteration string) TOption {
	return func(t *T) {
		t.Iteration = iteration
	}
}

// NewT returns a new T state
//
// Deprecated: Will be removed in favour of NewTWithOptions
func NewT(iter, scenarioName string) (*T, func()) {
	logger := slog.Default()

	t, teardown := NewTWithOptions(scenarioName,
		WithIteration(iter),
		WithLogrusLogger(log.NewSlogLogrusLogger(logger)),
		WithLogger(logger),
	)

	return t, teardown
}

func NewTWithOptions(scenarioName string, options ...TOption) (*T, func()) {
	t := &T{
		Scenario:      scenarioName,
		teardownStack: []func(){},
	}
	t.require = require.New(t)

	for _, opt := range options {
		opt(t)
	}

	return t, t.teardown
}

func (t *T) Reset(iter string) {
	t.Iteration = iter
	t.failed.Store(false)
	t.teardownFailed.Store(false)
	t.tearingDown = false
	t.teardownStack = []func(){}
}

// Logger returns a logrus logger, needed for backwards compatibility. Use StandardLogger
// instead.
//
// Internally it uses slog as a logging backend.
//
// Deprecated: logrus will be removed in future versions.
func (t *T) Logger() *logrus.Logger {
	return t.logrusLogger
}

func (t *T) StandardLogger() *slog.Logger {
	return t.logger
}

func (t *T) Require() *require.Assertions {
	return t.require
}

// Name returns the name of the running Scenario.
func (t *T) Name() string {
	return t.Scenario
}

// FailNow marks the function as having failed and stops its execution.
// Execution will continue at the next Scenario iteration. FailNow must be called from
// the goroutine running the Scenario, not from other goroutines created during the Scenario.
// Calling FailNow does not stop those other goroutines.
func (t *T) FailNow() {
	if t.tearingDown {
		t.teardownFailed.Store(true)
	} else {
		t.failed.Store(true)
	}

	panic(errFailNow)
}

// Fail marks the function as having failed but continues execution.
func (t *T) Fail() {
	if t.tearingDown {
		t.teardownFailed.Store(true)
	} else {
		t.failed.Store(true)
	}
}

// Errorf is equivalent to Logf followed by Fail.
func (t *T) Errorf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.Fail()
}

// Error is equivalent to Log followed by Fail.
func (t *T) Error(err error) {
	t.Logf("%s failed due to: %s", t.Iteration, err.Error())
	t.Fail()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (t *T) Fatalf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.FailNow()
}

// Fatal is equivalent to Log followed by FailNow.
func (t *T) Fatal(err error) {
	t.Logf("%s failed due to: %s", t.Iteration, err.Error())
	t.FailNow()
}

// Log formats its arguments using default formatting, analogous to Println, and records the text in the error log.
// The text will be printed only if f1 is running in verbose mode.
func (t *T) Log(args ...any) {
	t.logger.Error(fmt.Sprint(args...))
}

// Logf formats its arguments according to the format, analogous to Printf, and records the text in the error log.
// A final newline is added if not provided. The text will be printed only if f1 is running in verbose mode.
func (t *T) Logf(format string, args ...any) {
	t.logger.Error(fmt.Sprintf(format, args...))
}

// Failed reports whether the function has failed.
func (t *T) Failed() bool {
	return t.failed.Load()
}

func (t *T) TeardownFailed() bool {
	return t.teardownFailed.Load()
}

// Time records a metric for the duration of the given function
func (t *T) Time(stageName string, f func()) {
	start := time.Now()
	defer recordTime(t, stageName, start)
	f()
}

// Cleanup registers a function to be called when the scenario or the iteration completes.
// Cleanup functions will be called in last added, first called order.
func (t *T) Cleanup(f func()) {
	t.teardownStack = append(t.teardownStack, f)
}

func CheckResults(t *T, done chan<- struct{}) {
	handlePanic(t, recover())

	if done != nil {
		done <- struct{}{}
	}
}

func handlePanic(t *T, recovered any) {
	if recovered == nil {
		return
	}

	err, isError := recovered.(error)
	switch {
	case isError && errors.Is(err, errFailNow):
		return
	case isError:
		stack := debug.Stack()
		t.logger.Error("recovered panic in scenario",
			log.StackTraceAttr(stack),
			log.IterationAttr(t.Iteration),
			log.ErrorAttr(err),
		)
		t.Fail()
	default:
		stack := debug.Stack()
		t.logger.Error("recovered panic in scenario",
			log.StackTraceAttr(stack),
			log.IterationAttr(t.Iteration),
			log.ErrorAnyAttr(recovered),
		)
		t.Fail()
	}
}

func (t *T) teardown() {
	t.tearingDown = true

	for i := len(t.teardownStack) - 1; i >= 0; i-- {
		func() {
			defer CheckResults(t, nil)
			t.teardownStack[i]()
		}()
	}
}

func recordTime(t *T, stageName string, start time.Time) {
	metrics.Instance().RecordIterationStage(
		t.Scenario,
		stageName,
		metrics.Result(t.Failed()),
		time.Since(start).Nanoseconds(),
	)
}
