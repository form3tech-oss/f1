package scenarios

import (
	"sort"

	"github.com/form3tech-oss/f1/pkg/f1/testing"
	log "github.com/sirupsen/logrus"
)

type Scenarios struct {
	scenarios            map[string]testing.MultiStageSetupFn
	scenarioDescriptions []ScenarioInfo
}

type ScenarioInfo struct {
	Name        string
	Description string
	Parameters  []ScenarioParameter
}

type ScenarioParameter struct {
	Name        string
	Description string
	Default     string
}

type ScenarioOption func(info *ScenarioInfo)

func Description(d string) ScenarioOption {
	return func(i *ScenarioInfo) {
		i.Description = d
	}
}

func Parameter(parameter ScenarioParameter) ScenarioOption {
	return func(i *ScenarioInfo) {
		i.Parameters = append(i.Parameters, parameter)
	}
}
func New() *Scenarios {
	return &Scenarios{
		scenarios: make(map[string]testing.MultiStageSetupFn),
	}
}

func (s *Scenarios) Add(scenario ScenarioInfo, testSetup testing.ScenarioFn) *Scenarios {
	s.scenarioDescriptions = append(s.scenarioDescriptions, scenario)

	multiStageSetup := func(t *testing.T) []testing.Stage {
		run := testSetup(t)
		return []testing.Stage{{Name: "single", RunFn: run}}
	}

	return s.AddMultiStage(scenario.Name, multiStageSetup)
}

func (s *Scenarios) AddByName(name string, testSetup testing.ScenarioFn) *Scenarios {
	return s.Add(ScenarioInfo{Name: name}, testSetup)
}

func (s *Scenarios) AddMultiStage(name string, testSetup testing.MultiStageSetupFn) *Scenarios {
	log.Debugf("Registering test %s\n", name)
	s.scenarios[name] = testSetup
	return s
}

func (s *Scenarios) GetScenario(scenarioName string) testing.MultiStageSetupFn {
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

func WithNoSetup(fn func(t *testing.T)) testing.ScenarioFn {
	return func(t *testing.T) testing.RunFn {
		return fn
	}
}
