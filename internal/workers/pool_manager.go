package workers

import (
	"context"
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
	runningWorkers workerGroup
	iteration      atomic.Uint64
	maxIterations  uint64
}

// workerGroup counts running workers across all pools. Unlike sync.WaitGroup, it can be
// waited on any number of times without a goroutine per waiter, and new workers can start
// while earlier callers are still waiting.
type workerGroup struct {
	mu      sync.Mutex
	count   int
	stopped chan struct{} // closed when count drops to zero
}

func (g *workerGroup) Add(n int) {
	if n < 0 {
		panic("workers: negative workerGroup Add")
	}
	if n == 0 {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.count == 0 {
		g.stopped = make(chan struct{})
	}
	g.count += n
}

func (g *workerGroup) Done() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.count == 0 {
		panic("workers: workerGroup Done called more times than Add")
	}
	g.count--
	if g.count == 0 {
		close(g.stopped)
	}
}

func (g *workerGroup) wait() <-chan struct{} {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.stopped == nil {
		g.stopped = make(chan struct{})
		close(g.stopped)
	}

	return g.stopped
}

func New(maxIterations uint64, activeScenario *ActiveScenario) *PoolManager {
	w := &PoolManager{
		activeScenario: activeScenario,
		maxIterations:  maxIterations,
	}

	return w
}

// WaitForCompletion returns a channel that is closed once all running workers have stopped.
func (m *PoolManager) WaitForCompletion() <-chan struct{} {
	return m.runningWorkers.wait()
}

// WaitForCompletionOrDone blocks until all running workers have stopped or ctx ends,
// whichever comes first.
func (m *PoolManager) WaitForCompletionOrDone(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-m.WaitForCompletion():
	}
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

func (m *PoolManager) makeIterationStatePool(numWorkers int) []*iterationState {
	statePool := make([]*iterationState, numWorkers)
	for i := range numWorkers {
		statePool[i] = m.activeScenario.newIterationState(i)
	}

	return statePool
}
