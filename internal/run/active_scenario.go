package run

import (
	"runtime/debug"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/scenarios"

	"github.com/form3tech-oss/f1/pkg/f1/testing"

	"github.com/form3tech-oss/f1/internal/metrics"
	"github.com/google/uuid"
)

type ActiveScenario struct {
	scenario *scenarios.Scenario
	id       string
	m        *metrics.Metrics
	t        *testing.T
	Teardown func()
}

func NewActiveScenario(scenario *scenarios.Scenario) (*ActiveScenario, bool) {
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
		defer checkResults(t, done)
		s.scenario.RunFn = s.scenario.ScenarioFn(t)
	}()

	// wait for completion
	<-done
	s.m.Record(metrics.SetupResult, scenario.Name, "setup", metrics.Result(t.HasFailed()), time.Since(start).Nanoseconds())
	return s, !t.HasFailed()
}

// Run performs a single iteration of the test. It returns `true` if the test was successful, `false` otherwise.
func (s *ActiveScenario) Run(metric metrics.MetricType, stage, iter string, f func(t *testing.T)) bool {
	t, teardown := testing.NewT(iter, s.scenario.Name)
	defer teardown()

	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer checkResults(t, done)
		f(t)
	}()

	// wait for completion
	<-done
	s.m.Record(metric, s.scenario.Name, stage, metrics.Result(t.HasFailed()), time.Since(start).Nanoseconds())
	return !t.HasFailed()
}

func checkResults(t *testing.T, done chan<- struct{}) {
	r := recover()
	if r != nil {
		err, isError := r.(error)
		if isError {
			t.FailWithError(err)
			debug.PrintStack()
		} else {
			t.Errorf("panic in %s: %v", t.Iteration, err)
		}
	}
	close(done)
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.Record(metrics.IterationResult, s.scenario.Name, IterationStage, "dropped", 0)
}
