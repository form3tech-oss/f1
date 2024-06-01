package raterun_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/form3tech-oss/f1/v2/internal/raterun"
)

type RatedRunnerStage struct {
	runner    *raterun.Runner
	funcRuns  map[time.Duration]int
	t         *testing.T
	cancelRun context.CancelFunc
	rates     []raterun.Schedule
	m         sync.Mutex
}

func NewRatedRunnerStage(t *testing.T) (*RatedRunnerStage, *RatedRunnerStage, *RatedRunnerStage) {
	t.Helper()

	stage := RatedRunnerStage{
		t:        t,
		funcRuns: make(map[time.Duration]int),
	}
	return &stage, &stage, &stage
}

func (s *RatedRunnerStage) some_rates(rates []raterun.Schedule) *RatedRunnerStage {
	s.rates = rates
	return s
}

func (s *RatedRunnerStage) and() *RatedRunnerStage {
	return s
}

func (s *RatedRunnerStage) a_rate_runner() *RatedRunnerStage {
	runner, err := raterun.New(func(rate time.Duration) {
		s.m.Lock()
		defer s.m.Unlock()
		s.funcRuns[rate]++
	}, s.rates)

	require.NoError(s.t, err)

	s.runner = runner
	return s
}

func (s *RatedRunnerStage) runner_is_run() *RatedRunnerStage {
	ctx, cancel := context.WithCancel(context.TODO())
	s.cancelRun = cancel
	s.runner.Run(ctx)
	return s
}

func (s *RatedRunnerStage) time_passes(dur time.Duration) *RatedRunnerStage {
	time.Sleep(dur)
	return s
}

func (s *RatedRunnerStage) runner_is_terminated() *RatedRunnerStage {
	s.cancelRun()
	return s
}

func (s *RatedRunnerStage) function_ran_times(expectedRuns int) *RatedRunnerStage {
	s.m.Lock()
	defer s.m.Unlock()

	totalRuns := 0
	for _, timesRunForRate := range s.funcRuns {
		totalRuns += timesRunForRate
	}

	assert.Equal(s.t, expectedRuns, totalRuns)
	return s
}

func (s *RatedRunnerStage) runner_is_reset() *RatedRunnerStage {
	s.runner.Restart()
	return s
}

func (s *RatedRunnerStage) a_go_leak_is_found() *RatedRunnerStage {
	err := goleak.Find()
	assert.Error(s.t, err, "should have found a go leak")
	return s
}

func (s *RatedRunnerStage) a_go_leak_is_not_found() *RatedRunnerStage {
	err := goleak.Find()
	assert.NoError(s.t, err, "should not have found a go leak")
	return s
}
