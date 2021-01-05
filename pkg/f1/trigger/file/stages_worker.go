package file

import (
	"os"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/options"
	"github.com/form3tech-oss/f1/pkg/f1/trace"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"
	log "github.com/sirupsen/logrus"
)

func newStagesWorker(stages []runnableStage) api.WorkTriggerer {
	return func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
		for _, stage := range stages {
			setEnvs(stage.params)

			if stage.users == 0 {
				runStageWork(workTriggered, stop, workDone, stage)
			} else {
				runUsersStageWork(workTriggered, stop, workDone, stage)
			}

			unsetEnvs(stage.params)
		}
	}
}

func runStageWork(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, stage runnableStage) {
	startRate := stage.rate(time.Now())
	for i := 0; i < startRate; i++ {
		workTriggered <- true
	}

	totalDurationTicker := time.NewTicker(stage.stageDuration - 10*time.Millisecond)
	iterationTicker := time.NewTicker(stage.iterationDuration)

	for isListening := true; isListening; {
		select {
		case <-workDone:
			continue
		case <-stop:
			trace.ReceivedFromChannel("stop")
			iterationTicker.Stop()
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

			iterationRate := stage.rate(start)
			for i := 0; i < iterationRate; i++ {
				trace.SendingToChannel("workTriggered")
				workTriggered <- true
			}
		case <-totalDurationTicker.C:
			iterationTicker.Stop()
			totalDurationTicker.Stop()
			isListening = false
		}
	}
}

func runUsersStageWork(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, stage runnableStage) {
	for i := 0; i < stage.users; i++ {
		workTriggered <- true
	}

	totalDurationTicker := time.NewTicker(stage.stageDuration - 10*time.Millisecond)

	for isListening := true; isListening; {
		select {
		case <-stop:
			return
		case <-workDone:
			workTriggered <- true
		case <-totalDurationTicker.C:
			totalDurationTicker.Stop()
			isListening = false
		}
	}
}

func setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		if err != nil {
			log.WithError(err).Error("unable set environment variables for given scenario")
		}
	}
}

func unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		if err != nil {
			log.WithError(err).Error("unable unset environment variables for given scenario")
		}
	}
}
