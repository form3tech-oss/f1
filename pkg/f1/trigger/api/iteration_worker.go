package api

import (
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/options"
)

// NewIterationWorker produces a WorkTriggerer which triggers work at fixed intervals.
func NewIterationWorker(iterationDuration time.Duration, rate RateFunction) WorkTriggerer {
	return func(doWork chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
		startRate := rate(time.Now())
		for i := 0; i < startRate; i++ {
			doWork <- true
		}

		// start ticker to trigger subsequent iterations.
		iterationTicker := time.NewTicker(iterationDuration)

		// run more iterations on every tick, until duration has elapsed.
		go func() {
			for {
				select {
				case <-workDone:
					continue
				case <-stop:
					iterationTicker.Stop()
					return
				case start := <-iterationTicker.C:
					iterationRate := rate(start)
					for i := 0; i < iterationRate; i++ {
						doWork <- true
					}
				}
			}
		}()
	}
}
