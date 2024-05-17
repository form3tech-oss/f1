package templates_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

func Test_RenderResult(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		expected string
		data     templates.ResultData
	}{
		{
			name: "failed",
			data: templates.ResultData{
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
				LogFile:               "log/file/path.log",
			},
			expected: "\nLoad Test Failed\n" +
				"Error: errorMessage\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 2 (13.33%, 2/second) avg: 2µs, min: 1µs, max: 3µs\n" +
				"Failed Iterations: 10 (66.67%, 10) avg: 5µs, min: 4µs, max: 6µs\n" +
				"Dropped Iterations: 3 (20.00%, 3) (consider increasing --concurrency setting)\n" +
				"Full logs: log/file/path.log\n",
		},
		{
			name: "passed",
			data: templates.ResultData{
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
				LogFile:                  "log/file/path.log",
				Error:                    nil,
				FailedIterationCount:     0,
				DroppedIterationCount:    0,
			},
			expected: "\nLoad Test Passed\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 15 (100.00%, 15/second) avg: 2µs, min: 1µs, max: 3µs\n" +
				"Full logs: log/file/path.log\n",
		},
		{
			name: "passed with dropped iterations",
			data: templates.ResultData{
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
				LogFile:                  "log/file/path.log",
				FailedIterationCount:     0,
				Error:                    nil,
			},
			expected: "\nLoad Test Passed\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 5 (33.33%, 5/second) avg: 2µs, min: 1µs, max: 3µs\n" +
				"Dropped Iterations: 10 (66.67%, 10) (consider increasing --concurrency setting)\n" +
				"Full logs: log/file/path.log\n",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.Result(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}
