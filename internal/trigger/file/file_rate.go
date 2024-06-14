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
	Scenario            string
	Stages              []runnableStage
	stagesTotalDuration time.Duration
	MaxDuration         time.Duration
	Concurrency         int
	MaxIterations       uint64
	maxFailures         uint64
	maxFailuresRate     int
	IgnoreDropped       bool
}

type runnableStage struct {
	Rate              api.RateFunction
	Params            map[string]string
	StageDuration     time.Duration
	IterationDuration time.Duration
	UsersConcurrency  int
}

func Rate(outputer ui.Outputer) api.Builder {
	flags := pflag.NewFlagSet("file", pflag.ContinueOnError)

	return api.Builder{
		Name:        "file <filename>",
		Description: "triggers test iterations from a yaml config file",
		Flags:       flags,
		New: func(flags *pflag.FlagSet) (*api.Trigger, error) {
			filename := flags.Arg(0)
			fileContent, err := readFile(filename, outputer)
			if err != nil {
				return nil, err
			}
			runnableStages, err := ParseConfigFile(*fileContent, time.Now())
			if err != nil {
				return nil, err
			}

			return &api.Trigger{
				Trigger:     newStagesWorker(runnableStages.Stages),
				DryRun:      newDryRun(runnableStages.Stages),
				Description: fmt.Sprintf("%d different stages", len(runnableStages.Stages)),
				Duration:    runnableStages.stagesTotalDuration,
				Options: api.Options{
					Scenario:        runnableStages.Scenario,
					MaxDuration:     runnableStages.MaxDuration,
					Concurrency:     runnableStages.Concurrency,
					MaxIterations:   runnableStages.MaxIterations,
					MaxFailures:     runnableStages.maxFailures,
					MaxFailuresRate: runnableStages.maxFailuresRate,
					IgnoreDropped:   runnableStages.IgnoreDropped,
				},
			}, nil
		},
		IgnoreCommonFlags: true,
	}
}

func readFile(filename string, outputer ui.Outputer) (*[]byte, error) {
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			outputer.Display(ui.ErrorMessage{
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

func newDryRun(stagesToRun []runnableStage) api.RateFunction {
	var startTime time.Time
	started := false
	stageIdx := 0

	return func(time time.Time) int {
		if stageIdx >= (len(stagesToRun)) {
			return 0
		}

		if !started {
			startTime = time
			started = true
		}

		currentStage := stagesToRun[stageIdx]

		if startTime.Add(currentStage.StageDuration).Before(time) {
			startTime = startTime.Add(currentStage.StageDuration)
			stageIdx++
		}

		if currentStage.UsersConcurrency > 0 {
			return 1
		}

		rate := currentStage.Rate(time)
		return rate
	}
}
