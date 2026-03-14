package f1testing

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"sync/atomic"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v3/internal/log"
)

var errFailNow = errors.New("FailNow")

// IterationSetup is the value of T.Iteration during the setup phase.
// Iteration numbers from the run phase are 1-based, so 0 is never used for iterations.
const IterationSetup uint64 = 0

// T is a type passed to Scenario functions to manage test state and support formatted test logs. A
// test ends when its Scenario function returns or calls any of the methods FailNow, Fatal, Fatalf.
// Those methods must be called only from the goroutine running the Scenario function. The other
// reporting methods, such as the variations of Log and Error, may be called simultaneously from
// multiple goroutines.
type T struct {
	logger  *slog.Logger
	require *require.Assertions
	// Iteration is the iteration index (1-based) or IterationSetup (0) for the setup phase.
	Iteration uint64
	Scenario  string
	// VUID is the Virtual User ID - a stable identifier for the pool worker running this iteration.
	// Useful for correlating iterations with user-specific test data (e.g. in the "users" trigger mode).
	// VUID is -1 for setup; 0-based for pool workers.
	VUID           int
	teardownStack  []func()
	failed         atomic.Bool
	teardownFailed atomic.Bool
	tearingDown    bool
}

type TOption func(*T)

func WithLogger(logger *slog.Logger) TOption {
	return func(t *T) {
		t.logger = logger
	}
}

func WithIteration(iteration uint64) TOption {
	return func(t *T) {
		t.Iteration = iteration
	}
}

// WithVUID sets the Virtual User ID for the test context.
// Use -1 for setup phase; 0-based integers for pool workers.
func WithVUID(id int) TOption {
	return func(t *T) {
		t.VUID = id
	}
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
	if t.logger == nil {
		t.logger = slog.Default()
	}

	return t, t.teardown
}

func (t *T) Reset(iter uint64) {
	t.Iteration = iter
	t.failed.Store(false)
	t.teardownFailed.Store(false)
	t.tearingDown = false
	t.teardownStack = []func(){}
}

func (t *T) Logger() *slog.Logger {
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

// Errorf is equivalent to Logf followed by Fail. Logs at Error level.
func (t *T) Errorf(format string, args ...any) {
	t.logger.With(log.IterationAttr(t.Iteration), log.VUIDAttr(t.VUID)).Error(fmt.Sprintf(format, args...))
	t.Fail()
}

// Error is equivalent to Log followed by Fail. Logs at Error level.
func (t *T) Error(args ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(args...), "\n")
	t.logger.With(log.IterationAttr(t.Iteration), log.VUIDAttr(t.VUID)).Error(msg)
	t.Fail()
}

// Fatalf is equivalent to Logf followed by FailNow. Logs at Error level.
func (t *T) Fatalf(format string, args ...any) {
	t.logger.With(log.IterationAttr(t.Iteration), log.VUIDAttr(t.VUID)).Error(fmt.Sprintf(format, args...))
	t.FailNow()
}

// Fatal is equivalent to Log followed by FailNow. Logs at Error level.
func (t *T) Fatal(args ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(args...), "\n")
	t.logger.With(log.IterationAttr(t.Iteration), log.VUIDAttr(t.VUID)).Error(msg)
	t.FailNow()
}

// Log formats its arguments using default formatting, analogous to Println, and records the text in the error log.
// The text will be printed only if f1 is running in verbose mode.
// Aligns with testing.T: uses fmt.Sprintln for space-separated args (trailing newline trimmed for structured logs).
func (t *T) Log(args ...any) {
	msg := strings.TrimSuffix(fmt.Sprintln(args...), "\n")
	t.logger.With(log.IterationAttr(t.Iteration), log.VUIDAttr(t.VUID)).Info(msg)
}

// Logf formats its arguments according to the format, analogous to Printf, and records the text in the error log.
// A final newline is added if not provided. The text will be printed only if f1 is running in verbose mode.
// Aligns with testing.T: uses fmt.Sprintf.
func (t *T) Logf(format string, args ...any) {
	t.logger.With(log.IterationAttr(t.Iteration), log.VUIDAttr(t.VUID)).Info(fmt.Sprintf(format, args...))
}

// Failed reports whether the function has failed.
func (t *T) Failed() bool {
	return t.failed.Load()
}

func (t *T) TeardownFailed() bool {
	return t.teardownFailed.Load()
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
			log.VUIDAttr(t.VUID),
			log.ErrorAttr(err),
		)
		t.Fail()
	default:
		stack := debug.Stack()
		t.logger.Error("recovered panic in scenario",
			log.StackTraceAttr(stack),
			log.IterationAttr(t.Iteration),
			log.VUIDAttr(t.VUID),
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
