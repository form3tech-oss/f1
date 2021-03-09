package staged

import (
	"fmt"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"

	"github.com/spf13/pflag"
)

func StagedRate() api.Builder {
	flags := pflag.NewFlagSet("staged", pflag.ContinueOnError)
	flags.StringP("stages", "s", "0s:1, 10s:1", "Comma separated list of <stage_duration>:<target_concurrent_iterations>. During the stage, the number of concurrent iterations will ramp up or down to the target. ")
	flags.DurationP("iterationFrequency", "f", 1*time.Second, "How frequently iterations should be started")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")
	flags.String("distribution", "regular", "optional parameter to distribute the rate over steps of 100ms, which can be none|regular|random")

	return api.Builder{
		Name:        "staged <scenario>",
		Description: "triggers iterations at varying rates",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {

			jitterArg, err := params.GetFloat64("jitter")
			if err != nil {
				return nil, err
			}
			stg, err := params.GetString("stages")
			if err != nil {
				return nil, err
			}
			frequency, err := params.GetDuration("iterationFrequency")
			if err != nil {
				return nil, err
			}
			distributionTypeArg, err := params.GetString("distribution")
			if err != nil {
				return nil, err
			}

			rates, err := CalculateStagedRate(jitterArg, frequency, stg, distributionTypeArg)
			if err != nil {
				return nil, err
			}

			return &api.Trigger{
					Trigger:     api.NewIterationWorker(rates.IterationDuration, rates.Rate),
					DryRun:      rates.Rate,
					Description: fmt.Sprintf("Starting iterations every %s in numbers varying by time: %s, using distribution %s", frequency, stg, distributionTypeArg),
					Duration:    rates.Duration,
				},
				nil
		},
	}
}

func CalculateStagedRate(jitterArg float64, frequency time.Duration, stg, distributionTypeArg string) (*api.Rates, error) {
	stages, err := parseStages(stg)
	if err != nil {
		return nil, err
	}

	calculator := newRateCalculator(stages)
	rateFn := api.WithJitter(calculator.Rate, jitterArg)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(distributionTypeArg, frequency, rateFn)
	if err != nil {
		return nil, err
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
		Duration:          calculator.MaxDuration(),
	}, nil
}
