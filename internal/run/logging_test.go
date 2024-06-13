package run_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/run"
)

func TestProvidingCustomLogFilePathWithDirectoryThatDoesExist(t *testing.T) {
	t.Parallel()

	logPath := "/does-not-exist/my-scenario.log"
	expected := run.GetGeneratedLogFilePath("my-scenario")

	actual := run.GetLogFilePath("my-scenario", logPath)

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}

func TestProvidingCustomLogFilePathWithDirectoryThatDoesNotExistResultsInGeneratedLogFile(t *testing.T) {
	t.Parallel()

	expected := filepath.Join(os.TempDir(), "my-scenario.log")

	actual := run.GetLogFilePath("my-scenario", expected)

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}

func TestProvidingCustomLogFilePathWhichIsEmptyResultsInGeneratedLogFile(t *testing.T) {
	t.Parallel()

	expected := run.GetGeneratedLogFilePath("my-scenario")

	actual := run.GetLogFilePath("my-scenario", "")

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}
