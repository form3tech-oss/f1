package file

import (
	"context"
	"fmt"
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
			if ctx.Err() != nil || workers.MaxIterationsReached() {
				return
			}
			stageEnd = stageEnd.Add(stage.StageDuration)

			if timeLeft, skip := skipStage(stage, stageEnd); skip {
				output.Display(ui.WarningMessage{Message: fmt.Sprintf(
					"Skipping stage %d, as only %s of its %s is left",
					stage.Number, max(timeLeft, 0).Round(time.Millisecond), stage.StageDuration)})
				waitUntil(ctx, stageEnd)
				continue
			}
			runStage(ctx, output, workers, stage, stageEnd, options)
		}
	}
}

// skipStage returns how long a stage has left to run, until safeDurationBeforeNextStage
// before its scheduled end, and whether to skip it. A stage that starts late is skipped when
// it has less than half the time its first tick gets on schedule, as starting a whole tick
// of iterations in less time can drop most of them.
func skipStage(stage runnableStage, stageEnd time.Time) (time.Duration, bool) {
	timeLeft := time.Until(stageEnd) - safeDurationBeforeNextStage
	firstTick := min(stage.IterationDuration, stage.StageDuration) - safeDurationBeforeNextStage
	return timeLeft, timeLeft <= 0 || timeLeft < firstTick/2
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
	}

	// once the iteration limit stops the run, don't wait for the stage's scheduled end
	if workers.MaxIterationsReached() {
		return
	}

	waitUntil(ctx, stageEnd)
}

// waitUntil waits until t, unless ctx is done first
func waitUntil(ctx context.Context, t time.Time) {
	timer := time.NewTimer(time.Until(t))
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
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
