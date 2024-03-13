package f1

import (
	"errors"
	"fmt"
	"os"

	"github.com/form3tech-oss/f1/v2/internal/envsettings"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

// Represents an F1 CLI instance. Instantiate this struct to create an instance
// of the F1 CLI and to register new test scenarios.
type F1 struct {
	scenarios *scenarios.Scenarios
	profiling *profiling
	settings  envsettings.Settings
}

// Instantiates a new instance of an F1 CLI.
func New() *F1 {
	return &F1{
		scenarios: scenarios.New(),
		profiling: &profiling{},
		settings:  envsettings.Get(),
	}
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

func (f *F1) execute(args []string) error {
	rootCmd, err := buildRootCmd(f.scenarios, f.settings, f.profiling)
	if err != nil {
		return fmt.Errorf("building root command: %w", err)
	}

	if len(args) > 0 {
		rootCmd.SetArgs(args)
	}

	err = rootCmd.Execute()
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
		fmt.Println(err)
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

// CombineScenarios creates a single scenario that will call each ScenarioFn
// sequentially and return a testing.RunFn that will call each scenario's RunFn
// every iteration.
func CombineScenarios(scenarios ...testing.ScenarioFn) testing.ScenarioFn {
	return func(t *testing.T) testing.RunFn {
		var run []testing.RunFn
		for _, s := range scenarios {
			run = append(run, s(t))
		}

		return func(t *testing.T) {
			for _, r := range run {
				r(t)
			}
		}
	}
}
