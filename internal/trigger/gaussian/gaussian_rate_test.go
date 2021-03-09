package gaussian

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"

	"github.com/stretchr/testify/require"

	"github.com/guptarohit/asciigraph"
	"github.com/stretchr/testify/assert"
)

func TestTotalVolumes(t *testing.T) {
	for i, test := range []struct {
		peak      time.Duration
		stddev    time.Duration
		frequency time.Duration
		weights   []float64
		volume    float64
		repeat    time.Duration
		jitter    float64
	}{
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
			jitter:    2,
		},
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
			jitter:    20,
		},
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
			jitter:    50,
		},
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 10 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    2 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    3 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    4 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    24 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      0 * time.Hour,
			stddev:    10 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    3 * time.Hour,
			frequency: 1 * time.Minute,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    3 * time.Hour,
			frequency: 1 * time.Second,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Second,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    3 * time.Hour,
			frequency: 5 * time.Second,
			volume:    1000000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    3 * time.Hour,
			frequency: 100 * time.Millisecond,
			volume:    1000000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Hour,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Hour,
			volume:    100000,
			repeat:    24 * time.Hour,
		},
		{
			peak:      14 * time.Minute,
			stddev:    1 * time.Hour,
			frequency: 1 * time.Second,
			volume:    100000,
			repeat:    2 * time.Hour,
		},
	} {
		t.Run(fmt.Sprintf("%d: %f every %s, stddev: %s, peak: %s, jitter %f", i, test.volume, test.frequency.String(), test.stddev, test.peak, test.jitter), func(t *testing.T) {
			c := NewGaussianRateCalculator(test.peak, test.stddev, test.frequency, test.weights, test.volume, 24*time.Hour)
			total := 0.0
			current := time.Now().Truncate(24 * time.Hour)
			end := current.Add(24 * time.Hour)

			calculate := api.WithJitter(c.For, test.jitter)
			var rates []float64
			for ; current.Before(end); current = current.Add(test.frequency) {
				rate := calculate(current)
				rates = append(rates, float64(rate))
				total += float64(rate)
			}

			fmt.Println(asciigraph.Plot(rates, asciigraph.Height(15), asciigraph.Width(160), asciigraph.Caption(fmt.Sprintf("Rate per %s", test.frequency.String()))))
			diff := math.Abs(total - test.volume)
			fmt.Printf("Configured for volume %f, triggered %f. Difference of %f (%f%%)\n", test.volume, total, diff, 100*diff/test.volume)
			acceptableErrorPercent := 0.1
			assert.True(t, diff < test.volume*acceptableErrorPercent/100, "volumes differ by > %f%%", acceptableErrorPercent*100)
		})
	}
}

func TestWeightedVolumes(t *testing.T) {
	for i, test := range []struct {
		weights        []float64
		volume         int
		expectedTotals []int
	}{
		{
			volume:         1000000,
			weights:        []float64{1, 0.5, 1.5, 1},
			expectedTotals: []int{1000000, 500000, 1500000, 1000000},
		},
		{
			volume:         1000000,
			weights:        []float64{1, 2, 2, 1},
			expectedTotals: []int{666666, 1333334, 1333334, 666666},
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			iterationDuration := 10 * time.Second
			repeatEvery := 10 * time.Minute
			c := NewGaussianRateCalculator(repeatEvery/2, 1*time.Minute, iterationDuration, test.weights, float64(test.volume), repeatEvery)
			total := 0.0
			require.Equal(t, len(test.weights), len(test.expectedTotals))
			expectedTotal := 0
			for _, t := range test.expectedTotals {
				expectedTotal += t
			}

			current := time.Now().Truncate(repeatEvery * time.Duration(len(test.weights)))

			for i := 0; i < len(test.weights); i++ {
				fmt.Printf("Testing weight %d\n", i)
				repetitionEnd := current.Add(repeatEvery)
				repetitionTotal := 0.0
				for ; current.Before(repetitionEnd); current = current.Add(iterationDuration) {
					rate := c.For(current)
					total += float64(rate)
					repetitionTotal += float64(rate)
				}
				diff := math.Abs(repetitionTotal - float64(test.expectedTotals[i]))
				fmt.Printf("Configured for volume %d, triggered %f. Difference of %f (%f%%)\n", test.expectedTotals[i], repetitionTotal, diff, 100*diff/float64(test.expectedTotals[i]))
				acceptableErrorPercent := 0.1
				assert.True(t, diff < float64(test.expectedTotals[i])*acceptableErrorPercent/100.0, "volumes differ by > %f%%", acceptableErrorPercent*100.0)
			}
			require.Equal(t, expectedTotal, test.volume*len(test.weights))
		})
	}
}
