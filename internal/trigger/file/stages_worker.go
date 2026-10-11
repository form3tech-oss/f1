package file

import (
	"context"
	"os"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/options"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/ui"
	"github.com/form3tech-oss/f1/v2/internal/workers"
)

const safeDurationBeforeNextStage = 20 * time.Millisecond

// stagesWorker runs the stages of a config file in order. Stage parameters are passed
// to iterations as process environment variables.
type stagesWorker struct {
	// start is when the first stage starts, or the zero time to start it with the run.
	start  time.Time
	stages []runnableStage
	// output is the run's output, set by run.
	output *ui.Output
	// replacedEnv holds what the running stage's parameters replaced, or the last stage's
	// once it has run. cleanup restores it.
	replacedEnv previousEnv
}

func newStagesWorker(stages []runnableStage, start time.Time) *stagesWorker {
	return &stagesWorker{stages: stages, start: start}
}

// run runs each stage from its scheduled start to its scheduled end, skipping the stages
// that already ended. A late start therefore shortens a stage instead of moving the ones
// after it, so the run ends when the schedule does.
func (w *stagesWorker) run(
	ctx context.Context,
	output *ui.Output,
	workers *workers.PoolManager,
	options options.RunOptions,
) {
	w.output = output

	stageStart := w.start
	if stageStart.IsZero() {
		stageStart = time.Now()
	}
	for i, stage := range w.stages {
		if ctx.Err() != nil {
			return
		}
		stageEnd := stageStart.Add(stage.StageDuration)
		if !waitUntil(ctx, stageStart) {
			return
		}
		if w.start.IsZero() || stageEnd.After(time.Now()) {
			w.runStage(ctx, workers, stage, stageEnd, options, i == len(w.stages)-1)
		}
		stageStart = stageEnd
	}
}

// cleanup restores the environment the current stage replaced. The runner calls it after
// it stops waiting for in-flight iterations, so they keep their parameters.
func (w *stagesWorker) cleanup() {
	restoreEnvs(w.replacedEnv, w.output)
	w.replacedEnv = nil
}

func (w *stagesWorker) runStage(
	ctx context.Context,
	workers *workers.PoolManager,
	stage runnableStage,
	stageEnd time.Time,
	options options.RunOptions,
	last bool,
) {
	w.replacedEnv = setEnvs(stage.Params, w.output)

	// stop the stage early to avoid starting a new tick
	stageCtx, stageCancel := context.WithDeadline(ctx, stageEnd.Add(-safeDurationBeforeNextStage))
	defer stageCancel()

	if stage.UsersConcurrency == 0 {
		api.NewIterationWorker(stage.IterationDuration, stage.Rate)(stageCtx, w.output, workers, options)
	} else {
		workers.NewContinuousPool(stage.UsersConcurrency).Start(stageCtx)

		// The users stop when stageCtx ends. Let their in-flight iterations finish before
		// the next stage starts, unless the run ends. This wait is deliberately not bounded
		// by --wait-for-completion-timeout: it keeps the behaviour stages have always had,
		// where the next stage only starts once the previous users are done.
		workers.WaitForCompletionOrDone(ctx)
	}

	if ctx.Err() != nil {
		// The run is ending while iterations may still be in flight. Keep the parameters
		// set for them; the runner calls cleanup once it stops waiting.
		return
	}

	// Pause after every stage, the last one included: the trigger then returns once the
	// runner's max-duration deadline has passed, so the runner reports why the run ended.
	time.Sleep(safeDurationBeforeNextStage)

	if last {
		// Iterations may still be in flight. Keep the parameters set for them; the runner
		// calls cleanup once it stops waiting.
		return
	}

	// Known limitation: rate stages do not wait for in-flight iterations, so those can see
	// the next stage's parameters. Fixing this needs per-iteration parameters instead of
	// process environment variables, which is planned for v3.
	w.cleanup()
}

// waitUntil waits until t, and reports false if ctx ended first.
func waitUntil(ctx context.Context, t time.Time) bool {
	timer := time.NewTimer(time.Until(t))
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// previousEnv maps each variable a stage set to its value before the stage, or to nil
// if it wasn't set.
//
// Known limitation: environment variables are process-wide, so runs that overlap in one
// process and set the same parameter interfere with each other, and restoring can leave
// one run's value set. Per-iteration parameters in v3 remove this.
type previousEnv map[string]*string

func setEnvs(envs map[string]string, output *ui.Output) previousEnv {
	previous := make(previousEnv, len(envs))
	for key, value := range envs {
		if old, ok := os.LookupEnv(key); ok {
			previous[key] = &old
		} else {
			previous[key] = nil
		}

		err := os.Setenv(key, value)
		if err != nil {
			output.Display(ui.ErrorMessage{
				Message: "unable set environment variables for given scenario",
				Error:   err,
			})
		}
	}

	return previous
}

func restoreEnvs(previous previousEnv, output *ui.Output) {
	for key, value := range previous {
		var err error
		if value == nil {
			err = os.Unsetenv(key)
		} else {
			err = os.Setenv(key, *value)
		}
		if err != nil {
			output.Display(ui.ErrorMessage{
				Message: "unable restore environment variables for given scenario",
				Error:   err,
			})
		}
	}
}
