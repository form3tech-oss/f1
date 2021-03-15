package f1

import (
	"fmt"
	"os"

	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type F1 struct {
	scenarios *scenarios.Scenarios
	profiling *profiling
}

type profiling struct {
	cpuProfile *os.File
	memProfile string
}

func New() *F1 {
	return &F1{
		scenarios: scenarios.New(),
		profiling: &profiling{},
	}
}

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

func (f *F1) GetScenarios() *scenarios.Scenarios {
	return f.scenarios
}
