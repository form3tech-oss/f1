package workers_test

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/workers"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"
	f1_testing "github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

func TestPoolManager_WaitForCompletion_WithoutWorkers(t *testing.T) {
	t.Parallel()

	manager := workers.New(0, newActiveScenario(t, func(*f1_testing.T) {}))

	requireClosed(t, manager.WaitForCompletion())
}

func TestPoolManager_WaitForCompletion_ClosesWhenWorkersFinish(t *testing.T) {
	t.Parallel()

	running := make(chan struct{}, 2)
	release := make(chan struct{})
	manager := workers.New(0, newActiveScenario(t, signalRunningAndWaitForRelease(running, release)))

	ctx, cancel := context.WithCancel(t.Context())
	manager.NewContinuousPool(2).Start(ctx)
	<-running
	<-running
	cancel()

	completed := manager.WaitForCompletion()
	requireOpen(t, completed)

	close(release)

	requireClosed(t, completed)
}

// Callers share one completion channel rather than each getting a goroutine blocked
// until the workers finish.
func TestPoolManager_WaitForCompletion_ReturnsTheSameChannelToEveryCaller(t *testing.T) {
	t.Parallel()

	running := make(chan struct{}, 1)
	release := make(chan struct{})
	manager := workers.New(0, newActiveScenario(t, signalRunningAndWaitForRelease(running, release)))

	ctx, cancel := context.WithCancel(t.Context())
	manager.NewContinuousPool(1).Start(ctx)
	<-running
	cancel()

	first := manager.WaitForCompletion()
	second := manager.WaitForCompletion()

	close(release)

	require.Equal(t, first, second)
	requireClosed(t, first)
}

// Pools for consecutive stages start while callers of WaitForCompletion from the
// previous stage may still be waiting. This must not panic.
func TestPoolManager_StartingPoolsWhileWaitingForCompletion(t *testing.T) {
	t.Parallel()

	manager := workers.New(0, newActiveScenario(t, func(*f1_testing.T) {}))

	for range 2000 {
		ctx, cancel := context.WithCancel(t.Context())
		manager.NewContinuousPool(4).Start(ctx)
		manager.WaitForCompletion()
		cancel()
	}

	requireClosed(t, manager.WaitForCompletion())
}

// signalRunningAndWaitForRelease returns an iteration that reports it is running and then
// blocks until release is closed. Reports are dropped once running is full, so iterations
// started after release never block.
func signalRunningAndWaitForRelease(running chan<- struct{}, release <-chan struct{}) f1_testing.RunFn {
	return func(*f1_testing.T) {
		select {
		case running <- struct{}{}:
		default:
		}
		<-release
	}
}

func newActiveScenario(t *testing.T, runFn f1_testing.RunFn) *workers.ActiveScenario {
	t.Helper()

	logger := log.NewDiscardLogger()
	activeScenario := workers.NewActiveScenario(
		&scenarios.Scenario{
			Name: "test",
			ScenarioFn: func(*f1_testing.T) f1_testing.RunFn {
				return runFn
			},
		},
		metrics.NewInstance(prometheus.NewRegistry(), true, nil),
		&progress.Stats{},
		logger,
		log.NewSlogLogrusLogger(logger),
	)
	activeScenario.Setup()

	return activeScenario
}

func requireClosed(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		require.FailNow(t, "channel was not closed")
	}
}

func requireOpen(t *testing.T, ch <-chan struct{}) {
	t.Helper()

	select {
	case <-ch:
		require.FailNow(t, "channel was closed")
	case <-time.After(50 * time.Millisecond):
	}
}
