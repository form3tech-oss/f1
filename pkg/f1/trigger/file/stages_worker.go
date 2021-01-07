package file

import (
	"os"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trigger/users"

	"github.com/form3tech-oss/f1/pkg/f1/options"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"
	log "github.com/sirupsen/logrus"
)

func newStagesWorker(stages []runnableStage) api.WorkTriggerer {
	return func(workTriggered chan<- bool, stop <-chan bool, workDone <-chan bool, options options.RunOptions) {
		for _, stage := range stages {
			setEnvs(stage.params)

			stageDuration := stage.stageDuration - 10*time.Millisecond
			if stage.usersConcurrency == 0 {
				api.DoWork(workTriggered, stop, workDone, stage.iterationDuration, stageDuration, stage.rate)
			} else {
				users.DoWork(workTriggered, stop, workDone, stage.usersConcurrency, stageDuration)
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
	for key := range envs {
		err := os.Unsetenv(key)
		if err != nil {
			log.WithError(err).Error("unable unset environment variables for given scenario")
		}
	}
}
