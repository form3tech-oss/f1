package workers_test

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/workers"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	f1_testing "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

// stalledDeadlineContext has a deadline but is only cancelled by cancel, like a context
// whose deadline timer hasn't run yet.
type stalledDeadlineContext struct {
	deadline time.Time
	done     chan struct{}
	once     sync.Once
}

func newStalledDeadlineContext(deadline time.Time) *stalledDeadlineContext {
	return &stalledDeadlineContext{deadline: deadline, done: make(chan struct{})}
}

func (c *stalledDeadlineContext) Deadline() (time.Time, bool) { return c.deadline, true }

func (c *stalledDeadlineContext) Done() <-chan struct{} { return c.done }

func (c *stalledDeadlineContext) Err() error {
	select {
	case <-c.done:
		return context.DeadlineExceeded
	default:
		return nil
	}
}

func (*stalledDeadlineContext) Value(any) any { return nil }

func (c *stalledDeadlineContext) cancel() {
	c.once.Do(func() { close(c.done) })
}

type poolUnderTest struct {
	manager    *workers.PoolManager
	stats      *progress.Stats
	iterations *atomic.Uint32
	release    chan struct{}
}

func newPoolUnderTest(maxIterations uint64) poolUnderTest {
	release := make(chan struct{})
	close(release)
	return newPoolWithRelease(maxIterations, release)
}

// newBlockingPoolUnderTest returns a pool whose iterations block until releaseIterations
func newBlockingPoolUnderTest(maxIterations uint64) poolUnderTest {
	return newPoolWithRelease(maxIterations, make(chan struct{}))
}

func newPoolWithRelease(maxIterations uint64, release chan struct{}) poolUnderTest {
	iterations := &atomic.Uint32{}
	scenario := &scenarios.Scenario{
		Name: "test",
		RunFn: func(*f1_testing.T) {
			iterations.Add(1)
			<-release
		},
	}
	stats := &progress.Stats{}
	activeScenario := workers.NewActiveScenario(
		scenario,
		metrics.NewInstance(prometheus.NewRegistry(), true, nil),
		stats,
		slog.New(slog.DiscardHandler),
		logrus.New(),
	)

	return poolUnderTest{
		manager:    workers.New(maxIterations, activeScenario),
		stats:      stats,
		iterations: iterations,
		release:    release,
	}
}

func (p poolUnderTest) releaseIterations() {
	close(p.release)
}

// assertNotCompleted checks the pool's workers keep running for a while
func (p poolUnderTest) assertNotCompleted(t *testing.T) {
	t.Helper()

	completed := p.manager.WaitForCompletion()
	assert.Never(t, func() bool {
		select {
		case <-completed:
			return true
		default:
			return false
		}
	}, 50*time.Millisecond, time.Millisecond, "pool completed before its context was cancelled")
}

// assertNextIteration takes the next iteration and checks its number, to check how many
// iterations were taken before. Call it last, as it takes an iteration.
func (p poolUnderTest) assertNextIteration(t *testing.T, want uint64) {
	t.Helper()

	iteration, err := p.manager.NextIteration()
	require.NoError(t, err)
	assert.Equal(t, want, iteration, "unexpected number of iterations taken")
}

func (p poolUnderTest) waitForCompletion(t *testing.T) {
	t.Helper()

	select {
	case <-p.manager.WaitForCompletion():
	case <-time.After(5 * time.Second):
		require.FailNow(t, "workers did not stop", "iterations started: %d", p.iterations.Load())
	}
}
