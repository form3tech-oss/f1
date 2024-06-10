package file

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/users"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

const safeDurationBeforeNextStage = 20 * time.Millisecond

func newStagesWorker(stages []RunnableStage) api.WorkTriggerer {
	return func(ctx context.Context, workers *workers.PoolManager, options options.RunOptions, logger *slog.Logger) {
		for _, stage := range stages {
			if ctx.Err() != nil {
				return
			}
			runStage(ctx, workers, stage, options, logger)
		}
	}
}

func runStage(
	ctx context.Context,
	workers *workers.PoolManager,
	stage RunnableStage,
	options options.RunOptions,
	logger *slog.Logger,
) {
	setEnvs(stage.Params, logger)
	defer unsetEnvs(stage.Params, logger)

	// stop the stage early to avoid starting a new tick
	stageCtx, stageCancel := context.WithTimeout(ctx, stage.StageDuration-safeDurationBeforeNextStage)
	defer stageCancel()

	stageDone := make(chan struct{})

	go func() {
		defer close(stageDone)

		if stage.UsersConcurrency == 0 {
			doWork := api.NewIterationWorker(stage.IterationDuration, stage.Rate)
			doWork(stageCtx, workers, options, logger)
		} else {
			doWork := users.NewWorker(stage.UsersConcurrency)
			doWork(stageCtx, workers, options, logger)
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

func setEnvs(envs map[string]string, logger *slog.Logger) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		if err != nil {
			logger.Error("unable set environment variables for given scenario", log.ErrorAttr(err))
		}
	}
}

func unsetEnvs(envs map[string]string, logger *slog.Logger) {
	for key := range envs {
		err := os.Unsetenv(key)
		if err != nil {
			logger.Error("unable unset environment variables for given scenario", log.ErrorAttr(err))
		}
	}
}
