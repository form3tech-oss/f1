package templates_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/progress"
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
				Duration:                 1 * time.Minute,
				SuccessfulIterationCount: 10,
				DroppedIterationCount:    3,
				FailedIterationCount:     5,
				Period:                   10 * time.Second,
				SuccessfulIterationDurationsForPeriod: progress.IterationDurationsSnapshot{
					Average: 10 * time.Microsecond,
					Min:     1 * time.Microsecond,
					Max:     20 * time.Microsecond,
					Count:   10,
				},
			},
			expected: "[ 1m0s]  ✔    10  ⦸     3  ✘     5 (1/s)   avg: 10µs, min: 1µs, max: 20µs",
		},
		{
			name: "rate rounding",
			data: templates.ProgressData{
				Duration:                 1 * time.Minute,
				SuccessfulIterationCount: 10,
				DroppedIterationCount:    3,
				FailedIterationCount:     5,
				Period:                   980 * time.Millisecond,
				SuccessfulIterationDurationsForPeriod: progress.IterationDurationsSnapshot{
					Average: 10 * time.Microsecond,
					Min:     1 * time.Microsecond,
					Max:     20 * time.Microsecond,
					Count:   10,
				},
			},
			expected: "[ 1m0s]  ✔    10  ⦸     3  ✘     5 (10/s)   avg: 10µs, min: 1µs, max: 20µs",
		},
		{
			name: "period less than 500ms",
			data: templates.ProgressData{
				Duration:                 1 * time.Minute,
				SuccessfulIterationCount: 10,
				DroppedIterationCount:    3,
				FailedIterationCount:     5,
				Period:                   100 * time.Millisecond,
				SuccessfulIterationDurationsForPeriod: progress.IterationDurationsSnapshot{
					Average: 10 * time.Microsecond,
					Min:     1 * time.Microsecond,
					Max:     20 * time.Microsecond,
					Count:   10,
				},
			},
			expected: "[ 1m0s]  ✔    10  ⦸     3  ✘     5 (0/s)   avg: 10µs, min: 1µs, max: 20µs",
		},
		{
			name: "no iterations",
			data: templates.ProgressData{
				Duration:                 1 * time.Minute,
				SuccessfulIterationCount: 0,
				DroppedIterationCount:    0,
				FailedIterationCount:     0,
				Period:                   1 * time.Second,
				SuccessfulIterationDurationsForPeriod: progress.IterationDurationsSnapshot{
					Average: 0,
					Min:     0,
					Max:     0,
					Count:   0,
				},
			},
			expected: "[ 1m0s]  ✔     0  ✘     0 (0/s)   avg: 0s, min: 0s, max: 0s",
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
