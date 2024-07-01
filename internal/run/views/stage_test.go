package views_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/run/views"
)

func Test_RenderSetup(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		data        views.SetupData
		name        string
		expected    string
		expectedLog string
	}{
		{
			name: "with Error",
			data: views.SetupData{
				Error: errors.New("errorMessage"),
			},
			expected:    "[Setup]    ✘ errorMessage",
			expectedLog: "level=ERROR msg=\"setup failed\" error=errorMessage\n",
		},
		{
			name:        "without Error",
			data:        views.SetupData{Error: nil},
			expected:    "[Setup]    ✔",
			expectedLog: "level=INFO msg=\"setup completed\"\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.Setup(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}

func Test_RenderTeardown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		data        views.TeardownData
		name        string
		expected    string
		expectedLog string
	}{
		{
			name: "with Error",
			data: views.TeardownData{
				Error: errors.New("errorMessage"),
			},
			expected:    "[Teardown] ✘ errorMessage",
			expectedLog: "level=ERROR msg=\"teardown failed\" error=errorMessage\n",
		},
		{
			name:        "without Error",
			data:        views.TeardownData{Error: nil},
			expected:    "[Teardown] ✔",
			expectedLog: "level=INFO msg=\"teardown completed\"\n",
		},
	}

	v := views.New()
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			view := v.Teardown(testCase.data)

			output := view.Render()
			var logOutput bytes.Buffer
			view.Log(log.NewTestLogger(&logOutput))

			assert.Equal(t, testCase.expected, output)
			assert.Equal(t, testCase.expectedLog, logOutput.String())
		})
	}
}
