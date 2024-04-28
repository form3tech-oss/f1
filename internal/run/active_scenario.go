package run

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type iterationState struct {
	teardown func()
	t        *testing.T
	done     chan struct{}
}

func newIterationState(scenario string) *iterationState {
	state := &iterationState{}
	state.t, state.teardown = testing.NewT("", scenario)
	state.done = make(chan struct{}, 1)

	return state
}

type ActiveScenario struct {
	scenario *scenarios.Scenario
	m        *metrics.Metrics
	t        *testing.T
	Teardown func()
}

func NewActiveScenario(scenario *scenarios.Scenario, metricsInstance *metrics.Metrics) *ActiveScenario {
	t, teardown := testing.NewT("setup", scenario.Name)

	s := &ActiveScenario{
		scenario: scenario,
		m:        metricsInstance,
		t:        t,
		Teardown: teardown,
	}

	start := time.Now()
	done := make(chan struct{}, 1)
	go func() {
		defer testing.CheckResults(t, done)
		s.scenario.RunFn = s.scenario.ScenarioFn(t)
	}()

	// wait for completion
	<-done
	s.m.RecordSetupResult(scenario.Name, metrics.Result(t.Failed()), time.Since(start).Nanoseconds())
	return s
}

// Run performs a single iteration of the test. It returns `true` if the test was successful, `false` otherwise.
func (s *ActiveScenario) Run(state *iterationState) bool {
	defer state.teardown()

	start := time.Now()
	go func() {
		defer testing.CheckResults(state.t, state.done)
		s.scenario.RunFn(state.t)
	}()

	// wait for completion
	<-state.done

	failed := state.t.Failed()
	s.m.RecordIterationResult(s.scenario.Name, metrics.Result(failed), time.Since(start).Nanoseconds())
	return !failed
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.RecordIterationResult(s.scenario.Name, metrics.DroppedResult, 0)
}
