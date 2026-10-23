package workers_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startBusyPoolWithPendingJobs starts a pool with every worker busy on an iteration and 10
// jobs pending, and waits until its deadline has passed
func startBusyPoolWithPendingJobs(t *testing.T, p poolUnderTest) *stalledDeadlineContext {
	t.Helper()

	deadline := time.Now().Add(100 * time.Millisecond)
	ctx := newStalledDeadlineContext(deadline)
	t.Cleanup(ctx.cancel)

	pool := p.manager.NewTriggerPool(5)
	pool.Start(ctx)

	require.True(t, pool.Trigger(ctx, 5))
	require.Eventually(t, func() bool {
		return p.iterations.Load() == 5
	}, 5*time.Second, time.Millisecond, "workers did not start iterations")
	require.True(t, pool.Trigger(ctx, 10))

	time.Sleep(time.Until(deadline))
	assert.False(t, pool.Trigger(ctx, 10), "triggered past the deadline")

	return ctx
}

func TestTriggerPool_DoesNotStartIterationsPastDeadline(t *testing.T) {
	t.Parallel()

	for name, maxIterations := range map[string]uint64{
		"no iteration limit":       0,
		"limit below dropped jobs": 10,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := newBlockingPoolUnderTest(maxIterations)
			ctx := startBusyPoolWithPendingJobs(t, p)
			p.releaseIterations()

			// workers drop the pending jobs instead of starting them, but keep running until
			// the context is cancelled
			require.Eventually(t, func() bool {
				return p.stats.Total().DroppedIterationCount == 10
			}, 5*time.Second, time.Millisecond, "jobs were not dropped")
			p.assertNotCompleted(t)

			ctx.cancel()
			p.waitForCompletion(t)

			assert.Equal(t, uint32(5), p.iterations.Load())
			assert.Equal(t, uint64(10), p.stats.Total().DroppedIterationCount)
			assert.False(t, p.manager.MaxIterationsReached())
			p.assertNextIteration(t, 6)
		})
	}
}

func TestTriggerPool_StopsWithoutDroppingPastDeadlineWhenIterationLimitReached(t *testing.T) {
	t.Parallel()

	// the busy workers take all allowed iterations
	p := newBlockingPoolUnderTest(5)
	startBusyPoolWithPendingJobs(t, p)
	p.releaseIterations()

	// the pool stops itself at the iteration limit, without the context being cancelled
	p.waitForCompletion(t)

	assert.Equal(t, uint32(5), p.iterations.Load())
	assert.Zero(t, p.stats.Total().DroppedIterationCount)
	assert.True(t, p.manager.MaxIterationsReached())
}

func TestTriggerPool_StartsIterationsBeforeDeadline(t *testing.T) {
	t.Parallel()

	for name, deadline := range map[string]time.Time{
		"no deadline":     {},
		"future deadline": time.Now().Add(time.Hour),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := newPoolUnderTest(0)

			var ctx context.Context
			var cancel context.CancelFunc
			if deadline.IsZero() {
				ctx, cancel = context.WithCancel(context.Background())
			} else {
				ctx, cancel = context.WithDeadline(context.Background(), deadline)
			}
			t.Cleanup(cancel)

			pool := p.manager.NewTriggerPool(5)
			pool.Start(ctx)
			pool.Trigger(ctx, 10)

			assert.Eventually(t, func() bool {
				return p.iterations.Load() == 10
			}, 5*time.Second, 10*time.Millisecond)

			cancel()
			p.waitForCompletion(t)

			assert.Equal(t, uint32(10), p.iterations.Load())
			assert.Zero(t, p.stats.Total().DroppedIterationCount)
		})
	}
}
