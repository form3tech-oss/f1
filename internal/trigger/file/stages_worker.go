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

// stagesWorker runs the stages of a config file in order. Stage parameters are passed
// to iterations as process environment variables.
type stagesWorker struct {
	stages []runnableStage
	// output is the run's output, set by run.
	output *ui.Output
	// activeParams are the parameters of the stage that is running, or of the last stage
	// if the run ended during it. They are unset by cleanup.
	activeParams map[string]string
}

func newStagesWorker(stages []runnableStage) *stagesWorker {
	return &stagesWorker{stages: stages}
}

func (w *stagesWorker) run(
	ctx context.Context,
	output *ui.Output,
	workers *workers.PoolManager,
	options options.RunOptions,
) {
	w.output = output
	for _, stage := range w.stages {
		if ctx.Err() != nil {
			return
		}
		w.runStage(ctx, workers, stage, options)
	}
}

// cleanup unsets the parameters left set when the run ended during a stage. The runner
// calls it after it stops waiting for in-flight iterations, so they keep their parameters.
func (w *stagesWorker) cleanup() {
	unsetEnvs(w.activeParams, w.output)
	w.activeParams = nil
}

func (w *stagesWorker) runStage(
	ctx context.Context,
	workers *workers.PoolManager,
	stage runnableStage,
	options options.RunOptions,
) {
	setEnvs(stage.Params, w.output)
	w.activeParams = stage.Params

	// stop the stage early to avoid starting a new tick
	stageCtx, stageCancel := context.WithTimeout(ctx, stage.StageDuration-safeDurationBeforeNextStage)
	defer stageCancel()

	if stage.UsersConcurrency == 0 {
		api.NewIterationWorker(stage.IterationDuration, stage.Rate)(stageCtx, w.output, workers, options)
	} else {
		users.StartPool(stageCtx, workers, stage.UsersConcurrency)

		// The users stop when stageCtx ends. Let their in-flight iterations finish before
		// the next stage starts, unless the run ends. This wait is deliberately not bounded
		// by --wait-for-completion-timeout: it keeps the behaviour stages have always had,
		// where the next stage only starts once the previous users are done.
		select {
		case <-ctx.Done():
		case <-workers.WaitForCompletion():
		}
	}

	if ctx.Err() != nil {
		// The run is ending while iterations may still be in flight. Keep the parameters
		// set for them; cleanup unsets them once the runner stops waiting.
		return
	}

	// Known limitation: rate stages do not wait for in-flight iterations, so those can see
	// the next stage's parameters. Fixing this needs per-iteration parameters instead of
	// process environment variables, which is planned for v3.
	unsetEnvs(stage.Params, w.output)
	w.activeParams = nil
	time.Sleep(safeDurationBeforeNextStage)
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
