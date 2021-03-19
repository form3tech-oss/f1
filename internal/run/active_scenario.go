package run

import (
	"time"

	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"

	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/google/uuid"
)

type ActiveScenario struct {
	scenario *scenarios.Scenario
	id       string
	m        *metrics.Metrics
	t        *testing.T
	Teardown func()
}

func NewActiveScenario(scenario *scenarios.Scenario) *ActiveScenario {
	t, teardown := testing.NewT("setup", scenario.Name)

	s := &ActiveScenario{
		scenario: scenario,
		id:       uuid.New().String(),
		m:        metrics.Instance(),
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
	s.m.Record(metrics.SetupResult, scenario.Name, "setup", metrics.Result(t.Failed()), time.Since(start).Nanoseconds())
	return s
}

// Run performs a single iteration of the test. It returns `true` if the test was successful, `false` otherwise.
func (s *ActiveScenario) Run(metric metrics.MetricType, stage, iter string, f func(t *testing.T)) bool {
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
	s.m.Record(metric, s.scenario.Name, stage, metrics.Result(t.Failed()), time.Since(start).Nanoseconds())
	return !t.Failed()
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.Record(metrics.IterationResult, s.scenario.Name, IterationStage, "dropped", 0)
}
