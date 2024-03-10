package ramp

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trace"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/rate"
)

func Rate() api.Builder {
	flags := pflag.NewFlagSet("ramp", pflag.ContinueOnError)
	flags.StringP("start-rate", "s", "1/s", "number of iterations to start per interval, in the form <request>/<duration>")
	flags.StringP("end-rate", "e", "1/s", "number of iterations to end per interval, in the form <request>/<duration>")
	flags.DurationP("ramp-duration", "r", 1*time.Second, "ramp duration, if not provided then --max-duration will be used")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")
	flags.String("distribution", "regular", "optional parameter to distribute the rate over steps of 100ms, which can be none|regular|random")

	return api.Builder{
		Name:        "ramp <scenario>",
		Description: "ramp up or down requests for a certain duration",
		Flags:       flags,
		New: func(flags *pflag.FlagSet, tracer trace.Tracer) (*api.Trigger, error) {
			startRateArg, err := flags.GetString("start-rate")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			endRateArg, err := flags.GetString("end-rate")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			duration, err := flags.GetDuration("ramp-duration")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			if duration == 0 {
				duration, err = flags.GetDuration("max-duration")
				if err != nil {
					return nil, fmt.Errorf("getting flag: %w", err)
				}
			}
			jitterArg, err := flags.GetFloat64("jitter")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			distributionTypeArg, err := flags.GetString("distribution")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}

			rates, err := CalculateRampRate(startRateArg, endRateArg, distributionTypeArg, duration, jitterArg)
			if err != nil {
				return nil, fmt.Errorf("calculating ramp rate: %w", err)
			}

			return &api.Trigger{
				Trigger:     api.NewIterationWorker(rates.IterationDuration, rates.Rate, tracer),
				Description: fmt.Sprintf("starting iterations from %s to %s during %v, using distribution %s", startRateArg, endRateArg, duration, distributionTypeArg),
				DryRun:      rates.Rate,
			}, nil
		},
	}
}

func CalculateRampRate(startRateArg, endRateArg, distributionTypeArg string, duration time.Duration, jitterArg float64) (*api.Rates, error) {
	var startTime *time.Time

	startRate, startUnit, err := rate.ParseRate(startRateArg)
	if err != nil {
		return nil, fmt.Errorf("parsing start rate: %w", err)
	}
	endRate, endUnit, err := rate.ParseRate(endRateArg)
	if err != nil {
		return nil, fmt.Errorf("parsing end rate: %w", err)
	}

	if startRate == endRate {
		return nil, errors.New("start-rate and end-rate should be different, for constant rate try using the constant mode")
	}
	if startUnit != endUnit {
		return nil, errors.New("start-rate and end-rate are not using the same unit")
	}
	if duration < startUnit {
		return nil, errors.New("duration is lower than rate unit")
	}

	rateFn := func(now time.Time) int {
		if startTime == nil {
			startTime = &now
		}

		if startTime.Add(duration).Before(now) {
			return 0
		}

		offset := now.Sub(*startTime)
		position := float64(offset) / float64(duration)
		rate := startRate + int(position*float64(endRate-startRate))

		return rate
	}

	jitterRateFn := api.WithJitter(rateFn, jitterArg)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(distributionTypeArg, startUnit, jitterRateFn)
	if err != nil {
		return nil, fmt.Errorf("new distribution: %w", err)
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
		Duration:          duration,
	}, nil
}
