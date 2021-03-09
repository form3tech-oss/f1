package file

import (
	"os"
	"sync"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/trigger/users"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	log "github.com/sirupsen/logrus"
)

func newStagesWorker(stages []runnableStage) api.WorkTriggerer {
	return func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
		safeThresholdBeforeNextIteration := 20 * time.Millisecond
		stopStageCh := make(chan bool)
		wg := sync.WaitGroup{}

		for _, stage := range stages {
			wg.Add(1)
			setEnvs(stage.params)

			stopStageTicker := time.NewTicker(stage.stageDuration - safeThresholdBeforeNextIteration)

			go func() {
				defer wg.Done()

				if stage.usersConcurrency == 0 {
					doWork := api.NewIterationWorker(stage.iterationDuration, stage.rate)
					doWork(workTriggered, stopStageCh, workDone, options)
				} else {
					doWork := users.NewWorker(stage.usersConcurrency)
					doWork(workTriggered, stopStageCh, workDone, options)
				}
			}()

			// Wait until the current stage is completed or the program is stopped.
			// In any of the cases, it must wait for the worker to complete and avoid memory leak.
			// A stage needs to be stopped a bit earlier than the stage duration to avoid extra iterations.
			for isListening := true; isListening; {
				select {
				case <-stop:
					stopStageCh <- true
					wg.Wait()
					unsetEnvs(stage.params)
					stopStageTicker.Stop()
					return
				case <-stopStageTicker.C:
					select {
					case <-stop:
						continue
					default:
					}

					stopStageCh <- true
					wg.Wait()
					stopStageTicker.Stop()
					unsetEnvs(stage.params)
					time.Sleep(safeThresholdBeforeNextIteration)
					isListening = false
				default:
				}
			}
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
