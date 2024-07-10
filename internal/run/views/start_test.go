package views_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/run/views"
)

func Test_RenderStart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		expected    string
		expectedLog string
		data        views.StartData
	}{
		{
			name: "with MaxIterations",
			data: views.StartData{
				Scenario:        "scenarioName",
				MaxDuration:     1 * time.Minute,
				RateDescription: "rate-description",
				MaxIterations:   10,
			},
			expected: "F1 Load Tester\n" +
				"Running scenarioName scenario for up to 10 iterations or up to 1m0s at a rate of rate-description.\n",
			expectedLog: "level=INFO msg=\"Running scenarioName for up to 10 iterations or up to 1m0s at a rate of rate-description\"\n",
		},
		{
			name: "without MaxIterations",
			data: views.StartData{
				Scenario:        "scenarioName",
				MaxDuration:     1 * time.Minute,
				RateDescription: "rate-description",
				MaxIterations:   0,
			},
			expected: "F1 Load Tester\n" +
				"Running scenarioName scenario for 1m0s at a rate of rate-description.\n",
			expectedLog: "level=INFO msg=\"Running scenarioName for 1m0s at a rate of rate-description\"\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.Start(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}
