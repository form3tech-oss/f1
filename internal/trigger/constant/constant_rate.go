package constant

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v3/internal/trigger/api"
	"github.com/form3tech-oss/f1/v3/internal/trigger/rate"
	"github.com/form3tech-oss/f1/v3/internal/triggerflags"
)

const (
	flagRate = "rate"
)

func Rate() api.Builder {
	flags := pflag.NewFlagSet("constant", pflag.ContinueOnError)
	flags.StringP(flagRate, "r", "1/s",
		"iterations per interval, e.g. 10/s, 100/m")

	triggerflags.JitterFlag(flags)
	triggerflags.DistributionFlag(flags)

	return api.Builder{
		Name:        "constant <scenario>",
		Description: "triggers test iterations at a constant rate",
		Long:        "Short flags: -r rate, -j jitter",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {
			rateArg, err := params.GetString(flagRate)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			jitterArg, err := params.GetFloat64(triggerflags.FlagJitter)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			distributionTypeArg, err := params.GetString(triggerflags.FlagDistribution)
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
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(
		api.DistributionType(distributionTypeArg), iterationDuration, rateFn, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("new distribution: %w", err)
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
	}, nil
}
