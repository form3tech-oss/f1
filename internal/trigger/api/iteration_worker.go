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
		startRate := rate(time.Now())

		pool := workers.NewTriggerPool(opts.Concurrency)
		workerCtx := pool.Start(ctx)

		pool.Trigger(workerCtx, startRate)

		// start ticker to trigger subsequent iterations.
		iterationTicker := time.NewTicker(iterationDuration)
		defer iterationTicker.Stop()

		// run more iterations on every tick, until duration has elapsed.
		for {
			select {
			case <-workerCtx.Done():
				return
			case start := <-iterationTicker.C:
				iterationRate := rate(start)
				pool.Trigger(workerCtx, iterationRate)
			}
		}
	}
}
