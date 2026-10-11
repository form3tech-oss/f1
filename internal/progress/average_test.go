package progress_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/progress"
)

// CollectLifetime is called on every progress tick. It returns two snapshots:
//   - period:   only the iterations recorded since the previous call
//   - lifetime: every iteration recorded since the run started

func TestCollectLifetime_PeriodContainsOnlyRecordsSinceThePreviousCall(t *testing.T) {
	t.Parallel()

	var stats progress.DurationStats

	record(&stats, 10*time.Millisecond)
	stats.CollectLifetime()

	record(&stats, 20*time.Millisecond, 30*time.Millisecond)
	period, _ := stats.CollectLifetime()

	require.Equal(t, progress.IterationDurationsSnapshot{
		Count:   2,
		Average: 25 * time.Millisecond,
		Min:     20 * time.Millisecond,
		Max:     30 * time.Millisecond,
	}, period, "the 10ms record belongs to the previous period")
}

func TestCollectLifetime_LifetimeContainsRecordsFromEveryCall(t *testing.T) {
	t.Parallel()

	var stats progress.DurationStats

	record(&stats, 10*time.Millisecond)
	stats.CollectLifetime()

	record(&stats, 20*time.Millisecond, 30*time.Millisecond)
	_, lifetime := stats.CollectLifetime()

	require.Equal(t, progress.IterationDurationsSnapshot{
		Count:   3,
		Average: 20 * time.Millisecond,
		Min:     10 * time.Millisecond,
		Max:     30 * time.Millisecond,
	}, lifetime, "lifetime includes the 10ms record from the previous period")
}

func TestCollectLifetime_WithNoRecordsReturnsEmptySnapshots(t *testing.T) {
	t.Parallel()

	var stats progress.DurationStats

	period, lifetime := stats.CollectLifetime()

	require.Equal(t, progress.IterationDurationsSnapshot{}, period)
	require.Equal(t, progress.IterationDurationsSnapshot{}, lifetime)
}

// Workers keep recording while progress reporting calls CollectLifetime.
// A record that lands in the middle of a CollectLifetime call must not be lost.
func TestCollectLifetime_DoesNotLoseRecordsMadeWhileCollecting(t *testing.T) {
	t.Parallel()

	var stats progress.DurationStats

	stopCollecting := collectLifetimeInALoop(&stats)
	recordFromManyGoroutines(&stats, 8, 50_000)
	stopCollecting()

	_, lifetime := stats.CollectLifetime()

	const recordsMade uint64 = 8 * 50_000
	require.Equal(t, recordsMade, lifetime.Count,
		"lifetime counted %d of %d records", lifetime.Count, recordsMade)
}

func record(stats *progress.DurationStats, durations ...time.Duration) {
	for _, duration := range durations {
		stats.Record(int64(duration))
	}
}

// collectLifetimeInALoop calls CollectLifetime as fast as possible in the background
// until the returned stop function is called.
func collectLifetimeInALoop(stats *progress.DurationStats) func() {
	stopRequested := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		for {
			select {
			case <-stopRequested:
				return
			default:
				stats.CollectLifetime()
				runtime.Gosched()
			}
		}
	}()

	return func() {
		close(stopRequested)
		<-stopped
	}
}

// recordFromManyGoroutines records 1ns durations from several goroutines at once
// and returns when they have all finished.
func recordFromManyGoroutines(stats *progress.DurationStats, goroutines, recordsEach int) {
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range recordsEach {
				stats.Record(1)
			}
		})
	}
	wg.Wait()
}
