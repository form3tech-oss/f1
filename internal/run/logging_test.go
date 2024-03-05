package run

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvidingCustomLogFilePathWithDirectoryThatDoesExist(t *testing.T) {
	t.Setenv("LOG_FILE_PATH", "/does-not-exist/my-scenario.log")
	expected := getGeneratedLogFilePath("my-scenario")

	actual := getLogFilePath("my-scenario")

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}

func TestProvidingCustomLogFilePathWithDirectoryThatDoesNotExistResultsInGeneratedLogFile(t *testing.T) {
	expected := filepath.Join(os.TempDir(), "my-scenario.log")
	t.Setenv("LOG_FILE_PATH", expected)

	actual := getLogFilePath("my-scenario")

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}

func TestProvidingCustomLogFilePathWhichIsEmptyResultsInGeneratedLogFile(t *testing.T) {
	t.Setenv("LOG_FILE_PATH", "")
	expected := getGeneratedLogFilePath("my-scenario")

	actual := getLogFilePath("my-scenario")

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}
