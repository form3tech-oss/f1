package views_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/run/views"
)

func Test_RenderProgress(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		expected    string
		expectedLog string
		data        views.ProgressData
	}{
		{
			name: "complete",
			data: views.ProgressData{
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
			expectedLog: "level=INFO msg=progress " +
				"iteration_stats.started=18 " +
				"iteration_stats.successful=10 " +
				"iteration_stats.failed=5 " +
				"iteration_stats.dropped=3 " +
				"iteration_stats.period=10s\n",
		},
		{
			name: "rate rounding",
			data: views.ProgressData{
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
			expectedLog: "level=INFO msg=progress " +
				"iteration_stats.started=18 " +
				"iteration_stats.successful=10 " +
				"iteration_stats.failed=5 " +
				"iteration_stats.dropped=3 " +
				"iteration_stats.period=980ms\n",
		},
		{
			name: "period less than 500ms",
			data: views.ProgressData{
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
			expectedLog: "level=INFO msg=progress " +
				"iteration_stats.started=18 " +
				"iteration_stats.successful=10 " +
				"iteration_stats.failed=5 " +
				"iteration_stats.dropped=3 " +
				"iteration_stats.period=100ms\n",
		},
		{
			name: "no iterations",
			data: views.ProgressData{
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
			expectedLog: "level=INFO msg=progress " +
				"iteration_stats.started=0 " +
				"iteration_stats.successful=0 " +
				"iteration_stats.failed=0 " +
				"iteration_stats.dropped=0 " +
				"iteration_stats.period=1s\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.Progress(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}
