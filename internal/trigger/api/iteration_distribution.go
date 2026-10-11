package api

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type DistributionType string

const (
	NoneDistribution    DistributionType = "none"
	RegularDistribution DistributionType = "regular"
	RandomDistribution  DistributionType = "random"
)

// MinDistributionStep is the shortest step the regular and random distributions split a
// unit into. Units shorter than two steps aren't split.
const MinDistributionStep = 100 * time.Millisecond

func NewDistribution(
	distributionTypeArg DistributionType,
	iterationDuration time.Duration,
	rateFn RateFunction,
	randomFnArg func(int) int,
) (time.Duration, RateFunction, error) {
	// The constant, ramp, staged and gaussian triggers, and file stages that use them, build
	// their distribution here. A zero interval would otherwise panic in time.NewTicker when
	// the run starts.
	if iterationDuration <= 0 {
		return iterationDuration, rateFn, fmt.Errorf("iteration frequency %s must be positive", iterationDuration)
	}

	randomFn := randomFnArg
	if randomFn == nil {
		randomFn = rand.Intn
	}

	switch distributionTypeArg {
	case NoneDistribution:
		return iterationDuration, rateFn, nil
	case RegularDistribution, RandomDistribution:
	default:
		return iterationDuration, rateFn, fmt.Errorf("unable to parse distribution %s", distributionTypeArg)
	}

	tickSteps, distributedIterationDuration := distributionSteps(iterationDuration)
	if tickSteps == 1 {
		return iterationDuration, rateFn, nil
	}

	if distributionTypeArg == RegularDistribution {
		return distributedIterationDuration, withRegularDistribution(tickSteps, rateFn), nil
	}
	return distributedIterationDuration, withRandomDistribution(tickSteps, rateFn, randomFn), nil
}

// distributionSteps splits iterationDuration into as many steps of at least
// MinDistributionStep as fit, spread evenly over the whole duration. Durations that aren't
// a multiple of MinDistributionStep then keep their rate: 215ms is 2 steps of 107.5ms, not
// 2 steps of 100ms. Durations shorter than two steps are a single step.
func distributionSteps(iterationDuration time.Duration) (int, time.Duration) {
	steps := int64(iterationDuration / MinDistributionStep)
	if steps <= 1 {
		return 1, iterationDuration
	}
	return int(steps), iterationDuration / time.Duration(steps)
}

func withRegularDistribution(tickSteps int, rateFn RateFunction) RateFunction {
	rate := 0
	accRate := 0.0
	remainingSteps := 0

	distributedRateFn := func(time time.Time) int {
		if remainingSteps == 0 {
			rate = rateFn(time)
			accRate = 0.0
			remainingSteps = tickSteps
		}

		accRate += float64(rate) / float64(tickSteps)
		accRate = math.Ceil(accRate*10_000_000) / 10_000_000
		remainingSteps--

		if accRate < 1 {
			return 0
		}

		roundedAccRate := int(accRate)
		accRate -= float64(roundedAccRate)

		return roundedAccRate
	}

	return distributedRateFn
}

func withRandomDistribution(tickSteps int, rateFn RateFunction, randFn func(int) int) RateFunction {
	remainingSteps := 0
	remainingRate := 0

	distributedRateFn := func(time time.Time) int {
		if remainingSteps == 0 {
			remainingRate = rateFn(time)
			remainingSteps = tickSteps
		}

		var currentRate int
		if remainingSteps == 1 || remainingRate == 0 {
			currentRate = remainingRate
		} else {
			currentRate = min(randFn(remainingRate), remainingRate)
		}
		remainingRate -= currentRate
		remainingSteps--

		if currentRate < 1 {
			return 0
		}

		return currentRate
	}

	return distributedRateFn
}
