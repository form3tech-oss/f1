package api

import (
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trace"

	"github.com/form3tech-oss/f1/pkg/f1/options"
)

// NewIterationWorker produces a WorkTriggerer which triggers work at fixed intervals.
func NewIterationWorker(iterationDuration time.Duration, rate RateFunction) WorkTriggerer {
	return func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
		DoWork(workTriggered, stop, workDone, iterationDuration, options.MaxDuration, rate)

		//startRate := rate(time.Now())
		//for i := 0; i < startRate; i++ {
		//	workTriggered <- true
		//}
		//
		//// start ticker to trigger subsequent iterations.
		//iterationTicker := time.NewTicker(iterationDuration)
		//
		//// run more iterations on every tick, until duration has elapsed.
		//go func() {
		//	for {
		//		select {
		//		case <-workDone:
		//			continue
		//		case <-stop:
		//			trace.ReceivedFromChannel("stop")
		//			iterationTicker.Stop()
		//			trace.Event("Iteration worker stopped.")
		//			return
		//		case start := <-iterationTicker.C:
		//			// if both stop and the ticker are available at the same time
		//			// a `case` will be chosen at random.
		//			// double check the stop ch, continue to select again,
		//			// and expect its own handler to be called
		//			select {
		//			case <-stop:
		//				continue
		//			default:
		//			}
		//
		//			iterationRate := rate(start)
		//			for i := 0; i < iterationRate; i++ {
		//				trace.SendingToChannel("workTriggered")
		//				workTriggered <- true
		//			}
		//		}
		//	}
		//}()
	}
}

func DoWork(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, iterationDuration, totalDuration time.Duration, rate RateFunction) {
	if iterationDuration == 0 || totalDuration == 0 {
		return
	}

	startRate := rate(time.Now())
	for i := 0; i < startRate; i++ {
		workTriggered <- true
	}

	totalDurationTicker := time.NewTicker(totalDuration)
	iterationTicker := time.NewTicker(iterationDuration)

	for {
		select {
		case <-workDone:
			continue
		case <-stop:
			trace.ReceivedFromChannel("stop")
			iterationTicker.Stop()
			totalDurationTicker.Stop()
			trace.Event("Iteration worker stopped.")
			return
		case start := <-iterationTicker.C:
			select {
			case <-stop:
				continue
			case <-totalDurationTicker.C:
				continue
			default:
			}

			iterationRate := rate(start)
			for i := 0; i < iterationRate; i++ {
				trace.SendingToChannel("workTriggered")
				workTriggered <- true
			}
		case <-totalDurationTicker.C:
			iterationTicker.Stop()
			totalDurationTicker.Stop()
			return
		}
	}
}
