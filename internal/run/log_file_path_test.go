package run_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/run"
)

func TestLogFilePathOrDefault(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		customLogPath  string
		runName        string
		resultContains string
	}{
		{
			name:           "empty custom log path",
			customLogPath:  "",
			runName:        "name",
			resultContains: "f1-name",
		},
		{
			name:           "custom log path does not exist",
			customLogPath:  "/invalid-path/",
			runName:        "name",
			resultContains: "f1-name",
		},
		{
			name:           "custom log path exists",
			customLogPath:  filepath.Join(os.TempDir(), "custom-log-file.log"),
			runName:        "name",
			resultContains: "custom-log-file.log",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := run.LogFilePathOrDefault(testCase.customLogPath, testCase.runName)
			assert.Contains(t, path, testCase.resultContains)
		})
	}
}
