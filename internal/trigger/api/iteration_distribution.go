package api

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func NewDistribution(distributionTypeArg string, iterationDuration time.Duration, rateFn RateFunction) (time.Duration, RateFunction, error) {
	switch distributionTypeArg {
	case "none":
		return iterationDuration, rateFn, nil
	case "regular":
		distributedIterationDuration, distributedRateFn := withRegularDistribution(iterationDuration, rateFn)
		return distributedIterationDuration, distributedRateFn, nil
	case "random":
		randomFn := rand.Intn
		distributedIterationDuration, distributedRateFn := withRandomDistribution(iterationDuration, rateFn, randomFn)
		return distributedIterationDuration, distributedRateFn, nil
	default:
		return iterationDuration, rateFn, fmt.Errorf("unable to parse distribution %s", distributionTypeArg)
	}
}

func withRegularDistribution(iterationDuration time.Duration, rateFn RateFunction) (time.Duration, RateFunction) {
	distributedIterationDuration := 100 * time.Millisecond
	tolerance := 0.0000001

	if iterationDuration <= distributedIterationDuration {
		return iterationDuration, rateFn
	}

	rate := 0
	accRate := 0.0
	remainingSteps := 0
	tickSteps := int(iterationDuration.Milliseconds() / distributedIterationDuration.Milliseconds())

	distributedRateFn := func(time time.Time) int {
		if remainingSteps == 0 {
			rate = rateFn(time)
			accRate = 0.0
			remainingSteps = tickSteps
		}

		accRate += float64(rate) / float64(tickSteps)
		remainingSteps--

		if accRate < 1-tolerance {
			return 0
		}

		roundedAccRate := int(math.Round(accRate/tolerance) * tolerance)
		accRate -= float64(roundedAccRate)

		return roundedAccRate
	}

	return distributedIterationDuration, distributedRateFn
}

func withRandomDistribution(iterationDuration time.Duration, rateFn RateFunction, randFn func(int) int) (time.Duration, RateFunction) {
	distributedIterationDuration := 100 * time.Millisecond

	if iterationDuration <= distributedIterationDuration {
		return iterationDuration, rateFn
	}

	remainingSteps := 0
	remainingRate := 0
	tickSteps := int(iterationDuration.Milliseconds() / distributedIterationDuration.Milliseconds())

	distributedRateFn := func(time time.Time) int {
		if remainingSteps == 0 {
			remainingRate = rateFn(time)
			remainingSteps = tickSteps
		}

		var currentRate int
		if remainingSteps == 1 || remainingRate == 0 {
			currentRate = remainingRate
		} else {
			currentRate = randFn(remainingRate)

			if currentRate > remainingRate {
				currentRate = remainingRate
			}
		}
		remainingRate -= currentRate
		remainingSteps--

		if currentRate < 1 {
			return 0
		}

		return currentRate
	}

	return distributedIterationDuration, distributedRateFn
}
