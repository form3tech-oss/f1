package workers

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
)

func newContinuousPool(m *PoolManager, numWorkers int) *ContinuousPool {
	return &ContinuousPool{
		numWorkers:         numWorkers,
		iterationStatePool: m.makeIterationStatePool(numWorkers),
		manager:            m,
	}
}

type ContinuousPool struct {
	manager            *PoolManager
	workerCtxCancel    context.CancelFunc
	iterationStatePool []*iterationState
	numWorkers         int
	stopWorkers        atomic.Bool
}

func (p *ContinuousPool) Start(ctx context.Context) {
	workerCtx, workerCtxCancel := context.WithCancel(ctx)
	p.workerCtxCancel = workerCtxCancel

	workersStarted := sync.WaitGroup{}

	workersStarted.Add(p.numWorkers)
	p.manager.runningWorkers.Add(p.numWorkers)
	for _, iterationState := range p.iterationStatePool {
		go p.startWorker(iterationState, &workersStarted)
	}

	// context.Done() and context.Err() for context that can be cancelled use a Lock.
	// To avoid frequent locking - use an atomic.Bool for cancellation instead of checking the
	// context on each iteration
	go func() {
		<-workerCtx.Done()
		p.stopWorkers.Store(true)
	}()
}

func (p *ContinuousPool) maxIterationsReached() {
	p.workerCtxCancel()
}

func (p *ContinuousPool) startWorker(
	iterationState *iterationState,
	workersStarted *sync.WaitGroup,
) {
	defer p.manager.runningWorkers.Done()

	// wait for all workers to start before execution to make sure we're executing at the
	// concurrency requested
	workersStarted.Done()
	workersStarted.Wait()

	// use and atomic.Bool to control execution to avoid mutex usage in channels and context.Context
	for !p.stopWorkers.Load() {
		iteration, err := p.manager.NextIteration()
		if err != nil {
			p.maxIterationsReached()
			return
		}

		iterationState.t.Reset(strconv.FormatUint(iteration, 10))
		p.manager.activeScenario.Run(iterationState)
	}
}
