package workers

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
)

func newTriggerPool(m *PoolManager, numWorkers int) *TriggerPool {
	p := &TriggerPool{
		numWorkers:         numWorkers,
		iterationStatePool: make([]*iterationState, numWorkers),
		manager:            m,
		jobsAvailableCond:  sync.NewCond(&sync.Mutex{}),
	}

	for i := range numWorkers {
		p.iterationStatePool[i] = newIterationState(m.activeScenario.scenario.Name)
	}

	return p
}

type TriggerPool struct {
	manager         *PoolManager
	workerCtxCancel context.CancelFunc
	// jobsAvailableCond will notify blocked workers to start executing work again
	jobsAvailableCond  *sync.Cond
	iterationStatePool []*iterationState
	numWorkers         int
	// jobsToExecute holds a number of pending work to execute
	jobsToExecute jobCounter
	stopWorkers   atomic.Bool
}

// Trigger will trigger the execution of a numJobs in the worker pool,
// discarding anything that is currently scheduled for execution.
func (p *TriggerPool) Trigger(ctx context.Context, numJobs int) {
	if ctx.Err() != nil {
		return
	}
	p.sendJobsForExecution(numJobs)
}

func (p *TriggerPool) Start(ctx context.Context) context.Context {
	p.manager.runningWorkers.Add(p.numWorkers)

	startedWg := sync.WaitGroup{}
	startedWg.Add(p.numWorkers)

	workerCtx, cancel := context.WithCancel(ctx)
	p.workerCtxCancel = cancel

	for _, statePool := range p.iterationStatePool {
		go p.run(statePool, &startedWg)
	}

	// wait for all workers to start, to make sure we have the concurrency requested,
	// and work is not dropped
	startedWg.Wait()

	// context.Done() and context.Err() for context that can be cancelled use a Lock.
	// To avoid frequent locking - use an atomic.Bool for cancellation instead of checking the
	// context on each iteration
	go func() {
		<-workerCtx.Done()
		p.stop()
	}()

	return workerCtx
}

func (p *TriggerPool) running() bool {
	return !p.stopWorkers.Load()
}

func (p *TriggerPool) stop() {
	p.stopWorkers.Store(true)
	p.sendJobsForExecution(0)
}

func (p *TriggerPool) maxIterationsReached() {
	p.jobsToExecute.set(0)
	p.workerCtxCancel()
}

func (p *TriggerPool) sendJobsForExecution(numJobs int) {
	p.jobsAvailableCond.L.Lock()

	jobsDiscarded := p.jobsToExecute.set(numJobs)
	p.jobsAvailableCond.Broadcast()

	p.jobsAvailableCond.L.Unlock()

	for range jobsDiscarded {
		p.manager.activeScenario.RecordDroppedIteration()
	}
}

func (p *TriggerPool) waitForNewJobs() {
	p.jobsAvailableCond.L.Lock()

	for p.jobsToExecute.none() && p.running() {
		p.jobsAvailableCond.Wait()
	}
	p.jobsAvailableCond.L.Unlock()
}

func (p *TriggerPool) run(
	iterationState *iterationState,
	startWg *sync.WaitGroup,
) {
	defer p.manager.runningWorkers.Done()
	startWg.Done()

	for p.running() {
		if p.jobsToExecute.none() {
			p.waitForNewJobs()
		}

		if p.jobsToExecute.take() {
			iteration, err := p.manager.NextIteration()
			if err != nil {
				p.maxIterationsReached()
				return
			}

			iterationState.t.Reset(strconv.FormatUint(iteration, 10))
			p.manager.activeScenario.Run(iterationState)
		}
	}
}

type jobCounter struct {
	num atomic.Int64
}

func (w *jobCounter) set(n int) int64 {
	return w.num.Swap(int64(n))
}

func (w *jobCounter) none() bool {
	return w.num.Load() <= 0
}

func (w *jobCounter) take() bool {
	return w.num.Add(-1) >= 0
}
