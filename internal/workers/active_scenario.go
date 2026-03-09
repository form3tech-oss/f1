package workers

import (
	"log/slog"

	"github.com/form3tech-oss/f1/v3/internal/metrics"
	"github.com/form3tech-oss/f1/v3/internal/progress"
	"github.com/form3tech-oss/f1/v3/internal/xtime"
	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
	"github.com/form3tech-oss/f1/v3/pkg/f1/scenarios"
)

type ActiveScenario struct {
	scenario *scenarios.Scenario
	m        *metrics.Metrics
	progress *progress.Stats
	t        *f1testing.T
	Teardown func()
	logger   *slog.Logger
}

const instantDuration = 0

func NewActiveScenario(
	scenario *scenarios.Scenario,
	metricsInstance *metrics.Metrics,
	stats *progress.Stats,
	logger *slog.Logger,
) *ActiveScenario {
	t, teardown := f1testing.NewTWithOptions(scenario.Name,
		f1testing.WithIteration("setup"),
		f1testing.WithVUID(-1),
		f1testing.WithLogger(logger),
	)

	s := &ActiveScenario{
		scenario: scenario,
		m:        metricsInstance,
		t:        t,
		Teardown: teardown,
		progress: stats,
		logger:   logger,
	}

	return s
}

func (s *ActiveScenario) Setup() {
	start := xtime.NanoTime()
	func() {
		defer f1testing.CheckResults(s.t, nil)

		s.scenario.RunFn = s.scenario.ScenarioFn(s.t)
	}()
	duration := xtime.NanoTime() - start

	// wait for completion
	s.m.RecordSetupResult(s.scenario.Name, metrics.Result(s.t.Failed()), duration)
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
		defer f1testing.CheckResults(state.t, nil)
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

func (s *ActiveScenario) newIterationState(id int) *iterationState {
	t, teardown := f1testing.NewTWithOptions(s.scenario.Name,
		f1testing.WithVUID(id),
		f1testing.WithLogger(s.logger),
	)

	return &iterationState{
		t:        t,
		teardown: teardown,
	}
}
