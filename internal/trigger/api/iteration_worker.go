package api

import (
	"context"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

// NewIterationWorker produces a WorkTriggerer which triggers work at fixed intervals.
func NewIterationWorker(iterationDuration time.Duration, rate RateFunction) WorkTriggerer {
	return func(ctx context.Context, _ *ui.Output, workers *workers.PoolManager, opts options.RunOptions) {
		pool := workers.NewTriggerPool(opts.Concurrency)
		workerCtx := pool.Start(ctx)

		// the context's deadline timer cancels it from a separate goroutine, which may not
		// have run yet, so don't trigger at or past the deadline. Check the current time,
		// not the tick's, as a tick can be read well after it was sent.
		deadline, hasDeadline := workerCtx.Deadline()
		trigger := func(tick time.Time) bool {
			if hasDeadline && !time.Now().Before(deadline) {
				return false
			}
			pool.Trigger(workerCtx, rate(tick))
			return true
		}

		// take the start time and start the ticker together, after the pool has started, so
		// time-based rates see ticks at whole multiples of iterationDuration from startTime.
		startTime := time.Now()
		iterationTicker := time.NewTicker(iterationDuration)
		defer iterationTicker.Stop()

		if !trigger(startTime) {
			return
		}

		// run more iterations on every tick, until duration has elapsed.
		for {
			select {
			case <-workerCtx.Done():
				return
			case start := <-iterationTicker.C:
				if !trigger(start) {
					return
				}
			}
		}
	}
}
