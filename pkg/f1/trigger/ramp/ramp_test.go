package ramp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRampRate(t *testing.T) {
	for _, test := range []struct {
		testName, startRate, endRate, distribution string
		jitter                                     float64
		duration, expectedIterationDuration        time.Duration
		expectedRates                              []int
	}{
		{
			testName:                  "ramp rate as constant",
			startRate:                 "10/s",
			endRate:                   "10/s",
			duration:                  10 * time.Second,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		},
		{
			testName:                  "ramp up rate",
			startRate:                 "0/s",
			endRate:                   "100/s",
			duration:                  10 * time.Second,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "ramp down rate",
			startRate:                 "100/s",
			endRate:                   "0/s",
			duration:                  10 * time.Second,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10},
		},
		{
			testName:                  "ramp up rate using ms",
			startRate:                 "0/100ms",
			endRate:                   "100/100ms",
			duration:                  1 * time.Second,
			distribution:              "none",
			expectedIterationDuration: 100 * time.Millisecond,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "ramp up rate using minutes",
			startRate:                 "0/m",
			endRate:                   "60/m",
			duration:                  10 * time.Minute,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Minute,
			expectedRates:             []int{0, 6, 12, 18, 24, 30, 36, 42, 48, 54},
		},
		{
			testName:                  "ramp up rate using hours",
			startRate:                 "0/h",
			endRate:                   "100/h",
			duration:                  10 * time.Hour,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Hour,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "ramp up rate using multiple of durations",
			startRate:                 "0/90s",
			endRate:                   "100/90s",
			duration:                  15 * time.Minute,
			distribution:              "none",
			expectedIterationDuration: 90 * time.Second,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "use 1s as default when unit is not provided",
			startRate:                 "0",
			endRate:                   "100",
			duration:                  10 * time.Second,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90},
		},
		{
			testName:                  "duration is the same as start/end rate unit",
			startRate:                 "0/s",
			endRate:                   "10/s",
			duration:                  1 * time.Second,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0},
		},
		{
			testName:                  "ramp up rate using jitter",
			startRate:                 "0/s",
			endRate:                   "100/s",
			duration:                  10 * time.Second,
			jitter:                    20,
			distribution:              "none",
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 12, 16, 26, 38, 54, 76, 64, 86, 73},
		},
		{
			testName:                  "ramp up rate using distribution regular",
			startRate:                 "0/s",
			endRate:                   "100/s",
			duration:                  5 * time.Second,
			distribution:              "regular",
			expectedIterationDuration: 100 * time.Millisecond,
			expectedRates: []int{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
				4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
				6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
				8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
			},
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, "2020-12-10T10:00:00+00:00")

			rampRates, err := CalculateRampRate(test.startRate, test.endRate, test.distribution, test.duration, test.jitter)

			require.NoError(t, err)
			require.Equal(t, test.duration, rampRates.Duration)
			require.Equal(t, test.expectedIterationDuration, rampRates.IterationDuration)
			var rates []int
			for range test.expectedRates {
				now = now.Add(test.expectedIterationDuration)
				rate := rampRates.Rate(now)
				rates = append(rates, rate)
			}
			require.Equal(t, test.expectedRates, rates)
		})
	}
}

func TestRampRate_Errors(t *testing.T) {
	for _, test := range []struct {
		startRate, endRate, distribution string
		duration                         time.Duration
		jitter                           float64
		expectedError                    string
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
			startRate:     "-10/s",
			endRate:       "10/s",
			distribution:  "none",
			duration:      10 * time.Second,
			expectedError: "unable to parse rate arg -10/s",
		},
		{
			startRate:     "10/error",
			endRate:       "10/s",
			distribution:  "none",
			duration:      10 * time.Second,
			expectedError: "unable to parse rate arg 10/error",
		},
		{
			startRate:     "10/-100ms",
			endRate:       "10/100ms",
			distribution:  "none",
			duration:      10 * time.Second,
			expectedError: "unable to parse rate arg 10/-100ms",
		},
		{
			startRate:     "-100",
			endRate:       "10/100ms",
			distribution:  "none",
			duration:      10 * time.Second,
			expectedError: "unable to parse rate arg -100",
		},
		{
			startRate:     "10/s",
			endRate:       "100/s",
			duration:      100 * time.Millisecond,
			expectedError: "duration is lower than rate unit",
		},
		{
			startRate:     "10/s",
			endRate:       "100/s",
			duration:      10 * time.Second,
			distribution:  "invalid",
			expectedError: "unable to parse distribution invalid",
		},
	} {
		t.Run(test.expectedError, func(t *testing.T) {
			rampRates, err := CalculateRampRate(test.startRate, test.endRate, test.distribution, test.duration, test.jitter)

			require.Nil(t, rampRates)
			require.EqualError(t, err, test.expectedError)
		})
	}
}
