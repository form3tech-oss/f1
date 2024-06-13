package gaussian_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/trigger/gaussian"
)

func Benchmark_calculateVolume(b *testing.B) {
	tests := []struct {
		name     string
		peakTps  string
		peakTime time.Duration
		stddev   time.Duration
	}{
		{
			name:     "1400TPS",
			peakTps:  "1400/s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
		},
		{
			name:     "50TPS",
			peakTps:  "50/s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
		},
		{
			name:     "10TPS",
			peakTps:  "10/s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
		},
		{
			name:     "10TPS 2",
			peakTps:  "100/10s",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
		},
		{
			name:     "10TPS no unit",
			peakTps:  "10",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
		},
		{
			name:     "1TPms",
			peakTps:  "1/ms",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
		},
		{
			name:     "1000TPms",
			peakTps:  "1000/ms",
			stddev:   4 * time.Hour,
			peakTime: 14 * time.Hour,
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			_, err := gaussian.CalculateVolume(tt.peakTps, tt.peakTime, tt.stddev)
			require.NoError(b, err)
		})
	}
}
