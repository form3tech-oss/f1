package raterun

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"

	"github.com/stretchr/testify/assert"
)

type RatedRunnerStage struct {
	rates    []Rate
	runner   Runner
	m        sync.Mutex
	funcRuns map[time.Duration]int
	t        *testing.T
}

func NewRatedRunnerStage(t *testing.T) (*RatedRunnerStage, *RatedRunnerStage, *RatedRunnerStage) {
	stage := RatedRunnerStage{
		t:        t,
		funcRuns: make(map[time.Duration]int),
	}
	return &stage, &stage, &stage
}

func (s *RatedRunnerStage) some_rates(rates []Rate) *RatedRunnerStage {
	s.rates = rates
	return s
}

func (s *RatedRunnerStage) and() *RatedRunnerStage {
	return s
}

func (s *RatedRunnerStage) a_rate_runner() *RatedRunnerStage {
	runner, _ := New(func(rate time.Duration, t time.Time) {
		s.m.Lock()
		defer s.m.Unlock()
		s.funcRuns[rate]++
	}, s.rates)
	s.runner = runner
	return s
}

func (s *RatedRunnerStage) runner_is_run() *RatedRunnerStage {
	s.runner.Run()
	return s
}

func (s *RatedRunnerStage) time_passes(dur time.Duration) *RatedRunnerStage {
	time.Sleep(dur)
	return s
}

func (s *RatedRunnerStage) runner_is_terminated() *RatedRunnerStage {
	s.runner.Terminate()
	return s
}

func (s *RatedRunnerStage) function_ran_times(expectedRuns int) *RatedRunnerStage {
	s.m.Lock()
	defer s.m.Unlock()

	totalRuns := 0
	for _, timesRunForRate := range s.funcRuns {
		totalRuns += timesRunForRate
	}

	assert.Equal(s.t, totalRuns, expectedRuns)
	return s
}

func (s *RatedRunnerStage) runner_is_reset() *RatedRunnerStage {
	s.runner.RestartRate()
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
