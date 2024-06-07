package workers

import (
	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/xtime"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type ActiveScenario struct {
	scenario *scenarios.Scenario
	m        *metrics.Metrics
	progress *progress.Stats
	t        *testing.T
	Teardown func()
}

const instantDuration = 0

func NewActiveScenario(
	scenario *scenarios.Scenario,
	metricsInstance *metrics.Metrics,
	stats *progress.Stats,
) *ActiveScenario {
	t, teardown := testing.NewT("setup", scenario.Name)

	s := &ActiveScenario{
		scenario: scenario,
		m:        metricsInstance,
		t:        t,
		Teardown: teardown,
		progress: stats,
	}

	start := xtime.NanoTime()
	func() {
		defer testing.CheckResults(t, nil)

		s.scenario.RunFn = s.scenario.ScenarioFn(t)
	}()
	duration := xtime.NanoTime() - start

	// wait for completion
	s.m.RecordSetupResult(scenario.Name, metrics.Result(t.Failed()), duration)
	return s
}

func (s *ActiveScenario) TeardownFailed() bool {
	return s.t.TeardownFailed()
}

func (s *ActiveScenario) Failed() bool {
	return s.t.Failed()
}

// Run performs a single iteration of the test.
func (s *ActiveScenario) Run(state *iterationState) {
	defer state.teardown()

	start := xtime.NanoTime()
	func() {
		defer testing.CheckResults(state.t, nil)
		s.scenario.RunFn(state.t)
	}()

	failed := state.t.Failed()
	duration := xtime.NanoTime() - start

	s.m.RecordIterationResult(s.scenario.Name, metrics.Result(failed), duration)
	s.progress.Record(metrics.Result(failed), duration)
}

func (s *ActiveScenario) RecordDroppedIteration() {
	s.m.RecordIterationResult(s.scenario.Name, metrics.DroppedResult, instantDuration)
	s.progress.Record(metrics.DroppedResult, instantDuration)
}
