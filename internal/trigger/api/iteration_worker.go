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

		// take the start time and start the ticker together, after the pool has started, so
		// time-based rates see ticks at whole multiples of iterationDuration from startTime.
		startTime := time.Now()
		iterationTicker := time.NewTicker(iterationDuration)
		defer iterationTicker.Stop()

		if !pool.Trigger(workerCtx, rate(startTime)) {
			return
		}

		// run more iterations on every tick, until duration has elapsed.
		for {
			select {
			case <-workerCtx.Done():
				return
			case start := <-iterationTicker.C:
				if !pool.Trigger(workerCtx, rate(start)) {
					return
				}
			}
		}
	}
}
