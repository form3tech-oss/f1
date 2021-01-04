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

			if stage.users > 0 {
				for i := 0; i < stage.users; i++ {
					workTriggered <- true
				}

				totalDurationTicker := time.NewTicker(stage.stageDuration - 10*time.Millisecond)

				for isListening := true; isListening == true; {
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
			} else {
				startRate := stage.rate(time.Now())
				for i := 0; i < startRate; i++ {
					workTriggered <- true
				}

				// start ticker to trigger subsequent iterations.
				totalDurationTicker := time.NewTicker(stage.stageDuration - 10*time.Millisecond)
				iterationTicker := time.NewTicker(stage.iterationDuration)

				// run more iterations on every tick, until duration has elapsed.
				for isListening := true; isListening == true; {
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

			unsetEnvs(stage.params)
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
	for key, _ := range envs {
		err := os.Unsetenv(key)
		if err != nil {
			log.WithError(err).Error("unable unset environment variables for given scenario")
		}
	}
}
