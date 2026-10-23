package workers

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
)

func newTriggerPool(m *PoolManager, numWorkers int) *TriggerPool {
	return &TriggerPool{
		numWorkers:         numWorkers,
		iterationStatePool: m.makeIterationStatePool(numWorkers),
		manager:            m,
		jobsAvailableCond:  sync.NewCond(&sync.Mutex{}),
	}
}

type TriggerPool struct {
	manager         *PoolManager
	workerCtxCancel context.CancelFunc
	// jobsAvailableCond will notify blocked workers to start executing work again
	jobsAvailableCond  *sync.Cond
	iterationStatePool []*iterationState
	numWorkers         int
	deadline           deadline
	// jobsToExecute holds a number of pending work to execute
	jobsToExecute jobCounter
	stopWorkers   atomic.Bool
}

// Trigger will trigger the execution of a numJobs in the worker pool,
// discarding anything that is currently scheduled for execution. It returns false without
// triggering once ctx is done or the pool's deadline has passed.
func (p *TriggerPool) Trigger(ctx context.Context, numJobs int) bool {
	if ctx.Err() != nil || p.deadline.passed() {
		return false
	}
	p.sendJobsForExecution(numJobs)
	return true
}

func (p *TriggerPool) Start(ctx context.Context) context.Context {
	p.manager.runningWorkers.Add(p.numWorkers)

	startedWg := sync.WaitGroup{}
	startedWg.Add(p.numWorkers)

	workerCtx, cancel := context.WithCancel(ctx)
	p.workerCtxCancel = cancel
	p.deadline = newDeadline(workerCtx)

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

	p.recordDropped(jobsDiscarded)
}

func (p *TriggerPool) recordDropped(numJobs int64) {
	for range numJobs {
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
			// don't start iterations past the deadline. Drop this job and the pending ones,
			// and wait for the context to be cancelled, which stops the pool. Check before
			// taking an iteration, so dropped jobs don't use up the iteration limit. If the
			// limit is reached, NextIteration fails and stops the pool instead.
			if p.deadline.passed() && !p.manager.iterationsExhausted() {
				p.recordDropped(1 + max(p.jobsToExecute.set(0), 0))
				continue
			}

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
