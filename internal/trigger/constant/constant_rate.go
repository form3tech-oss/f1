package constant

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/rate"
)

func Rate() api.Builder {
	flags := pflag.NewFlagSet("constant", pflag.ContinueOnError)
	flags.StringP("rate", "r", "1/s", "number of iterations to start per interval, in the form <request>/<duration>")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")
	flags.String("distribution", "regular", "optional parameter to distribute the rate over steps of 100ms, which can be none|regular|random")

	return api.Builder{
		Name:        "constant <scenario>",
		Description: "triggers test iterations at a constant rate",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			rateArg, err := params.GetString("rate")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			jitterArg, err := params.GetFloat64("jitter")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			distributionTypeArg, err := params.GetString("distribution")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}

			rates, err := CalculateConstantRate(jitterArg, rateArg, distributionTypeArg)
			if err != nil {
				return nil, fmt.Errorf("calculating constant rate: %w", err)
			}

			return &api.Trigger{
					Trigger:     api.NewIterationWorker(rates.IterationDuration, rates.Rate),
					Description: fmt.Sprintf("%s constant rate, using distribution %s", rateArg, distributionTypeArg),
					DryRun:      rates.Rate,
				},
				nil
		},
	}
}

func CalculateConstantRate(jitterArg float64, rateArg, distributionTypeArg string) (*api.Rates, error) {
	rate, iterationDuration, err := rate.ParseRate(rateArg)
	if err != nil {
		return nil, fmt.Errorf("unable to parse rate %s: %w", rateArg, err)
	}

	rateFn := api.WithJitter(func(time.Time) int { return rate }, jitterArg)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(distributionTypeArg, iterationDuration, rateFn)
	if err != nil {
		return nil, fmt.Errorf("new distribution: %w", err)
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
	}, nil
}
