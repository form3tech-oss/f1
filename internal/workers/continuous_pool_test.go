package workers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContinuousPool_DoesNotStartIterationsPastDeadline(t *testing.T) {
	t.Parallel()

	p := newPoolUnderTest(0)

	ctx := newStalledDeadlineContext(time.Now().Add(-time.Second))
	t.Cleanup(ctx.cancel)

	pool := p.manager.NewContinuousPool(5)
	pool.Start(ctx)

	// workers don't start iterations, but keep running until the context is cancelled
	p.assertNotCompleted(t)

	ctx.cancel()
	p.waitForCompletion(t)

	assert.Zero(t, p.iterations.Load())
	assert.Zero(t, p.stats.Total().DroppedIterationCount)
	p.assertNextIteration(t, 1)
}

func TestContinuousPool_StopsPastDeadlineWhenIterationLimitReached(t *testing.T) {
	t.Parallel()

	p := newPoolUnderTest(2)
	// start both allowed iterations
	for range 2 {
		_, err := p.manager.NextIteration()
		require.NoError(t, err)
	}

	ctx := newStalledDeadlineContext(time.Now().Add(-time.Second))
	t.Cleanup(ctx.cancel)

	pool := p.manager.NewContinuousPool(5)
	pool.Start(ctx)

	// the pool stops itself at the iteration limit, without the context being cancelled
	p.waitForCompletion(t)

	assert.Zero(t, p.iterations.Load())
	assert.Zero(t, p.stats.Total().DroppedIterationCount)
	assert.True(t, p.manager.MaxIterationsReached())
}
