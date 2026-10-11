package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

type RunnableStages struct {
	Scenario                 string
	Stages                   []runnableStage
	stageStart               *time.Time
	totalDuration            time.Duration
	MaxDuration              time.Duration
	Concurrency              int
	MaxIterations            uint64
	maxFailures              uint64
	maxFailuresRate          int
	IgnoreDropped            bool
	WaitForCompletionTimeout time.Duration
}

type runnableStage struct {
	Rate              api.RateFunction
	Params            map[string]string
	StageDuration     time.Duration
	IterationDuration time.Duration
	UsersConcurrency  int
}

func Rate(output *ui.Output) api.Builder {
	flags := pflag.NewFlagSet("file", pflag.ContinueOnError)

	return api.Builder{
		Name:        "file <filename>",
		Description: "triggers test iterations from a yaml config file",
		Flags:       flags,
		New: func(flags *pflag.FlagSet) (*api.Trigger, error) {
			filename := flags.Arg(0)
			fileContent, err := readFile(filename, output)
			if err != nil {
				return nil, err
			}
			runnableStages, err := ParseConfigFile(*fileContent)
			if err != nil {
				return nil, err
			}

			var start time.Time
			if runnableStages.stageStart != nil {
				start = *runnableStages.stageStart
			}
			worker := newStagesWorker(runnableStages.Stages, start)

			return &api.Trigger{
				Trigger:     worker.run,
				Cleanup:     worker.cleanup,
				DryRun:      newDryRun(runnableStages.Stages, start),
				Description: fmt.Sprintf("%d different stages", len(runnableStages.Stages)),
				Start:       start,
				Duration:    runnableStages.totalDuration,
				Options: api.Options{
					Scenario:                 runnableStages.Scenario,
					MaxDuration:              runnableStages.MaxDuration,
					Concurrency:              runnableStages.Concurrency,
					MaxIterations:            runnableStages.MaxIterations,
					MaxFailures:              runnableStages.maxFailures,
					MaxFailuresRate:          runnableStages.maxFailuresRate,
					IgnoreDropped:            runnableStages.IgnoreDropped,
					WaitForCompletionTimeout: runnableStages.WaitForCompletionTimeout,
				},
			}, nil
		},
		IgnoreCommonFlags: true,
	}
}

func readFile(filename string, output *ui.Output) (*[]byte, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			output.Display(ui.ErrorMessage{
				Message: "unable to close the config file",
				Error:   err,
			})
		}
	}()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return &fileContent, nil
}

// newDryRun returns the rate of the stage running at each time, with the stages starting
// at start, or at the first time asked for if start is the zero time.
func newDryRun(stagesToRun []runnableStage, start time.Time) api.RateFunction {
	stageStart := start
	stageIdx := 0

	return func(now time.Time) int {
		if stageStart.IsZero() {
			stageStart = now
		}
		for stageIdx < len(stagesToRun) && !now.Before(stageStart.Add(stagesToRun[stageIdx].StageDuration)) {
			stageStart = stageStart.Add(stagesToRun[stageIdx].StageDuration)
			stageIdx++
		}
		if stageIdx >= len(stagesToRun) || now.Before(stageStart) {
			return 0
		}

		currentStage := stagesToRun[stageIdx]
		if currentStage.UsersConcurrency > 0 {
			return 1
		}

		return currentStage.Rate(now)
	}
}
