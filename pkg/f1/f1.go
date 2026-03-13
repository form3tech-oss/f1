package f1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/form3tech-oss/f1/v3/internal/envsettings"
	"github.com/form3tech-oss/f1/v3/internal/ui"
	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
	"github.com/form3tech-oss/f1/v3/pkg/f1/scenarios"
)

const (
	// signalChanBufferSize of 2 so that we force exit on a consecutive SIGTERM or SIGINT.
	//
	// Package signal will not block sending to c: the caller must ensure that c has sufficient buffer
	// space to keep up with the expected signal rate.
	// For a channel used for notification of just one signal value, a buffer of size 1 is sufficient.
	signalChanBufferSize = 2
)

// F1 represents an F1 CLI instance. Instantiate this struct to create an instance
// of the F1 CLI and to register new test scenarios.
type F1 struct {
	scenarios *scenarios.Scenarios
	profiling *profiling
	settings  envsettings.Settings
	options   *f1Options
}

type f1Options struct {
	output         *ui.Output
	staticMetrics  map[string]string
	loggerExplicit bool
}

// Option configures an F1 instance at construction.
type Option func(*F1)

// WithLogger specifies the logger for internal and scenario logs.
// When used, WithLogLevel, WithLogFormat, F1_LOG_LEVEL and F1_LOG_FORMAT
// have no effect because the caller controls the logger directly.
func WithLogger(logger *slog.Logger) Option {
	return func(f *F1) {
		f.options.output = ui.NewDefaultOutputWithLogger(logger)
		f.options.loggerExplicit = true
	}
}

// WithStaticMetrics registers additional labels with fixed values for f1 metrics.
func WithStaticMetrics(labels map[string]string) Option {
	return func(f *F1) {
		f.options.staticMetrics = labels
	}
}

// New instantiates a new F1 CLI. Pass options to configure logger, metrics, etc.
//
// Construction order:
//  1. Load settings from environment variables
//  2. Apply options (may override individual settings or clear them via WithoutEnvSettings)
//  3. Build default output from final settings unless WithLogger was used
func New(opts ...Option) *F1 {
	f := &F1{
		scenarios: scenarios.New(),
		profiling: &profiling{},
		settings:  envsettings.Get(),
		options:   &f1Options{},
	}
	for _, opt := range opts {
		opt(f)
	}

	if !f.options.loggerExplicit {
		f.options.output = ui.NewDefaultOutput(f.settings.Log.SlogLevel(), f.settings.Log.IsFormatJSON())
	}

	return f
}

// AddScenario registers a new test scenario with the given name. This is the name used when running
// load test scenarios. For example, calling the function with the following arguments:
//
//	f.AddScenario("myTest", myScenario)
//
// will result in the test "myTest" being runnable from the command line:
//
//	f1 run constant -r 1/s -d 10s myTest
func (f *F1) AddScenario(name string, scenarioFn f1testing.ScenarioFn, options ...scenarios.ScenarioOption) *F1 {
	info := &scenarios.Scenario{
		Name:       name,
		ScenarioFn: scenarioFn,
	}

	for _, opt := range options {
		opt(info)
	}

	f.scenarios.AddScenario(info)
	return f
}

// newSignalContext returns a context that is cancelled when parent is cancelled or
// when SIGINT/SIGTERM is received. If a signal is received a second time, the app exits.
func newSignalContext(parent context.Context, stopCh <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(parent)

	c := make(chan os.Signal, signalChanBufferSize)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-c:
			cancel()
		case <-parent.Done():
			return
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

// Run runs the CLI with the given args. Returns error on failure; never exits.
// ctx controls cancellation; SIGINT/SIGTERM also cancel via internal signal handling.
// Pass nil for args to use os.Args (e.g. when called from main).
func (f *F1) Run(ctx context.Context, args []string) error {
	if err := f.execute(ctx, args); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

// Execute runs the CLI and exits with code 1 on error. Convenience for main().
func (f *F1) Execute() {
	if err := f.Run(context.Background(), nil); err != nil {
		f.options.output.Display(ui.ErrorMessage{Message: "f1 failed", Error: err})
		os.Exit(1)
	}
}

// GetScenarios returns the list of registered scenarios.
func (f *F1) GetScenarios() *scenarios.Scenarios {
	return f.scenarios
}

func (f *F1) execute(ctx context.Context, args []string) error {
	stopCh := make(chan struct{})
	defer close(stopCh)
	execCtx := newSignalContext(ctx, stopCh)

	rootCmd, err := buildRootCmd(execCtx, f.scenarios, f.settings, f.profiling, f.options.output, f.options.staticMetrics)
	if err != nil {
		return fmt.Errorf("building root command: %w", err)
	}

	if len(args) > 0 {
		rootCmd.SetArgs(args)
	}

	err = rootCmd.ExecuteContext(execCtx)
	// stop profiling regardless of err
	profilingErr := f.profiling.stop()

	errs := errors.Join(err, profilingErr)

	if errs != nil {
		return fmt.Errorf("command execution: %w", errs)
	}

	return nil
}
