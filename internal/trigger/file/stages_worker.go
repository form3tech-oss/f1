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
		// schedule each stage from the start of the run, so a stage that overruns shortens
		// the next one instead of delaying all the stages after it
		stageEnd := time.Now()
		for _, stage := range stages {
			if ctx.Err() != nil {
				return
			}
			stageEnd = stageEnd.Add(stage.StageDuration)
			runStage(ctx, output, workers, stage, stageEnd, options)
		}
	}
}

func runStage(
	ctx context.Context,
	output *ui.Output,
	workers *workers.PoolManager,
	stage runnableStage,
	stageEnd time.Time,
	options options.RunOptions,
) {
	setEnvs(stage.Params, output)
	defer unsetEnvs(stage.Params, output)

	// stop the stage early to avoid starting a new tick
	stageCtx, stageCancel := context.WithDeadline(ctx, stageEnd.Add(-safeDurationBeforeNextStage))
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
		time.Sleep(time.Until(stageEnd))
	}
}

func setEnvs(envs map[string]string, output *ui.Output) {
	for key, value := range envs {
		err := os.Setenv(key, value)
		if err != nil {
			output.Display(ui.ErrorMessage{
				Message: "unable to set environment variables for given scenario",
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
				Message: "unable to unset environment variables for given scenario",
				Error:   err,
			})
		}
	}
}
