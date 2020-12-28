package staged

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"

	"github.com/spf13/pflag"
)

func StagedRate() api.Builder {
	flags := pflag.NewFlagSet("staged", pflag.ContinueOnError)
	flags.StringP("stages", "s", "0s:1, 10s:1", "Comma separated list of <stage_duration>:<target_concurrent_iterations>. During the stage, the number of concurrent iterations will ramp up or down to the target. ")
	flags.DurationP("iterationFrequency", "f", 1*time.Second, "How frequently iterations should be started")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")
	flags.String("distribution", "none", "optional parameter to distribute the rate over steps of 100ms, which can be none|regular|random")

	return api.Builder{
		Name:        "staged",
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

			stages, err := parseStages(stg)
			if err != nil {
				return nil, err
			}

			calculator := newRateCalculator(stages)
			rateFn := api.WithJitter(calculator.Rate, jitterArg)

			var distributedIterationDuration time.Duration
			var distributedRateFn func(time time.Time) int
			switch distributionTypeArg {
			case "none":
				distributedIterationDuration, distributedRateFn = frequency, rateFn
			case "regular":
				distributedIterationDuration, distributedRateFn = api.WithRegularDistribution(frequency, rateFn)
			case "random":
				randomFn := func(limit int) int { return rand.Intn(limit) }
				distributedIterationDuration, distributedRateFn = api.WithRandomDistribution(frequency, rateFn, randomFn)
			default:
				return nil, fmt.Errorf("unable to parse distribution %s", distributionTypeArg)
			}

			return &api.Trigger{
					Trigger:     api.NewIterationWorker(distributedIterationDuration, distributedRateFn),
					DryRun:      distributedRateFn,
					Description: fmt.Sprintf("Starting iterations every %s in numbers varying by time: %s,", distributedIterationDuration, stg),
					Duration:    calculator.MaxDuration(),
				},
				nil
		},
	}
}
