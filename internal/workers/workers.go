package workers

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type iterationState struct {
	teardown func()
	t        *testing.T
}

func newIterationState(scenario string, logrusLogger *logrus.Logger) *iterationState {
	state := &iterationState{}
	state.t, state.teardown = testing.NewTWithOptions(scenario, testing.WithLogrusLogger(logrusLogger))

	return state
}

type PoolManager struct {
	activeScenario *ActiveScenario
	logrusLogger   *logrus.Logger
	runningWorkers sync.WaitGroup
	iteration      atomic.Uint64
	maxIterations  uint64
}

func New(maxIterations uint64, activeScenario *ActiveScenario, logrusLogger *logrus.Logger) *PoolManager {
	w := &PoolManager{
		activeScenario: activeScenario,
		maxIterations:  maxIterations,
		logrusLogger:   logrusLogger,
	}

	return w
}

func (m *PoolManager) makeIterationStatePool(numWorkers int) []*iterationState {
	statePool := make([]*iterationState, numWorkers)
	for i := range numWorkers {
		statePool[i] = newIterationState(m.activeScenario.scenario.Name, m.logrusLogger)
	}

	return statePool
}

func (m *PoolManager) Logger() *logrus.Logger {
	return m.logrusLogger
}

func (m *PoolManager) WaitForCompletion() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.runningWorkers.Wait()
	}()
	return done
}

func (m *PoolManager) MaxIterationsReached() bool {
	if m.maxIterations > 0 && m.iteration.Load() > m.maxIterations {
		return true
	}

	return false
}

var errMaxIterationsReached = errors.New("max iterations reached")

func (m *PoolManager) NextIteration() (uint64, error) {
	iteration := m.iteration.Add(1)
	if m.maxIterations > 0 && iteration > m.maxIterations {
		return 0, errMaxIterationsReached
	}

	return iteration, nil
}

func (m *PoolManager) NewTriggerPool(numWorkers int) *TriggerPool {
	return newTriggerPool(m, numWorkers)
}

func (m *PoolManager) NewContinuousPool(numWorkers int) *ContinuousPool {
	return newContinuousPool(m, numWorkers)
}
