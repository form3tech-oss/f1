package views_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/run/views"
)

func Test_RenderResult(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		expected    string
		expectedLog string
		data        views.ResultData
	}{
		{
			name: "failed",
			data: views.ResultData{
				Failed:                   true,
				Error:                    errors.New("errorMessage"),
				IterationsStarted:        20,
				Duration:                 1 * time.Second,
				SuccessfulIterationCount: 2,
				Iterations:               15,
				SuccessfulIterationDurations: progress.IterationDurationsSnapshot{
					Min:     1 * time.Microsecond,
					Average: 2 * time.Microsecond,
					Max:     3 * time.Microsecond,
				},
				FailedIterationCount: 10,
				FailedIterationDurations: progress.IterationDurationsSnapshot{
					Min:     4 * time.Microsecond,
					Average: 5 * time.Microsecond,
					Max:     6 * time.Microsecond,
				},
				DroppedIterationCount: 3,
				LogFilePath:           "log/file/path.log",
			},
			expected: "\nLoad Test Failed\n" +
				"Error: errorMessage\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 2 (13.33%, 2/second) avg: 2µs, min: 1µs, max: 3µs\n" +
				"Failed Iterations: 10 (66.67%, 10) avg: 5µs, min: 4µs, max: 6µs\n" +
				"Dropped Iterations: 3 (20.00%, 3) (consider increasing --concurrency setting)\n" +
				"Full logs: log/file/path.log\n",
			expectedLog: "level=ERROR msg=\"Load Test Failed\" " +
				"error=errorMessage " +
				"iteration_stats.started=20 " +
				"iteration_stats.successful=2 " +
				"iteration_stats.failed=10 " +
				"iteration_stats.dropped=3 " +
				"iteration_stats.period=1s\n",
		},
		{
			name: "failed without error",
			data: views.ResultData{
				Failed:                   true,
				Error:                    nil,
				IterationsStarted:        20,
				Duration:                 1 * time.Second,
				SuccessfulIterationCount: 2,
				Iterations:               15,
				SuccessfulIterationDurations: progress.IterationDurationsSnapshot{
					Min:     1 * time.Microsecond,
					Average: 2 * time.Microsecond,
					Max:     3 * time.Microsecond,
				},
				FailedIterationCount: 10,
				FailedIterationDurations: progress.IterationDurationsSnapshot{
					Min:     4 * time.Microsecond,
					Average: 5 * time.Microsecond,
					Max:     6 * time.Microsecond,
				},
				DroppedIterationCount: 3,
				LogFilePath:           "log/file/path.log",
			},
			expected: "\nLoad Test Failed\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 2 (13.33%, 2/second) avg: 2µs, min: 1µs, max: 3µs\n" +
				"Failed Iterations: 10 (66.67%, 10) avg: 5µs, min: 4µs, max: 6µs\n" +
				"Dropped Iterations: 3 (20.00%, 3) (consider increasing --concurrency setting)\n" +
				"Full logs: log/file/path.log\n",
			expectedLog: "level=ERROR msg=\"Load Test Failed\" " +
				"iteration_stats.started=20 " +
				"iteration_stats.successful=2 " +
				"iteration_stats.failed=10 " +
				"iteration_stats.dropped=3 " +
				"iteration_stats.period=1s\n",
		},
		{
			name: "passed",
			data: views.ResultData{
				Failed:                   false,
				IterationsStarted:        20,
				Duration:                 1 * time.Second,
				SuccessfulIterationCount: 15,
				Iterations:               15,
				SuccessfulIterationDurations: progress.IterationDurationsSnapshot{
					Min:     1 * time.Microsecond,
					Average: 2 * time.Microsecond,
					Max:     3 * time.Microsecond,
				},
				FailedIterationDurations: progress.IterationDurationsSnapshot{},
				LogFilePath:              "log/file/path.log",
				Error:                    nil,
				FailedIterationCount:     0,
				DroppedIterationCount:    0,
			},
			expected: "\nLoad Test Passed\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 15 (100.00%, 15/second) avg: 2µs, min: 1µs, max: 3µs\n" +
				"Full logs: log/file/path.log\n",
			expectedLog: "level=INFO msg=\"Load Test Passed\" " +
				"iteration_stats.started=20 " +
				"iteration_stats.successful=15 " +
				"iteration_stats.failed=0 " +
				"iteration_stats.dropped=0 " +
				"iteration_stats.period=1s\n",
		},
		{
			name: "passed with dropped iterations",
			data: views.ResultData{
				Failed:                   false,
				IterationsStarted:        20,
				Duration:                 1 * time.Second,
				SuccessfulIterationCount: 5,
				Iterations:               15,
				SuccessfulIterationDurations: progress.IterationDurationsSnapshot{
					Min:     1 * time.Microsecond,
					Average: 2 * time.Microsecond,
					Max:     3 * time.Microsecond,
				},
				FailedIterationDurations: progress.IterationDurationsSnapshot{},
				DroppedIterationCount:    10,
				LogFilePath:              "log/file/path.log",
				FailedIterationCount:     0,
				Error:                    nil,
			},
			expected: "\nLoad Test Passed\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 5 (33.33%, 5/second) avg: 2µs, min: 1µs, max: 3µs\n" +
				"Dropped Iterations: 10 (66.67%, 10) (consider increasing --concurrency setting)\n" +
				"Full logs: log/file/path.log\n",
			expectedLog: "level=INFO msg=\"Load Test Passed\" " +
				"iteration_stats.started=20 " +
				"iteration_stats.successful=5 " +
				"iteration_stats.failed=0 " +
				"iteration_stats.dropped=10 " +
				"iteration_stats.period=1s\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.Result(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}
