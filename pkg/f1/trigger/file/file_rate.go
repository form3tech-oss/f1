package file

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"
	"github.com/spf13/pflag"
)

type runnableStages struct {
	stages              []runnableStage
	stagesTotalDuration time.Duration
	maxDuration         time.Duration
	concurrency         int
	maxIterations       int32
}

type runnableStage struct {
	stageDuration      time.Duration
	iterationDuration  time.Duration
	rate               api.RateFunction
	iterationFrequency time.Duration
	users              int
	params             map[string]string
}

func FileRate() api.Builder {
	flags := pflag.NewFlagSet("file", pflag.ContinueOnError)
	flags.String("config-file", "config-file.yaml", "filename containing list of stages to run")

	return api.Builder{
		Name:        "file",
		Description: "triggers test iterations from a yaml config file",
		Flags:       flags,
		New: func(flags *pflag.FlagSet) (*api.Trigger, error) {
			filename, err := flags.GetString("config-file")
			if err != nil {
				return nil, err
			}
			fileContent, err := ioutil.ReadFile(filename)
			if err != nil {
				return nil, err
			}
			runnableStages, err := parseConfigFile(fileContent, time.Now())
			if err != nil {
				return nil, err
			}

			return &api.Trigger{
				Trigger:     newStagesWorker(runnableStages.stages),
				DryRun:      newDryRun(runnableStages.stages),
				Description: fmt.Sprintf("Running %d stages", len(runnableStages.stages)),
				Duration:    runnableStages.stagesTotalDuration,
				Options: api.Options{
					MaxDuration:   runnableStages.maxDuration,
					Concurrency:   runnableStages.concurrency,
					MaxIterations: runnableStages.maxIterations,
				},
			}, nil
		},
		IgnoreCommonFlags: true,
	}
}

func newDryRun(stagesToRun []runnableStage) api.RateFunction {
	var now time.Time
	started := false
	stageIdx := 0

	return func(time time.Time) int {
		if stageIdx >= (len(stagesToRun)) {
			return 0
		}

		if !started {
			now = time
			started = true
		}

		currentStage := stagesToRun[stageIdx]

		if now.Add(currentStage.stageDuration).Before(time) {
			now = now.Add(currentStage.stageDuration)
			stageIdx++
		}

		if currentStage.users > 0 {
			return 1
		}

		rate := currentStage.rate(now)
		return rate
	}
}
