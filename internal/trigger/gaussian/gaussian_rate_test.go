package gaussian

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/guptarohit/asciigraph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
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
		{
			peak:      5 * time.Minute,
			stddev:    75 * time.Second,
			frequency: 1 * time.Second,
			volume:    23499,
			repeat:    1 * time.Hour,
		},
	} {
		t.Run(fmt.Sprintf("%d: %f every %s, stddev: %s, peak: %s, jitter %f", i, test.volume, test.frequency.String(), test.stddev, test.peak, test.jitter), func(t *testing.T) {
			c := NewGaussianRateCalculator(test.peak, test.stddev, test.frequency, test.weights, test.volume, test.repeat)
			total := 0.0
			current := time.Now().Truncate(test.repeat)
			end := current.Add(test.repeat)

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
			assert.Less(t, diff, test.volume*acceptableErrorPercent/100, "volumes differ by > %f%%", acceptableErrorPercent*100)
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
				assert.Less(t, diff, float64(test.expectedTotals[i])*acceptableErrorPercent/100.0, "volumes differ by > %f%%", acceptableErrorPercent*100.0)
			}
			require.Equal(t, expectedTotal, test.volume*len(test.weights))
		})
	}
}

func Test_calculateVolume(t *testing.T) {
	tests := []struct {
		name     string
		peakTps  string
		peakTime time.Duration
		stddev   time.Duration
		want     float64
		wantErr  bool
	}{
		{
			name:     "50TPS",
			peakTps:  "50/s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     1793144,
			wantErr:  false,
		},
		{
			name:     "10TPS",
			peakTps:  "10/s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     358629,
			wantErr:  false,
		},
		{
			name:     "10TPS 2",
			peakTps:  "100/10s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     358629,
			wantErr:  false,
		},
		{
			name:     "10TPS no unit",
			peakTps:  "10",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     358629,
			wantErr:  false,
		},
		{
			name:     "1TPms",
			peakTps:  "1/ms",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     35862889,
			wantErr:  false,
		},
		{
			name:     "1000TPms",
			peakTps:  "1000/ms",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     35862888782,
			wantErr:  false,
		},
		{
			name:     "1TPH",
			peakTps:  "1/h",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     10,
			wantErr:  false,
		},
		{
			name:     "error",
			peakTps:  "ms",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     -1,
			wantErr:  true,
		},
		{
			name:     "invalid value",
			peakTps:  "-10/s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     -1,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateVolume(tt.peakTps, tt.peakTime, tt.stddev)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
