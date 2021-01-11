package file

import (
	"os"
	"sync"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trigger/users"

	"github.com/form3tech-oss/f1/pkg/f1/options"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"
	log "github.com/sirupsen/logrus"
)

func newStagesWorker(stages []runnableStage) api.WorkTriggerer {
	return func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
		stopStage := make(chan bool)
		wg := sync.WaitGroup{}

		for _, stage := range stages {
			wg.Add(1)
			setEnvs(stage.params)

			totalDurationTicker := time.NewTicker(stage.stageDuration - 10*time.Millisecond)

			go func() {
				if stage.usersConcurrency == 0 {
					doWork := api.NewIterationWorker(stage.iterationDuration, stage.rate)
					doWork(workTriggered, stopStage, workDone, options)
				} else {
					doWork := users.NewWorker(stage.usersConcurrency)
					doWork(workTriggered, stopStage, workDone, options)
				}
				wg.Done()
			}()

			for isListening := true; isListening; {
				select {
				case <-stop:
					stopStage <- true
					wg.Wait()
					unsetEnvs(stage.params)
					totalDurationTicker.Stop()
					return
				case <-totalDurationTicker.C:
					select {
					case <-stop:
						continue
					default:
					}

					stopStage <- true
					wg.Wait()
					unsetEnvs(stage.params)
					totalDurationTicker.Stop()
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
