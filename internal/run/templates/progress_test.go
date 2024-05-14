package templates_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

func Test_RenderProgress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		expected string
		data     templates.ProgressData
	}{
		{
			name: "complete",
			data: templates.ProgressData{
				Duration:                   1 * time.Minute,
				SuccessfulIterationCount:   10,
				DroppedIterationCount:      3,
				FailedIterationCount:       5,
				RecentDuration:             10 * time.Second,
				RecentSuccessfulIterations: 10,
				SuccessfulIterationDurations: metrics.DurationPercentileMap{
					0.5:  10 * time.Microsecond,
					0.95: 15 * time.Microsecond,
					1.0:  20 * time.Microsecond,
				},
			},
			expected: "[ 1m0s]  ✔    10  ⦸     3  ✘     5 (1/s)   p(50): 10µs,  p(95): 15µs, p(100): 20µs",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.Progress(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}
