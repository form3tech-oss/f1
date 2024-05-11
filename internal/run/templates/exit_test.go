package templates_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

func Test_RenderTimeout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		expected string
		data     templates.TimeoutData
	}{
		{
			name: "timeout",
			data: templates.TimeoutData{
				Duration: 1 * time.Minute,
			},
			expected: "[ 1m0s]  Max Duration Elapsed - waiting for active tests to complete",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.Timeout(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}

func Test_RenderMaxIterationsReached(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		expected string
		data     templates.MaxIterationsReachedData
	}{
		{
			name: "max iterations",
			data: templates.MaxIterationsReachedData{
				Duration: 1 * time.Minute,
			},
			expected: "[ 1m0s]  Max Iterations Reached - waiting for active tests to complete",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.MaxIterationsReached(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}

func Test_Interrupt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		expected string
		data     templates.InterruptData
	}{
		{
			name: "interrupt",
			data: templates.InterruptData{
				Duration: 1 * time.Minute,
			},
			expected: "[ 1m0s]  Interrupted - waiting for active tests to complete",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.Interrupt(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}
