package templates_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/run/templates"
)

func Test_RenderSetup(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		data     templates.SetupData
		name     string
		expected string
	}{
		{
			name: "with Error",
			data: templates.SetupData{
				Error: errors.New("errorMessage"),
			},
			expected: "[Setup]    ✘ errorMessage",
		},
		{
			name:     "without Error",
			data:     templates.SetupData{Error: nil},
			expected: "[Setup]    ✔",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.Setup(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}

func Test_RenderTeardown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		data     templates.TeardownData
		name     string
		expected string
	}{
		{
			name: "with Error",
			data: templates.TeardownData{
				Error: errors.New("errorMessage"),
			},
			expected: "[Teardown] ✘ errorMessage",
		},
		{
			name:     "without Error",
			data:     templates.TeardownData{Error: nil},
			expected: "[Teardown] ✔",
		},
	}

	tmpl := templates.Parse(templates.DisableRenderTermColors)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output := tmpl.Teardown(testCase.data)
			assert.Equal(t, testCase.expected, output)
		})
	}
}
