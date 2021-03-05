package testing

import (
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Scenarios struct {
	scenarios            map[string]MultiStageSetupFn
	scenarioDescriptions []ScenarioInfo
}

// scenarios which have been set up, so may receive run or teardown calls
// map[string]*ActiveScenario
var activeScenarios sync.Map

func New() *Scenarios {
	return &Scenarios{
		scenarios: make(map[string]MultiStageSetupFn),
	}
}

func (s *Scenarios) Add(scenario ScenarioInfo, testSetup SetupFn) *Scenarios {
	s.scenarioDescriptions = append(s.scenarioDescriptions, scenario)

	multiStageSetup := func(t *T) ([]Stage, TeardownFn) {
		run, teardown := testSetup(t)
		return []Stage{{Name: "single", RunFn: run}}, teardown
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
	return func(t *T) (testFunction RunFn, teardownFunction TeardownFn) {
		testFunction = func(t *T) {
			fn(t)
		}
		return
	}
}
