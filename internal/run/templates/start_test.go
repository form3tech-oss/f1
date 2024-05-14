package templates_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

func Test_RenderStart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		expected string
		data     templates.StartData
	}{
		{
			name: "with MaxIterations",
			data: templates.StartData{
				Scenario:        "scenarioName",
				MaxDuration:     1 * time.Minute,
				RateDescription: "rate-description",
				MaxIterations:   10,
			},
			expected: "F1 Load Tester\n" +
				"Running scenarioName scenario for up to 10 iterations or up to 1m0s at a rate of rate-description.\n",
		},
		{
			name: "without MaxIterations",
			data: templates.StartData{
				Scenario:        "scenarioName",
				MaxDuration:     1 * time.Minute,
				RateDescription: "rate-description",
				MaxIterations:   0,
			},
			expected: "F1 Load Tester\n" +
				"Running scenarioName scenario for 1m0s at a rate of rate-description.\n",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.Start(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}
