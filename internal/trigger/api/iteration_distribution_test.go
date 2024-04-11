package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegularRateDistribution(t *testing.T) {
	t.Parallel()

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
			rate:                     25,
			expectedDistributedRates: []int{2, 3, 2, 3, 2, 3, 2, 3, 2, 3},
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
			t.Parallel()

			rateFn := func(time.Time) int { return test.rate }

			distributedIterationDuration, distributedRate := withRegularDistribution(test.iterationDuration, rateFn)
			var result []int
			for range len(test.expectedDistributedRates) {
				result = append(result, distributedRate(time.Now()))
			}

			require.Equal(t, 100*time.Millisecond, distributedIterationDuration)
			require.Equal(t, test.expectedDistributedRates, result)
		})
	}
}

func TestRegularRateDistributionWithSmallIterationDuration(t *testing.T) {
	t.Parallel()

	iterationDuration := 10 * time.Millisecond
	rateFn := func(time.Time) int { return 10_000 }

	distributedIterationDuration, distributedRate := withRegularDistribution(iterationDuration, rateFn)

	require.Equal(t, 10*time.Millisecond, distributedIterationDuration)
	require.Equal(t, 10_000, distributedRate(time.Now()))
}

func TestRegularRateDistributionWithVariableRate(t *testing.T) {
	t.Parallel()

	iterationDuration := 1 * time.Second
	rates := []int{5, 15, 12, 8}
	idx := -1
	rateFn := func(time.Time) int { idx++; return rates[idx] }
	expectedDistributedRates := []int{
		0, 1, 0, 1, 0, 1, 0, 1, 0, 1,
		1, 2, 1, 2, 1, 2, 1, 2, 1, 2,
		1, 1, 1, 1, 2, 1, 1, 1, 1, 2,
		0, 1, 1, 1, 1, 0, 1, 1, 1, 1,
	}

	distributedIterationDuration, distributedRate := withRegularDistribution(iterationDuration, rateFn)
	result := make([]int, len(expectedDistributedRates))
	for i := range len(expectedDistributedRates) {
		result[i] = distributedRate(time.Now())
	}

	require.Equal(t, 100*time.Millisecond, distributedIterationDuration)
	require.Equal(t, expectedDistributedRates, result)
}

func TestRandomRateDistribution(t *testing.T) {
	t.Parallel()

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
			iterationDuration:        200 * time.Millisecond,
			rate:                     1,
			randomValues:             []int{5},
			expectedDistributedRates: []int{1, 0},
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
			rate:                     28,
			randomValues:             []int{0, 1, 0, 0, 1, 0, 0, 0, 7},
			expectedDistributedRates: []int{0, 1, 0, 0, 1, 0, 0, 0, 7, 19},
		},
		{
			iterationDuration:        1 * time.Second,
			rate:                     100,
			randomValues:             []int{20, 5, 5, 10, 0, 5, 15, 10, 10},
			expectedDistributedRates: []int{20, 5, 5, 10, 0, 5, 15, 10, 10, 20},
		},
		{
			iterationDuration:        1 * time.Minute,
			rate:                     600,
			randomValues:             repeatSlice([]int{0, 0, 0, 1, 0, 4, 0, 5, 0, 0}, 60),
			expectedDistributedRates: repeatSlice([]int{0, 0, 0, 1, 0, 4, 0, 5, 0, 0}, 60),
		},
	} {
		t.Run(fmt.Sprintf("%d: iteration duration %s, rate %d", i, test.iterationDuration, test.rate), func(t *testing.T) {
			t.Parallel()

			rateFn := func(time.Time) int { return test.rate }
			idx := -1
			randFn := func(int) int { idx++; return test.randomValues[idx] }

			distributedIterationDuration, distributedRate := withRandomDistribution(test.iterationDuration, rateFn, randFn)
			var result []int
			for range len(test.expectedDistributedRates) {
				result = append(result, distributedRate(time.Now()))
			}

			require.Equal(t, 100*time.Millisecond, distributedIterationDuration)
			require.Equal(t, test.expectedDistributedRates, result)
		})
	}
}

func TestRandomRateDistributionWithVariableRate(t *testing.T) {
	t.Parallel()

	iterationDuration := 1 * time.Second
	rates := []int{5, 15, 12, 8}
	idx := -1
	rateFn := func(time.Time) int { idx++; return rates[idx] }
	randValues := []int{
		0, 1, 0, 1, 0, 1, 0, 1, 0,
		1, 2, 1, 2, 1, 2, 1, 2, 1,
		1, 1, 1, 1, 2, 1, 1, 1, 1,
		0, 1, 1, 1, 1, 0, 1, 1, 1,
	}
	randIdx := -1
	randFn := func(int) int { randIdx++; return randValues[randIdx] }
	expectedDistributedRates := []int{
		0, 1, 0, 1, 0, 1, 0, 1, 0, 1,
		1, 2, 1, 2, 1, 2, 1, 2, 1, 2,
		1, 1, 1, 1, 2, 1, 1, 1, 1, 2,
		0, 1, 1, 1, 1, 0, 1, 1, 1, 1,
	}

	distributedIterationDuration, distributedRate := withRandomDistribution(iterationDuration, rateFn, randFn)
	result := make([]int, len(expectedDistributedRates))
	for i := range len(expectedDistributedRates) {
		result[i] = distributedRate(time.Now())
	}

	require.Equal(t, 100*time.Millisecond, distributedIterationDuration)
	require.Equal(t, expectedDistributedRates, result)
}

func repeatSlice(arr []int, times int) []int {
	var newArr []int

	for range times {
		newArr = append(newArr, arr...)
	}

	return newArr
}

func repeatValue(value int, times int) []int {
	return repeatSlice([]int{value}, times)
}
