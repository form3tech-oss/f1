package file

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trace"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/users"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

const safeDurationBeforeNextStage = 20 * time.Millisecond

func newStagesWorker(stages []runnableStage, tracer trace.Tracer) api.WorkTriggerer {
	return func(ctx context.Context, workers *workers.PoolManager, options options.RunOptions) {
		for _, stage := range stages {
			if ctx.Err() != nil {
				return
			}
			runStage(ctx, workers, stage, options, tracer)
		}
	}
}

func runStage(
	ctx context.Context,
	workers *workers.PoolManager,
	stage runnableStage,
	options options.RunOptions,
	tracer trace.Tracer,
) {
	setEnvs(stage.params)
	defer unsetEnvs(stage.params)

	// stop the stage early to avoid starting a new tick
	stageCtx, stageCancel := context.WithTimeout(ctx, stage.stageDuration-safeDurationBeforeNextStage)
	defer stageCancel()

	stageDone := make(chan struct{})

	go func() {
		defer close(stageDone)

		if stage.usersConcurrency == 0 {
			doWork := api.NewIterationWorker(stage.iterationDuration, stage.rate, tracer)
			doWork(stageCtx, workers, options)
		} else {
			doWork := users.NewWorker(stage.usersConcurrency)
			doWork(stageCtx, workers, options)
		}
	}()

	select {
	case <-ctx.Done():
		<-stageDone
		return
	case <-stageDone:
		time.Sleep(safeDurationBeforeNextStage)
	}
}

func setEnvs(envs map[string]string) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		if err != nil {
			logrus.WithError(err).Error("unable set environment variables for given scenario")
		}
	}
}

func unsetEnvs(envs map[string]string) {
	for key := range envs {
		err := os.Unsetenv(key)
		if err != nil {
			logrus.WithError(err).Error("unable unset environment variables for given scenario")
		}
	}
}
