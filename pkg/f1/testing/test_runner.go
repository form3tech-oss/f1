package testing

import (
	"sort"

	log "github.com/sirupsen/logrus"
)

type Scenarios struct {
	scenarios            map[string]MultiStageSetupFn
	scenarioDescriptions []ScenarioInfo
}

func New() *Scenarios {
	return &Scenarios{
		scenarios: make(map[string]MultiStageSetupFn),
	}
}

func (s *Scenarios) Add(scenario ScenarioInfo, testSetup SetupFn) *Scenarios {
	s.scenarioDescriptions = append(s.scenarioDescriptions, scenario)

	multiStageSetup := func(t *T) []Stage {
		run := testSetup(t)
		return []Stage{{Name: "single", RunFn: run}}
	}

	return s.AddMultiStage(scenario.Name, multiStageSetup)
}

func (s *Scenarios) AddByName(name string, testSetup SetupFn) *Scenarios {
	return s.Add(ScenarioInfo{Name: name}, testSetup)
}

func (s *Scenarios) AddMultiStage(name string, testSetup MultiStageSetupFn) *Scenarios {
	log.Debugf("Registering test %s\n", name)
	s.scenarios[name] = testSetup
	return s
}

func (s *Scenarios) GetScenario(scenarioName string) MultiStageSetupFn {
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

func WithNoSetup(fn func(t *T)) SetupFn {
	return func(t *T) RunFn {
		return fn
	}
}
