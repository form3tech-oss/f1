package progress_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/progress"
)

func TestSnapshot_FailedIterationsRate_NoIterations(t *testing.T) {
	t.Parallel()

	snapshot := progress.Snapshot{}

	assert.Zero(t, snapshot.FailedIterationsRate())
}

func TestStats_SnapshotDoesNotLoseConcurrentRecords(t *testing.T) {
	t.Parallel()

	const iterations uint64 = 100_000
	stats := &progress.Stats{}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range iterations {
			stats.Record(metrics.SuccessResult, 1)
		}
	}()

	var periodTotal uint64
	for running := true; running; {
		select {
		case <-done:
			running = false
		default:
			periodTotal += stats.Snapshot(time.Second).SuccessfulIterationDurationsForPeriod.Count
		}
	}

	last := stats.Snapshot(time.Second)
	periodTotal += last.SuccessfulIterationDurationsForPeriod.Count

	assert.Equal(t, iterations, last.SuccessfulIterationDurations.Count)
	assert.Equal(t, iterations, periodTotal, "every record must be counted in exactly one period")
	assert.Equal(t, iterations, stats.Total().SuccessfulIterationDurations.Count)
}
