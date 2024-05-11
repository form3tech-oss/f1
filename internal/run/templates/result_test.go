package templates_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
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
				SuccessfulIterationDurations: metrics.DurationPercentileMap{
					0.1: 100 * time.Nanosecond,
				},
				FailedIterationCount: 10,
				FailedIterationDurations: metrics.DurationPercentileMap{
					0.2: 200 * time.Nanosecond,
				},
				DroppedIterationCount: 3,
				LogFile:               "log/file/path.log",
			},
			expected: "\nLoad Test Failed\n" +
				"Error: errorMessage\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 2 (13.33%, 2/second)  p(10.00): 100ns,\n" +
				"Failed Iterations: 10 (66.67%, 10)  p(20.00): 200ns,\n" +
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
				SuccessfulIterationDurations: metrics.DurationPercentileMap{
					0.1: 100 * time.Nanosecond,
				},
				LogFile:                  "log/file/path.log",
				FailedIterationDurations: nil,
				Error:                    nil,
				FailedIterationCount:     0,
				DroppedIterationCount:    0,
			},
			expected: "\nLoad Test Passed\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 15 (100.00%, 15/second)  p(10.00): 100ns,\n" +
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
				SuccessfulIterationDurations: metrics.DurationPercentileMap{
					0.1: 100 * time.Nanosecond,
				},
				DroppedIterationCount:    10,
				LogFile:                  "log/file/path.log",
				FailedIterationCount:     0,
				Error:                    nil,
				FailedIterationDurations: nil,
			},
			expected: "\nLoad Test Passed\n" +
				"20 iterations started in 1s (20/second)\n" +
				"Successful Iterations: 5 (33.33%, 5/second)  p(10.00): 100ns,\n" +
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
