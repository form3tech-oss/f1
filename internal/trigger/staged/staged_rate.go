package staged

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trace"
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
	flags.String(flagStartTime, "", "Starting point of stage calculation, defaults to now")

	triggerflags.JitterFlag(flags)
	triggerflags.DistributionFlag(flags)

	return api.Builder{
		Name:        "staged <scenario>",
		Description: "triggers iterations at varying rates",
		Flags:       flags,
		New: func(params *pflag.FlagSet, tracer trace.Tracer) (*api.Trigger, error) {
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
			if parsedStartTime, err := time.Parse("2006-01-02T15:04:05+07:00", startTimeStr); err == nil {
				startTime = &parsedStartTime
			}

			rates, err := CalculateStagedRate(jitterArg, frequency, stg, distributionTypeArg, startTime)
			if err != nil {
				return nil, err
			}

			return &api.Trigger{
					Trigger: api.NewIterationWorker(rates.IterationDuration, rates.Rate, tracer),
					DryRun:  rates.Rate,
					Description: fmt.Sprintf(
						"Starting iterations every %s in numbers varying by time: %s, using distribution %s",
						frequency, stg, distributionTypeArg),
					Duration: rates.Duration,
				},
				nil
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
	stages, err := parseStages(stg)
	if err != nil {
		return nil, fmt.Errorf("parsing stages: %w", err)
	}

	calculator := newRateCalculator(stages, startTime)
	rateFn := api.WithJitter(calculator.Rate, jitterArg)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(
		api.DistributionType(distributionTypeArg), frequency, rateFn,
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
