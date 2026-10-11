package staged

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/triggerflags"
)

const (
	flagStages             = "stages"
	flagIterationFrequency = "iterationFrequency"
	flagStartTime          = "startTime"
)

func Rate() api.Builder {
	flags := pflag.NewFlagSet("staged", pflag.ContinueOnError)
	flags.StringP("stages", "s", "0s:1, 10s:1",
		"Comma separated list of <stage_duration>:<target_concurrent_iterations>. "+
			"During the stage, the number of concurrent iterations will ramp up or down to the target.")
	flags.DurationP(flagIterationFrequency, "f", 1*time.Second,
		"How frequently iterations should be started")
	flags.String(flagStartTime, "",
		"Starting point of stage calculation in RFC3339 format, defaults to now. "+
			"The run ends when the last stage does. A --max-duration counts from this time when it is in the future")

	triggerflags.JitterFlag(flags)
	triggerflags.DistributionFlag(flags)

	return api.Builder{
		Name:        "staged <scenario>",
		Description: "triggers iterations at varying rates",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			jitterArg, err := params.GetFloat64(triggerflags.FlagJitter)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			stg, err := params.GetString(flagStages)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			frequency, err := params.GetDuration(flagIterationFrequency)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			distributionTypeArg, err := params.GetString(triggerflags.FlagDistribution)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			var startTime *time.Time
			startTimeStr, err := params.GetString(flagStartTime)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			if startTimeStr != "" {
				parsedStartTime, err := time.Parse(time.RFC3339, startTimeStr)
				if err != nil {
					return nil, fmt.Errorf("parsing %s: %w", flagStartTime, err)
				}
				startTime = &parsedStartTime
			}

			rates, err := CalculateStagedRate(jitterArg, frequency, stg, distributionTypeArg, startTime)
			if err != nil {
				return nil, err
			}
			// The run's length comes from the stages, so with no duration it would fall back
			// to --max-duration and start no iterations. Config-file stages are not checked
			// here: they take their length from the file stage's own duration.
			if rates.Duration <= 0 {
				return nil, fmt.Errorf("stages `%s` must have a total duration above 0", stg)
			}

			// With a start time the run ends when the last stage does, so it waits for a start
			// time in the future and resumes mid-schedule from one in the past.
			trigger := &api.Trigger{
				Trigger: api.NewIterationWorker(rates.IterationDuration, rates.Rate),
				DryRun:  rates.Rate,
				Description: fmt.Sprintf(
					"Starting iterations every %s in numbers varying by time: %s, using distribution %s",
					frequency, stg, distributionTypeArg),
				Duration: rates.Duration,
			}
			if startTime != nil {
				trigger.Start = *startTime
			}

			return trigger, nil
		},
	}
}

func CalculateStagedRate(
	jitterArg float64,
	frequency time.Duration,
	stg string,
	distributionTypeArg string,
	startTime *time.Time,
) (*api.Rates, error) {
	stages, err := ParseStages(stg)
	if err != nil {
		return nil, fmt.Errorf("parsing stages: %w", err)
	}

	calculator := NewRateCalculator(stages, startTime)
	rateFn := api.WithJitter(calculator.Rate, jitterArg)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(
		api.DistributionType(distributionTypeArg), frequency, rateFn, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("new distribution: %w", err)
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
		Duration:          calculator.MaxDuration(),
	}, nil
}
