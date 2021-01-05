package rampup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRampUpRate(t *testing.T) {
	for _, test := range []struct {
		testName, startRate, endRate        string
		duration, expectedIterationDuration time.Duration
		expectedRates                       []int
	}{
		{
			testName:                  "ramp up rate",
			startRate:                 "0/s",
			endRate:                   "100/s",
			duration:                  10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "ramp down rate",
			startRate:                 "100/s",
			endRate:                   "0/s",
			duration:                  10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10},
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, "2020-12-10T10:00:00+00:00")

			iterationDuration, rateFn, err := CalculateRampUpRate(test.startRate, test.endRate, test.duration)

			require.NoError(t, err)
			require.Equal(t, test.expectedIterationDuration, iterationDuration)
			var rates []int
			for range test.expectedRates {
				now = now.Add(test.expectedIterationDuration)
				rate := rateFn(now)
				rates = append(rates, rate)
			}
			require.Equal(t, test.expectedRates, rates)
		})
	}
}
