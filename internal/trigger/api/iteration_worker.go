package api

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trace"
)

// NewIterationWorker produces a WorkTriggerer which triggers work at fixed intervals.
func NewIterationWorker(iterationDuration time.Duration, rate RateFunction, tracer trace.Tracer) WorkTriggerer {
	return func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, _ options.RunOptions) {
		startRate := rate(time.Now())
		for range startRate {
			workTriggered <- true
		}

		// start ticker to trigger subsequent iterations.
		iterationTicker := time.NewTicker(iterationDuration)

		// run more iterations on every tick, until duration has elapsed.
		for {
			select {
			case <-workDone:
				continue
			case <-stop:
				tracer.ReceivedFromChannel("stop")
				iterationTicker.Stop()
				tracer.Event("Iteration worker stopped.")
				return
			case start := <-iterationTicker.C:
				// if both stop and the ticker are available at the same time
				// a `case` will be chosen at random.
				// double check the stop ch, continue to select again,
				// and expect its own handler to be called
				select {
				case <-stop:
					continue
				default:
				}

				iterationRate := rate(start)
				for range iterationRate {
					tracer.SendingToChannel("workTriggered")
					workTriggered <- true
				}
			}
		}
	}
}
