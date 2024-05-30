package workers

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/form3tech-oss/f1/v2/internal/trace"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

type iterationState struct {
	teardown func()
	t        *testing.T
}

func newIterationState(scenario string) *iterationState {
	state := &iterationState{}
	state.t, state.teardown = testing.NewT("", scenario)

	return state
}

type PoolManager struct {
	tracer         trace.Tracer
	activeScenario *ActiveScenario
	iteration      atomic.Uint64
	maxIterations  uint64
	runningWorkers sync.WaitGroup
}

func New(maxIterations uint64, activeScenario *ActiveScenario, tracer trace.Tracer) *PoolManager {
	w := &PoolManager{
		activeScenario: activeScenario,
		tracer:         tracer,
		maxIterations:  maxIterations,
	}

	return w
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

type Pool struct {
	workQueue          chan struct{}
	manager            *PoolManager
	workerCtxCancel    context.CancelFunc
	iterationStatePool []*iterationState
	totalWorkers       int
	busyWorkers        atomic.Int32
}

func (m *PoolManager) NewPool(totalWorkers int) *Pool {
	p := &Pool{
		totalWorkers:       totalWorkers,
		iterationStatePool: make([]*iterationState, totalWorkers),
		workQueue:          make(chan struct{}, totalWorkers),
		manager:            m,
	}

	for i := range totalWorkers {
		p.iterationStatePool[i] = newIterationState(m.activeScenario.scenario.Name)
	}

	return p
}

func (p *Pool) QueueOrDrop(ctx context.Context, triggersCount int) {
	for range triggersCount {
		if p.busyWorkers.Load() >= int32(p.totalWorkers) {
			p.manager.activeScenario.RecordDroppedIteration()
			continue
		}

		select {
		case <-ctx.Done():
			return
		case p.workQueue <- struct{}{}:
		}
	}
}

func (p *Pool) Queue(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case p.workQueue <- struct{}{}:
		return true
	}
}

func (p *Pool) Start(ctx context.Context) context.Context {
	p.manager.runningWorkers.Add(p.totalWorkers)

	startedWg := sync.WaitGroup{}
	startedWg.Add(p.totalWorkers)

	workerCtx, cancel := context.WithCancel(ctx)
	p.workerCtxCancel = cancel

	for i := range p.totalWorkers {
		go p.run(workerCtx, i, p.iterationStatePool[i], &startedWg)
	}
	startedWg.Wait()

	return workerCtx
}

func (p *Pool) run(
	ctx context.Context,
	worker int,
	iterationState *iterationState,
	startWg *sync.WaitGroup,
) {
	defer p.manager.runningWorkers.Done()

	p.manager.tracer.WorkerEvent("Started worker", worker)
	startWg.Done()
	for {
		select {
		case <-ctx.Done():
			p.manager.tracer.WorkerEvent("Stopping worker", worker)
			return
		case <-p.workQueue:
			if ctx.Err() != nil {
				return
			}

			iteration := p.manager.iteration.Add(1)
			if p.manager.maxIterations > 0 && iteration > p.manager.maxIterations {
				p.workerCtxCancel()
				return
			}
			p.manager.tracer.IterationEvent("Received work from Channel 'doWork'", iteration)

			iterationState.t.Reset(strconv.FormatUint(iteration, 10))

			p.busyWorkers.Add(1)
			p.manager.activeScenario.Run(iterationState)
			p.busyWorkers.Add(-1)

			p.manager.tracer.IterationEvent("Completed iteration", iteration)
		}
	}
}
