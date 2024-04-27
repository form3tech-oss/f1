package run

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

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
	done := make(chan struct{})
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
func (s *ActiveScenario) Run(iter string, f func(t *testing.T)) bool {
	t, teardown := testing.NewT(iter, s.scenario.Name)
	defer teardown()

	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer testing.CheckResults(t, done)
		f(t)
	}()

	// wait for completion
	<-done
	s.m.RecordIterationResult(s.scenario.Name, metrics.Result(t.Failed()), time.Since(start).Nanoseconds())
	return !t.Failed()
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.RecordIterationResult(s.scenario.Name, metrics.DroppedResult, 0)
}
