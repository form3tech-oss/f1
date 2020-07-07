package api

import (
	"math"
	"math/rand"
	"time"
)

func WithJitter(rate RateFunction, multiple float64) RateFunction {
	balance := 0.0
	if multiple == 0 {
		return rate
	}
	return func(now time.Time) int {
		variationFactor := 1 + (math.Cos(rand.Float64()*2*math.Pi))*multiple/100
		requestedRate := float64(rate(now)) + balance
		proposed := requestedRate * variationFactor
		rounded := math.Max(0, math.Round(proposed))
		balance = requestedRate - rounded
		return int(rounded)
	}
}
