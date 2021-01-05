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
			testName:                  "constant rate",
			startRate:                 "10/s",
			endRate:                   "10/s",
			duration:                  10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		},
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
		{
			testName:                  "ramp up rate using ms",
			startRate:                 "0/100ms",
			endRate:                   "100/100ms",
			duration:                  1 * time.Second,
			expectedIterationDuration: 100 * time.Millisecond,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "ramp up rate using minutes",
			startRate:                 "0/m",
			endRate:                   "60/m",
			duration:                  10 * time.Minute,
			expectedIterationDuration: 1 * time.Minute,
			expectedRates:             []int{0, 6, 12, 18, 24, 30, 36, 42, 48, 54},
		},
		{
			testName:                  "ramp up rate using hours",
			startRate:                 "0/h",
			endRate:                   "100/h",
			duration:                  10 * time.Hour,
			expectedIterationDuration: 1 * time.Hour,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "ramp up rate using multiple of durations",
			startRate:                 "0/90s",
			endRate:                   "100/90s",
			duration:                  15 * time.Minute,
			expectedIterationDuration: 90 * time.Second,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "use 1s as default when unit is not provided",
			startRate:                 "0",
			endRate:                   "100",
			duration:                  10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "duration is the same as start/end rate unit",
			startRate:                 "0/s",
			endRate:                   "10/s",
			duration:                  1 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0},
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, "2020-12-10T10:00:00+00:00")

			iterationDuration, rateFn, err := CalculateRampUpRate(test.startRate, test.endRate, test.duration)

			require.NoError(t, err)
			require.Equal(t, test.expectedIterationDuration, *iterationDuration)
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

func TestRampUpRate_Errors(t *testing.T) {
	for _, test := range []struct {
		startRate, endRate string
		duration           time.Duration
		expectedError      string
	}{
		{
			startRate:     "error",
			endRate:       "10/s",
			duration:      10 * time.Second,
			expectedError: "unable to parse rate arg error",
		},
		{
			startRate:     "10/s",
			endRate:       "error",
			duration:      10 * time.Second,
			expectedError: "unable to parse rate arg error",
		},
		{
			startRate:     "10/s",
			endRate:       "1/100ms",
			duration:      10 * time.Second,
			expectedError: "start-rate and end-rate are not using the same unit",
		},
		{
			startRate:     "10/s",
			endRate:       "100/s",
			duration:      100 * time.Millisecond,
			expectedError: "duration is lower than rate unit",
		},
	} {
		t.Run(test.expectedError, func(t *testing.T) {
			_, _, err := CalculateRampUpRate(test.startRate, test.endRate, test.duration)

			require.EqualError(t, err, test.expectedError)
		})
	}
}
