package views_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/run/views"
)

func Test_RenderTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		expected    string
		expectedLog string
		data        views.TimeoutData
	}{
		{
			name: "timeout",
			data: views.TimeoutData{
				Duration: 1 * time.Minute,
			},
			expected:    "[ 1m0s]  Max Duration Elapsed - waiting for active tests to complete",
			expectedLog: "level=INFO msg=\"Max Duration Elapsed - waiting for active tests to complete\" duration=1m0s\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.Timeout(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}

func Test_RenderMaxIterationsReached(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		expected    string
		expectedLog string
		data        views.MaxIterationsReachedData
	}{
		{
			name: "max iterations",
			data: views.MaxIterationsReachedData{
				Duration: 1 * time.Minute,
			},
			expected:    "[ 1m0s]  Max Iterations Reached - waiting for active tests to complete",
			expectedLog: "level=INFO msg=\"Max Iterations Reached - waiting for active tests to complete\" duration=1m0s\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.MaxIterationsReached(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}

func Test_Interrupt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		expected    string
		expectedLog string
		data        views.InterruptData
	}{
		{
			name: "interrupt",
			data: views.InterruptData{
				Duration: 1 * time.Minute,
			},
			expected:    "[ 1m0s]  Interrupted - waiting for active tests to complete",
			expectedLog: "level=INFO msg=\"Interrupted - waiting for active tests to complete\" duration=1m0s\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.Interrupt(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}
