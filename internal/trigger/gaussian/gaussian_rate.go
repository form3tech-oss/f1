package gaussian

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/gaussian"
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/rate"
	"github.com/form3tech-oss/f1/v2/internal/triggerflags"
)

const defaultVolume = 24 * 60 * 60

const (
	flagVolume             = "volume"
	flagRepeat             = "repeat"
	flagIterationFrequency = "iteration-frequency"
	flagWeights            = "weights"
	flagPeak               = "peak"
	flagPeakRate           = "peak-rate"
	flagStandardDeviation  = "standard-deviation"
)

func Rate() api.Builder {
	flags := pflag.NewFlagSet("gaussian", pflag.ContinueOnError)
	flags.Float64(flagVolume, defaultVolume,
		"The desired volume to be achieved with the calculated load profile. "+
			"Will be ignored if --peak-rate is also provided.")
	flags.Duration(flagRepeat, 24*time.Hour,
		"How often the cycle should repeat")
	flags.Duration(flagIterationFrequency, 1*time.Second,
		"How frequently iterations should be started")
	flags.String(flagWeights, "",
		"Optional scaling factor to apply per repetition. "+
			"This can be used for example with daily repetitions to set different weights per day of the week")
	flags.Duration(flagPeak, 14*time.Hour,
		"The offset within the repetition window when the load should reach its maximum. "+
			"Default 14 hours (with 24 hour default repeat)")
	flags.StringP(flagPeakRate, "r", "",
		"number of iterations per interval in peak time, "+
			"in the form <request>/<duration> (e.g. 1/s). If --peak-rate is provided, "+
			"the value given for --volume will be ignored.")
	flags.Duration(flagStandardDeviation, 150*time.Minute,
		"The standard deviation to use for the distribution of load")

	triggerflags.JitterFlag(flags)
	triggerflags.DistributionFlag(flags)

	return api.Builder{
		Name:        "gaussian <scenario>",
		Description: "distributes load to match a desired monthly volume",
		Flags:       flags,
		New: func(flags *pflag.FlagSet) (*api.Trigger, error) {
			volume, err := flags.GetFloat64(flagVolume)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			repeat, err := flags.GetDuration(flagRepeat)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			frequency, err := flags.GetDuration("iteration-frequency")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			weights, err := flags.GetString(flagWeights)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			peakDuration, err := flags.GetDuration(flagPeak)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			stddevDuration, err := flags.GetDuration(flagStandardDeviation)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			jitter, err := flags.GetFloat64(triggerflags.FlagJitter)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			distributionTypeArg, err := flags.GetString(triggerflags.FlagDistribution)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			peakRate, err := flags.GetString(flagPeakRate)
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			if peakRate != "" {
				if volume != defaultVolume {
					logrus.Warn("--peak-rate is provided, the value given for --volume will be ignored")
				}
				volume, err = CalculateVolume(peakRate, peakDuration, stddevDuration)
				if err != nil {
					return nil, err
				}
			}

			rates, err := CalculateGaussianRate(
				volume,
				jitter,
				repeat,
				frequency,
				peakDuration,
				stddevDuration,
				weights,
				distributionTypeArg,
			)
			if err != nil {
				return nil, err
			}

			jitterDesc := ""
			if jitter != 0 {
				jitterDesc = fmt.Sprintf(" with jitter of %.2f%%", jitter)
			}

			description := fmt.Sprintf(
				"Gaussian distribution triggering %d iterations per %s, "+
					"peaking at %s with standard deviation of %s%s, using distribution %s",
				int(volume),
				repeat,
				peakDuration,
				stddevDuration,
				jitterDesc,
				distributionTypeArg,
			)

			return &api.Trigger{
					Trigger:     api.NewIterationWorker(rates.IterationDuration, rates.Rate),
					DryRun:      rates.Rate,
					Description: description,
					Duration:    rates.Duration,
				},
				nil
		},
	}
}

type Calculator struct {
	dist          *gaussian.Distribution
	weights       []float64
	repeatWindow  time.Duration
	frequency     time.Duration
	remainder     float64
	multiplier    float64
	averageWeight float64
}

func CalculateGaussianRate(
	volume, jitter float64,
	repeat, frequency, peak, stddev time.Duration,
	weightsArg, distributionTypeArg string,
) (*api.Rates, error) {
	weights := strings.Split(weightsArg, ",")
	weightsSlice := make([]float64, 0, len(weights))

	for _, s := range weights {
		if s == "" {
			continue
		}
		weight, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse weights: %w", err)
		}
		weightsSlice = append(weightsSlice, weight)
	}

	calculator, err := NewCalculator(peak, stddev, frequency, weightsSlice, volume, repeat)
	if err != nil {
		return nil, fmt.Errorf("calculator: %w", err)
	}

	rateFn := api.WithJitter(calculator.For, jitter)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(
		api.DistributionType(distributionTypeArg), frequency, rateFn, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("new distribution: %w", err)
	}

	return &api.Rates{
		IterationDuration: distributedIterationDuration,
		Rate:              distributedRateFn,
		Duration:          time.Hour * 24 * 356,
	}, nil
}

func (c *Calculator) For(now time.Time) int {
	// this will be called every tick. Work out how many we should be sending now.
	start := now.Truncate(c.repeatWindow)
	slot := float64(now.Sub(start))
	instantRate := c.dist.PDF(slot)

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

func NewCalculator(
	peak time.Duration,
	stddev time.Duration,
	frequency time.Duration,
	weights []float64,
	volume float64,
	repeatWindow time.Duration,
) (*Calculator, error) {
	multiplier := volume * float64(frequency)
	gauss, err := gaussian.NewDistribution(float64(peak), float64(stddev))
	if err != nil {
		return nil, fmt.Errorf("gaussian: %w", err)
	}

	averageWeight := 1.0
	if len(weights) > 0 {
		totalWeight := 0.0
		for _, weight := range weights {
			totalWeight += weight
		}
		averageWeight = totalWeight / float64(len(weights))
	}

	// account for large standard deviations or peaks beyond the window
	coveredRegion := gauss.CDF(float64(repeatWindow-frequency)) - gauss.CDF(0)
	multiplier /= coveredRegion

	return &Calculator{
		frequency:     frequency,
		dist:          gauss,
		weights:       weights,
		averageWeight: averageWeight,
		multiplier:    multiplier,
		repeatWindow:  repeatWindow,
	}, nil
}

func CalculateVolume(peakTps string, peakTime, stddev time.Duration) (float64, error) {
	amplitude, err := parseRateToTPS(peakTps) // the desired peak TPS
	if err != nil {
		return -1, err
	}
	mean := peakTime.Seconds()
	standardDeviation := stddev.Seconds()

	dist, err := gaussian.NewDistribution(mean, standardDeviation)
	if err != nil {
		return 0.0, fmt.Errorf("distribution: %w", err)
	}

	secondsInADay := 60 * 60 * 24

	var total float64
	for x := range secondsInADay {
		total += dist.Exponent(float64(x))
	}

	return math.Round(amplitude * total), nil
}

func parseRateToTPS(rateArg string) (float64, error) {
	rate, unit, err := rate.ParseRate(rateArg)
	if err != nil {
		return -1, fmt.Errorf("parse to tps %s: %w", rateArg, err)
	}
	return float64(rate) / unit.Seconds(), nil
}
