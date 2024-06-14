package gaussian_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/gaussian"
)

func TestTotalVolumes(t *testing.T) {
	t.Parallel()

	for i, test := range []struct {
		weights   []float64
		peak      time.Duration
		stddev    time.Duration
		frequency time.Duration
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
			t.Parallel()

			c, err := gaussian.NewCalculator(test.peak, test.stddev, test.frequency, test.weights, test.volume, test.repeat)
			require.NoError(t, err)

			total := 0.0
			current := time.Now().Truncate(test.repeat)
			end := current.Add(test.repeat)

			calculate := api.WithJitter(c.For, test.jitter)
			for ; current.Before(end); current = current.Add(test.frequency) {
				rate := calculate(current)
				total += float64(rate)
			}

			// allow for less than 0.0015% difference
			require.InDelta(t, test.volume, total, 0.0015*test.volume)
		})
	}
}

func TestWeightedVolumes(t *testing.T) {
	t.Parallel()

	for i, test := range []struct {
		weights        []float64
		expectedTotals []int
		volume         int
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
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			iterationDuration := 10 * time.Second
			repeatEvery := 10 * time.Minute

			c, err := gaussian.NewCalculator(repeatEvery/2, 1*time.Minute, iterationDuration, test.weights, float64(test.volume), repeatEvery)
			require.NoError(t, err)

			total := 0.0
			require.Equal(t, len(test.weights), len(test.expectedTotals))

			expectedTotal := 0
			for _, t := range test.expectedTotals {
				expectedTotal += t
			}

			current := time.Now().Truncate(repeatEvery * time.Duration(len(test.weights)))

			for i := range len(test.weights) {
				repetitionEnd := current.Add(repeatEvery)
				repetitionTotal := 0.0
				for ; current.Before(repetitionEnd); current = current.Add(iterationDuration) {
					rate := c.For(current)
					total += float64(rate)
					repetitionTotal += float64(rate)
				}

				require.InDelta(t, test.expectedTotals[i], int(repetitionTotal), 1)
			}
			require.Equal(t, expectedTotal, test.volume*len(test.weights))
		})
	}
}

func Test_calculateVolume(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		peakTps  string
		peakTime time.Duration
		stddev   time.Duration
		want     float64
		wantErr  bool
	}{
		{
			name:     "1400TPS",
			peakTps:  "1400/s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
			want:     50208044,
			wantErr:  false,
		},
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
			t.Parallel()

			got, err := gaussian.CalculateVolume(tt.peakTps, tt.peakTime, tt.stddev)
			if tt.wantErr {
				require.Error(t, err)
			}

			assert.Equal(t, int64(tt.want), int64(got))
		})
	}
}
