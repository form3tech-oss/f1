package ramp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"
	"github.com/spf13/pflag"
)

func RampRate() api.Builder {
	flags := pflag.NewFlagSet("ramp", pflag.ContinueOnError)
	flags.String("start-rate", "1/s", "number of iterations to start per interval, in the form <request>/<duration>")
	flags.String("end-rate", "1/s", "number of iterations to end per interval, in the form <request>/<duration>")
	flags.Duration("duration", 1*time.Second, "ramp duration")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")
	flags.String("distribution", "regular", "optional parameter to distribute the rate over steps of 100ms, which can be none|regular|random")

	return api.Builder{
		Name:        "ramp <scenario>",
		Description: "ramp up or down requests on every iteration",
		Flags:       flags,
		New: func(flags *pflag.FlagSet) (*api.Trigger, error) {
			startRateArg, err := flags.GetString("start-rate")
			if err != nil {
				return nil, err
			}
			endRateArg, err := flags.GetString("end-rate")
			if err != nil {
				return nil, err
			}
			duration, err := flags.GetDuration("duration")
			if err != nil {
				return nil, err
			}
			jitterArg, err := flags.GetFloat64("jitter")
			if err != nil {
				return nil, err
			}
			distributionTypeArg, err := flags.GetString("distribution")
			if err != nil {
				return nil, err
			}

			rates, _ := CalculateRampRate(startRateArg, endRateArg, distributionTypeArg, duration, jitterArg)

			return &api.Trigger{
				Trigger:     api.NewIterationWorker(rates.DistributedIterationDuration, rates.DistributedRate),
				Description: fmt.Sprintf("from %s to %s in %v, using distribution %s", startRateArg, endRateArg, duration, distributionTypeArg),
				DryRun:      rates.Rate,
			}, nil
		},
	}
}

func CalculateRampRate(startRateArg, endRateArg, distributionTypeArg string, duration time.Duration, jitterArg float64) (*api.Rates, error) {
	var startTime *time.Time

	startRate, startUnit, err := parseRateArg(startRateArg)
	if err != nil {
		return nil, err
	}
	endRate, endUnit, err := parseRateArg(endRateArg)
	if err != nil {
		return nil, err
	}

	if *startUnit != *endUnit {
		return nil, fmt.Errorf("start-rate and end-rate are not using the same unit")
	}
	if duration < *startUnit {
		return nil, fmt.Errorf("duration is lower than rate unit")
	}

	rateFn := func(now time.Time) int {
		if startTime == nil {
			startTime = &now
		}

		offset := now.Sub(*startTime)
		position := float64(offset) / float64(duration)
		rate := *startRate + int(position*float64(*endRate-*startRate))

		return rate
	}

	jitterRateFn := api.WithJitter(rateFn, jitterArg)
	distributedIterationDuration, distributedRateFn, _ := api.NewDistribution(distributionTypeArg, *startUnit, jitterRateFn)

	return &api.Rates{
		IterationDuration:            *startUnit,
		DistributedIterationDuration: distributedIterationDuration,
		Rate:                         rateFn,
		DistributedRate:              distributedRateFn,
		Duration:                     duration,
	}, nil
}

func parseRateArg(rateArg string) (*int, *time.Duration, error) {
	if strings.Contains(rateArg, "/") {
		rate, err := strconv.Atoi((rateArg)[0:strings.Index(rateArg, "/")])
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse rate arg %s", rateArg)
		}
		startUnitArg := (rateArg)[strings.Index(rateArg, "/")+1:]
		if !govalidator.IsNumeric(startUnitArg[0:1]) {
			startUnitArg = "1" + startUnitArg
		}
		unit, err := time.ParseDuration(startUnitArg)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse rate arg %s", rateArg)
		}

		return &rate, &unit, nil
	} else {
		rate, err := strconv.Atoi(rateArg)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse rate arg %s", rateArg)
		}
		unit := 1 * time.Second

		return &rate, &unit, nil
	}
}
