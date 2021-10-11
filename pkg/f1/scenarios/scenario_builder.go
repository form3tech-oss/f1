package scenarios

import (
	"sort"

	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

// Represents a list of test scenarios.
type Scenarios struct {
	scenarios map[string]*Scenario
}

// Represents a test scenario.
type Scenario struct {
	// The name of the test scenario, which can be used to invoke it via the CLI.
	Name        string
	Description string
	Parameters  []ScenarioParameter
	ScenarioFn  testing.ScenarioFn
	// The function that is invoked on each iteration of the test scenario.
	RunFn testing.RunFn
}

type ScenarioParameter struct {
	Name        string
	Description string
	Default     string
}

type ScenarioOption func(info *Scenario)

func Description(d string) ScenarioOption {
	return func(i *Scenario) {
		i.Description = d
	}
}

func Parameter(parameter ScenarioParameter) ScenarioOption {
	return func(i *Scenario) {
		i.Parameters = append(i.Parameters, parameter)
	}
}

func New() *Scenarios {
	return &Scenarios{
		scenarios: make(map[string]*Scenario),
	}
}

func (s *Scenarios) Add(scenario *Scenario) *Scenarios {
	s.scenarios[scenario.Name] = scenario
	return s
}

func (s *Scenarios) GetScenario(scenarioName string) *Scenario {
	return s.scenarios[scenarioName]
}

func (s *Scenarios) GetScenarioNames() []string {
	names := make([]string, len(s.scenarios))
	index := 0
	for key := range s.scenarios {
		names[index] = key
		index++
	}
	sort.Strings(names)
	return names
}
