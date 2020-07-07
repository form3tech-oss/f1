package testing

import (
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
)

// all scenarios by name
var scenarios = make(map[string]MultiStageSetupFn)
var scenarioDescriptions []ScenarioInfo

// scenarios which have been set up, so may receive run or teardown calls
// map[string]*ActiveScenario
var activeScenarios sync.Map

func Add(name string, testSetup SetupFn) {
	AddScenario(ScenarioInfo{Name: name}, testSetup)
}

func AddScenario(scenario ScenarioInfo, testSetup SetupFn) {
	log.Debugf("Registering test %s\n", scenario.Name)
	scenarioDescriptions = append(scenarioDescriptions, scenario)
	scenarios[scenario.Name] = func(t *T) ([]Stage, TeardownFn) {
		run, teardown := testSetup(t)
		return []Stage{{Name: "single", RunFn: run}}, teardown
	}
}

func GetScenario(scenarioName string) MultiStageSetupFn {
	return scenarios[scenarioName]
}

func GetScenarioNames() []string {
	names := make([]string, len(scenarios))
	index := 0
	for key := range scenarios {
		names[index] = key
		index++
	}
	sort.Strings(names)
	return names
}

func AddMultiStage(name string, testSetup MultiStageSetupFn) {
	log.Debugf("Registering test %s\n", name)
	scenarios[name] = testSetup
}

func WithNoSetup(fn func(t *T)) SetupFn {
	return func(t *T) (testFunction RunFn, teardownFunction TeardownFn) {
		testFunction = func(t *T) {
			fn(t)
		}
		return
	}
}
