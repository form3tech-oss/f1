package gaussian

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/chobie/go-gaussian"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/rate"
)

const defaultVolume = 24 * 60 * 60

func GaussianRate() api.Builder {
	flags := pflag.NewFlagSet("gaussian", pflag.ContinueOnError)
	flags.Float64("volume", defaultVolume, "The desired volume to be achieved with the calculated load profile. Will be ignored if --peak-rate is also provided.")
	flags.Duration("repeat", 24*time.Hour, "How often the cycle should repeat")
	flags.Duration("iteration-frequency", 1*time.Second, "How frequently iterations should be started")
	flags.String("weights", "", "Optional scaling factor to apply per repetition. This can be used for example with daily repetitions to set different weights per day of the week")
	flags.Duration("peak", 14*time.Hour, "The offset within the repetition window when the load should reach its maximum. Default 14 hours (with 24 hour default repeat)")
	flags.StringP("peak-rate", "r", "", "number of iterations per interval in peak time, in the form <request>/<duration> (e.g. 1/s). If --peak-rate is provided, the value given for --volume will be ignored.")
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
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			repeat, err := flags.GetDuration("repeat")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			frequency, err := flags.GetDuration("iteration-frequency")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			weights, err := flags.GetString("weights")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			peak, err := flags.GetDuration("peak")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			stddev, err := flags.GetDuration("standard-deviation")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			jitter, err := flags.GetFloat64("jitter")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			distributionTypeArg, err := flags.GetString("distribution")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			peakRate, err := flags.GetString("peak-rate")
			if err != nil {
				return nil, fmt.Errorf("getting flag: %w", err)
			}
			jitterDesc := ""
			if jitter != 0 {
				jitterDesc = fmt.Sprintf(" with jitter of %.2f%%", jitter)
			}
			if peakRate != "" {
				if volume != defaultVolume {
					log.Warn("--peak-rate is provided, the value given for --volume will be ignored")
				}
				volume, err = calculateVolume(peakRate, peak, stddev)
				if err != nil {
					return nil, err
				}
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
		weight, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse weights")
		}
		weightsSlice = append(weightsSlice, weight)
	}

	calculator := NewGaussianRateCalculator(peak, stddev, frequency, weightsSlice, volume, repeat)

	rateFn := api.WithJitter(calculator.For, jitter)
	distributedIterationDuration, distributedRateFn, err := api.NewDistribution(distributionTypeArg, frequency, rateFn)
	if err != nil {
		return nil, fmt.Errorf("new distribution: %w", err)
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

func calculateVolume(peakTps string, peakTime, stddev time.Duration) (float64, error) {
	a, err := parseRateToTPS(peakTps) // the desired peak TPS
	if err != nil {
		return -1, err
	}
	b := peakTime.Seconds()
	c := stddev.Seconds()

	var total float64
	for i := 0; i < 3600*24; i++ {
		total += gauss(a, b, c, float64(i))
	}

	return math.Round(total), nil
}

func gauss(a, b, c, x float64) float64 {
	return a * math.Exp(-(math.Pow(x-b, 2) / (2 * c * c)))
}

func parseRateToTPS(rateArg string) (float64, error) {
	rate, unit, err := rate.ParseRate(rateArg)
	if err != nil {
		return -1, fmt.Errorf("parse to tps %s: %w", rateArg, err)
	}
	return float64(rate) / unit.Seconds(), nil
}
