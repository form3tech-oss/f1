package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConstantRateDistribution(t *testing.T) {
	for i, test := range []struct {
		iterationDuration        time.Duration
		rate                     int
		expectedDistributedRates []int
	}{
		{
			iterationDuration:        100 * time.Millisecond,
			rate:                     1,
			expectedDistributedRates: []int{1},
		},
		{
			iterationDuration:        100 * time.Millisecond,
			rate:                     10,
			expectedDistributedRates: []int{10},
		},
		{
			iterationDuration:        200 * time.Millisecond,
			rate:                     10,
			expectedDistributedRates: []int{5, 5},
		},
		{
			iterationDuration:        900 * time.Millisecond,
			rate:                     7,
			expectedDistributedRates: []int{0, 1, 1, 1, 0, 1, 1, 1, 1},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     1,
			expectedDistributedRates: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     5,
			expectedDistributedRates: []int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     7,
			expectedDistributedRates: []int{0, 1, 1, 0, 1, 1, 0, 1, 1, 1},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     10,
			expectedDistributedRates: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     15,
			expectedDistributedRates: []int{1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     100,
			expectedDistributedRates: []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     200,
			expectedDistributedRates: []int{20, 20, 20, 20, 20, 20, 20, 20, 20, 20},
		},
		{
			iterationDuration:        2 * time.Second,
			rate:                     10,
			expectedDistributedRates: repeatSlice([]int{0, 1, 0, 1, 0, 1, 0, 1, 0, 1}, 2),
		},
		{
			iterationDuration:        2 * time.Second,
			rate:                     100,
			expectedDistributedRates: repeatSlice([]int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, 2),
		},
		{
			iterationDuration:        1 * time.Minute,
			rate:                     60,
			expectedDistributedRates: repeatSlice([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, 60),
		},
		{
			iterationDuration:        10 * time.Minute,
			rate:                     10,
			expectedDistributedRates: repeatSlice(append(repeatValue(0, 599), 1), 10),
		},
		{
			iterationDuration:        1 * time.Minute,
			rate:                     600,
			expectedDistributedRates: repeatSlice([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 600),
		},
		{
			iterationDuration:        1 * time.Minute,
			rate:                     6_000,
			expectedDistributedRates: repeatSlice([]int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, 600),
		},
		{
			iterationDuration:        1 * time.Hour,
			rate:                     3_600,
			expectedDistributedRates: repeatSlice([]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, 3_600),
		},
		{
			iterationDuration:        1 * time.Hour,
			rate:                     36_000,
			expectedDistributedRates: repeatSlice([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 3_600),
		},
		{
			iterationDuration:        1 * time.Hour,
			rate:                     360_000,
			expectedDistributedRates: repeatSlice([]int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, 3_600),
		},
		{
			iterationDuration:        100 * time.Second,
			rate:                     900,
			expectedDistributedRates: repeatSlice([]int{0, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 100),
		},
	} {
		t.Run(fmt.Sprintf("%d: iteration duration %s, rate %d", i, test.iterationDuration, test.rate), func(t *testing.T) {
			rateFn := func(time time.Time) int { return test.rate }

			distributedIterationDuration, distributedRate := WithConstantDistribution(test.iterationDuration, rateFn)
			var result []int
			for i := 0; i < len(test.expectedDistributedRates); i++ {
				result = append(result, distributedRate(time.Now()))
			}

			require.Equal(t, 100*time.Millisecond, distributedIterationDuration)
			require.Equal(t, test.expectedDistributedRates, result)
		})
	}
}

func TestConstantRateDistributionWithSmallIterationDuration(t *testing.T) {
	iterationDuration := 10 * time.Millisecond
	rateFn := func(time time.Time) int { return 10_000 }

	distributedIterationDuration, distributedRate := WithConstantDistribution(iterationDuration, rateFn)

	require.Equal(t, 10*time.Millisecond, distributedIterationDuration)
	require.Equal(t, 10_000, distributedRate(time.Now()))
}

func TestConstantRateDistributionWithVariableRate(t *testing.T) {
	iterationDuration := 1 * time.Second
	rates := []int{5, 15, 12, 8}
	var idx = -1
	rateFn := func(time time.Time) int { idx++; return rates[idx] }
	expectedDistributedRates := []int{
		0, 1, 0, 1, 0, 1, 0, 1, 0, 1,
		1, 2, 1, 2, 1, 2, 1, 2, 1, 2,
		1, 1, 1, 1, 2, 1, 1, 1, 1, 2,
		0, 1, 1, 1, 1, 0, 1, 1, 1, 1,
	}

	distributedIterationDuration, distributedRate := WithConstantDistribution(iterationDuration, rateFn)
	var result []int
	for i := 0; i < len(expectedDistributedRates); i++ {
		result = append(result, distributedRate(time.Now()))
	}

	require.Equal(t, 100*time.Millisecond, distributedIterationDuration)
	require.Equal(t, expectedDistributedRates, result)
}

func TestRandomRateDistribution(t *testing.T) {
	for i, test := range []struct {
		iterationDuration        time.Duration
		rate                     int
		randomValues             []int
		expectedDistributedRates []int
	}{
		{
			iterationDuration:        100 * time.Millisecond,
			rate:                     1,
			randomValues:             []int{},
			expectedDistributedRates: []int{1},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     1,
			randomValues:             []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedDistributedRates: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     1,
			randomValues:             []int{1},
			expectedDistributedRates: []int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     100,
			randomValues:             []int{20, 5, 5, 10, 0, 5, 15, 10, 10},
			expectedDistributedRates: []int{20, 5, 5, 10, 0, 5, 15, 10, 10, 20},
		},
	} {
		t.Run(fmt.Sprintf("%d: iteration duration %s, rate %d", i, test.iterationDuration, test.rate), func(t *testing.T) {
			rateFn := func(time time.Time) int { return test.rate }
			var idx = -1
			randFn := func(limit int) int { idx++; return test.randomValues[idx] }

			distributedIterationDuration, distributedRate := WithRandomDistribution(test.iterationDuration, rateFn, randFn)
			var result []int
			for i := 0; i < len(test.expectedDistributedRates); i++ {
				result = append(result, distributedRate(time.Now()))
			}

			require.Equal(t, 100*time.Millisecond, distributedIterationDuration)
			require.Equal(t, test.expectedDistributedRates, result)
		})
	}
}

func repeatSlice(arr []int, times int) []int {
	var newArr []int

	for i := 0; i < times; i++ {
		newArr = append(newArr, arr...)
	}

	return newArr
}

func repeatValue(value int, times int) []int {
	return repeatSlice([]int{value}, times)
}
