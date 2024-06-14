package workers

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type iterationState struct {
	teardown func()
	t        *testing.T
}

type PoolManager struct {
	activeScenario *ActiveScenario
	runningWorkers sync.WaitGroup
	iteration      atomic.Uint64
	maxIterations  uint64
}

func New(maxIterations uint64, activeScenario *ActiveScenario) *PoolManager {
	w := &PoolManager{
		activeScenario: activeScenario,
		maxIterations:  maxIterations,
	}

	return w
}

func (m *PoolManager) makeIterationStatePool(numWorkers int) []*iterationState {
	statePool := make([]*iterationState, numWorkers)
	for i := range numWorkers {
		statePool[i] = m.activeScenario.newIterationState()
	}

	return statePool
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
