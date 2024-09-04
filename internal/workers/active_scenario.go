package workers

import (
	"log/slog"

	"github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/xtime"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type ActiveScenario struct {
	scenario     *scenarios.Scenario
	m            *metrics.Metrics
	progress     *progress.Stats
	t            *testing.T
	Teardown     func()
	logger       *slog.Logger
	logrusLogger *logrus.Logger
}

const instantDuration = 0

func NewActiveScenario(
	scenario *scenarios.Scenario,
	metricsInstance *metrics.Metrics,
	stats *progress.Stats,
	logger *slog.Logger,
	logrusLogger *logrus.Logger,
) *ActiveScenario {
	t, teardown := testing.NewTWithOptions(scenario.Name,
		testing.WithIteration("setup"),
		testing.WithLogger(logger),
		testing.WithLogrusLogger(logrusLogger),
	)

	s := &ActiveScenario{
		scenario:     scenario,
		m:            metricsInstance,
		t:            t,
		Teardown:     teardown,
		progress:     stats,
		logger:       logger,
		logrusLogger: logrusLogger,
	}

	return s
}

func (s *ActiveScenario) Setup() {
	start := xtime.NanoTime()
	func() {
		defer testing.CheckResults(s.t, nil)

		s.scenario.RunFunc = s.scenario.ScenarioFunc(s.t)
		s.scenario.RunFn = s.scenario.ScenarioFn(s.t)
	}()
	duration := xtime.NanoTime() - start

	// wait for completion
	s.m.RecordSetupResult(s.scenario.Name, metrics.Result(s.t.Failed()), duration)
}

func (s *ActiveScenario) newIterationState() *iterationState {
	t, teardown := testing.NewTWithOptions(s.scenario.Name,
		testing.WithLogger(s.logger),
		testing.WithLogrusLogger(s.logrusLogger),
	)

	return &iterationState{
		t:        t,
		teardown: teardown,
	}
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
		if s.scenario.RunFunc != nil {
			s.scenario.RunFunc(state.t)
		} else {
			s.scenario.RunFn(state.t)
		}
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
