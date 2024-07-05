package file

import (
	"context"
	"os"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/users"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

const safeDurationBeforeNextStage = 20 * time.Millisecond

func newStagesWorker(stages []runnableStage) api.WorkTriggerer {
	return func(ctx context.Context, output *ui.Output, workers *workers.PoolManager, options options.RunOptions) {
		for _, stage := range stages {
			if ctx.Err() != nil {
				return
			}
			runStage(ctx, output, workers, stage, options)
		}
	}
}

func runStage(
	ctx context.Context,
	output *ui.Output,
	workers *workers.PoolManager,
	stage runnableStage,
	options options.RunOptions,
) {
	setEnvs(stage.Params, output)
	defer unsetEnvs(stage.Params, output)

	// stop the stage early to avoid starting a new tick
	stageCtx, stageCancel := context.WithTimeout(ctx, stage.StageDuration-safeDurationBeforeNextStage)
	defer stageCancel()

	stageDone := make(chan struct{})

	go func() {
		defer close(stageDone)

		if stage.UsersConcurrency == 0 {
			doWork := api.NewIterationWorker(stage.IterationDuration, stage.Rate)
			doWork(stageCtx, output, workers, options)
		} else {
			doWork := users.NewWorker(stage.UsersConcurrency)
			doWork(stageCtx, output, workers, options)
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

func setEnvs(envs map[string]string, output *ui.Output) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		if err != nil {
			output.Display(ui.ErrorMessage{
				Message: "unable set environment variables for given scenario",
				Error:   err,
			})
		}
	}
}

func unsetEnvs(envs map[string]string, output *ui.Output) {
	for key := range envs {
		err := os.Unsetenv(key)
		if err != nil {
			output.Display(ui.ErrorMessage{
				Message: "unable unset environment variables for given scenario",
				Error:   err,
			})
		}
	}
}
