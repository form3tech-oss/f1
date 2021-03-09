package gaussian

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"

	"github.com/chobie/go-gaussian"
	"github.com/spf13/pflag"
)

func GaussianRate() api.Builder {
	flags := pflag.NewFlagSet("gaussian", pflag.ContinueOnError)
	flags.Float64("volume", 24*60*60, "The desired volume to be achieved with the calculated load profile.")
	flags.Duration("repeat", 24*time.Hour, "How often the cycle should repeat")
	flags.Duration("iteration-frequency", 1*time.Second, "How frequently iterations should be started")
	flags.String("weights", "", "Optional scaling factor to apply per repetition. This can be used for example with daily repetitions to set different weights per day of the week")
	flags.Duration("peak", 14*time.Hour, "The offset within the repetition window when the load should reach its maximum. Default 14 hours (with 24 hour default repeat)")
	flags.Duration("standard-deviation", 150*time.Minute, "The standard deviation to use for the distribution of load")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")
	flags.String("distribution", "regular", "optional parameter to distribute the rate over steps of 100ms, which can be none|regular|random")

	return api.Builder{
		Name:        "gaussian <scenario>",
		Description: "distributes load to match a desired monthly volume",
		Flags:       flags,
		New: func(flags *pflag.FlagSet) (*api.Trigger, error) {
			volume, err := flags.GetFloat64("volume")
			if err != nil {
				return nil, err
			}
			repeat, err := flags.GetDuration("repeat")
			if err != nil {
				return nil, err
			}
			frequency, err := flags.GetDuration("iteration-frequency")
			if err != nil {
				return nil, err
			}
			weights, err := flags.GetString("weights")
			if err != nil {
				return nil, err
			}
			peak, err := flags.GetDuration("peak")
			if err != nil {
				return nil, err
			}
			stddev, err := flags.GetDuration("standard-deviation")
			if err != nil {
				return nil, err
			}
			jitter, err := flags.GetFloat64("jitter")
			if err != nil {
				return nil, err
			}
			distributionTypeArg, err := flags.GetString("distribution")
			if err != nil {
				return nil, err
			}

			jitterDesc := ""
			if jitter != 0 {
				jitterDesc = fmt.Sprintf(" with jitter of %.2f%%", jitter)
			}

			rates, err := CalculateGaussianRate(volume, jitter, repeat, frequency, peak, stddev, weights, distributionTypeArg)
			if err != nil {
				return nil, err
			}

			return &api.Trigger{
					Trigger: api.NewIterationWorker(rates.IterationDuration, rates.Rate),
					DryRun:  rates.Rate,
					Description: fmt.Sprintf(
						"Gaussian distribution triggering %d iterations per %s, peaking at %s with standard deviation of %s%s, using distribution %s",
						int(volume),
						repeat,
						peak,
						stddev,
						jitterDesc,
						distributionTypeArg,
					),
					Duration: rates.Duration,
				},
				nil
		},
	}
}

type gaussianRateCalculator struct {
	repeatWindow  time.Duration
	frequency     time.Duration
	dist          *gaussian.Gaussian
	weights       []float64
	dailyVolume   float64
	remainder     float64
	multiplier    float64
	averageWeight float64
}

func CalculateGaussianRate(volume, jitter float64, repeat, frequency, peak, stddev time.Duration, weights, distributionTypeArg string) (*api.Rates, error) {
	var weightsSlice []float64
	for _, s := range strings.Split(weights, ",") {
		if s == "" {
			continue
		}
		weight, err := strconv.ParseFloat(s, 10)
		if err != nil {
			return nil, fmt.Errorf("unable to parse weights")
		}
		weightsSlice = append(weightsSlice, weight)
	}

	calculator := NewGaussianRateCalculator(peak, stddev, frequency, weightsSlice, volume, repeat)

	rateFn := api.WithJitter(calculator.For, jitter)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(distributionTypeArg, frequency, rateFn)
	if err != nil {
		return nil, err
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
		Duration:          time.Hour * 24 * 356,
	}, nil
}

func (c *gaussianRateCalculator) For(now time.Time) int {
	// this will be called every tick. Work out how many we should be sending now.
	start := now.Truncate(c.repeatWindow)
	slot := float64(now.Sub(start))
	instantRate := c.dist.Pdf(slot)

	rate := instantRate * c.multiplier

	if len(c.weights) > 0 {
		startOfWeight := now.Truncate(c.repeatWindow * time.Duration(len(c.weights)))
		i := 0
		for startOfWeight != start {
			i++
			startOfWeight = startOfWeight.Add(c.repeatWindow)
		}
		rate = rate * c.weights[i] / c.averageWeight
	}

	// the rate must be an integer. Save up any fractions and add them to the next iteration.
	rateWithRemainder := rate + c.remainder
	floorRate := math.Floor(rateWithRemainder)
	c.remainder = rateWithRemainder - floorRate
	return int(floorRate)
}

func NewGaussianRateCalculator(peak time.Duration, stddev time.Duration, frequency time.Duration, weights []float64, volume float64, repeatWindow time.Duration) *gaussianRateCalculator {
	variance := math.Pow(float64(stddev), 2)
	multiplier := volume * float64(frequency)
	gauss := gaussian.NewGaussian(float64(peak), variance)

	averageWeight := 1.0
	if len(weights) > 0 {
		totalWeight := 0.0
		for _, weight := range weights {
			totalWeight += weight
		}
		averageWeight = totalWeight / float64(len(weights))
	}

	// account for large standard deviations or peaks beyond the window
	coveredRegion := gauss.Cdf(float64(repeatWindow-frequency)) - gauss.Cdf(0)
	multiplier = multiplier / coveredRegion

	return &gaussianRateCalculator{
		frequency:     frequency,
		dist:          gauss,
		dailyVolume:   volume,
		weights:       weights,
		averageWeight: averageWeight,
		multiplier:    multiplier,
		repeatWindow:  repeatWindow,
	}
}
