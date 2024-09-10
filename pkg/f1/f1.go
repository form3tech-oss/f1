package f1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

const (
	// signalChanBufferSize of 2 so that we force exit on a consecutive SIGTERM or SIGINT.
	//
	// Package signal will not block sending to c: the caller must ensure that c has sufficient buffer
	// space to keep up with the expected signal rate.
	// For a channel used for notification of just one signal value, a buffer of size 1 is sufficient.
	signalChanBufferSize = 2
)

// Represents an F1 CLI instance. Instantiate this struct to create an instance
// of the F1 CLI and to register new test scenarios.
type F1 struct {
	output    *ui.Output
	scenarios *scenarios.Scenarios
	profiling *profiling
	settings  envsettings.Settings
}

// New instantiates a new instance of an F1 CLI.
func New() *F1 {
	settings := envsettings.Get()

	return &F1{
		scenarios: scenarios.New(),
		profiling: &profiling{},
		settings:  settings,
		output:    ui.NewDefaultOutput(settings.Log.SlogLevel(), settings.Log.IsFormatJSON()),
	}
}

// WithLogger allows specifying logger to be used for all internal and scenario logs
//
// This will disable the F1_LOG_LEVEL and F1_LOG_FORMAT options, as they only relate to the built-in
// logger.
//
// The logger will be used for non-interactive output, file logs or when `--verbose` is specified.
func (f *F1) WithLogger(logger *slog.Logger) *F1 {
	f.output = ui.NewDefaultOutputWithLogger(logger)
	return f
}

// Registers a new test scenario with the given name. This is the name used when running
// load test scenarios. For example, calling the function with the following arguments:
//
//	f.Add("myTest", myScenario)
//
// will result in the test "myTest" being runnable from the command line:
//
//	f1 run constant -r 1/s -d 10s myTest
func (f *F1) Add(name string, scenarioFn testing.ScenarioFn, options ...scenarios.ScenarioOption) *F1 {
	info := &scenarios.Scenario{
		Name:       name,
		ScenarioFn: scenarioFn,
	}

	for _, opt := range options {
		opt(info)
	}

	f.scenarios.Add(info)
	return f
}

// NewSignalContext returns a context.Context that is cancelled whenever
// 'SIGINT' or 'SIGTERM' are received.
// If one of these two signals is received a second time, the application exits.
func newSignalContext(stopCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, signalChanBufferSize)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-c:
			cancel()
		case <-stopCh:
			return
		}

		select {
		case <-c:
			signal.Reset()
			os.Exit(1)
		case <-stopCh:
			return
		}
	}()
	return ctx
}

func (f *F1) execute(args []string) error {
	rootCmd, err := buildRootCmd(f.scenarios, f.settings, f.profiling, f.output)
	if err != nil {
		return fmt.Errorf("building root command: %w", err)
	}

	if len(args) > 0 {
		rootCmd.SetArgs(args)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	ctx := newSignalContext(stopCh)

	err = rootCmd.ExecuteContext(ctx)
	// stop profiling regardless of err
	profilingErr := f.profiling.stop()

	errs := errors.Join(err, profilingErr)

	if errs != nil {
		return fmt.Errorf("command execution: %w", err)
	}

	return nil
}

// Synchronously runs the F1 CLI. This function is the blocking entrypoint to the CLI,
// so you should register your test scenarios with the Add function prior to calling this
// function.
func (f *F1) Execute() {
	if err := f.execute(nil); err != nil {
		f.output.Display(ui.ErrorMessage{Message: "f1 failed", Error: err})
		os.Exit(1)
	}
}

// Similar to Execute, but takes command line arguments from the args array. Useful
// for testing F1 test scenarios.
func (f *F1) ExecuteWithArgs(args []string) error {
	if err := f.execute(args); err != nil {
		return fmt.Errorf("execute with args: %w", err)
	}

	return nil
}

// Returns the list of registered scenarios.
func (f *F1) GetScenarios() *scenarios.Scenarios {
	return f.scenarios
}
