package f1

import (
	"fmt"
	"os"

	"github.com/form3tech-oss/f1/pkg/f1/testing"
)

type F1 struct {
	scenarios *testing.Scenarios
	profiling *profiling
}

type profiling struct {
	cpuProfile *os.File
	memProfile string
}

func New() *F1 {
	return &F1{
		scenarios: testing.New(),
		profiling: &profiling{},
	}
}

func (f *F1) WithScenario(name string, setupFn testing.SetupFn) *F1 {
	f.scenarios.AddByName(name, setupFn)
	return f
}

func (f *F1) WithScenarioDescription(info testing.ScenarioInfo, setupFn testing.SetupFn) *F1 {
	f.scenarios.Add(info, setupFn)
	return f
}

func (f *F1) WithMultiStageScenario(name string, setupFn testing.MultiStageSetupFn) *F1 {
	f.scenarios.AddMultiStage(name, setupFn)
	return f
}

func (f *F1) Execute() {
	if err := buildRootCmd(f.scenarios, f.profiling).Execute(); err != nil {
		writeProfiles(f.profiling)
		fmt.Println(err)
		os.Exit(1)
	}
}

func (f *F1) ExecuteWithArgs(args []string) error {
	rootCmd := buildRootCmd(f.scenarios, f.profiling)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	writeProfiles(f.profiling)
	return err
}
