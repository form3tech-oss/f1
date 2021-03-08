package run

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvidingCustomLogFilePathWithDirectoryThatDoesExist(t *testing.T) {
	defer os.Setenv("LOG_FILE_PATH", os.Getenv("LOG_FILE_PATH"))
	os.Setenv("LOG_FILE_PATH", "/does-not-exist/my-scenario.log")
	expected := getGeneratedLogFilePath("my-scenario")

	actual := getLogFilePath("my-scenario")

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}

func TestProvidingCustomLogFilePathWithDirectoryThatDoesNotExistResultsInGeneratedLogFile(t *testing.T) {
	defer os.Setenv("LOG_FILE_PATH", os.Getenv("LOG_FILE_PATH"))
	expected := filepath.Join(os.TempDir(), "my-scenario.log")
	os.Setenv("LOG_FILE_PATH", expected)

	actual := getLogFilePath("my-scenario")

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}

func TestProvidingCustomLogFilePathWhichIsEmptyResultsInGeneratedLogFile(t *testing.T) {
	defer os.Setenv("LOG_FILE_PATH", os.Getenv("LOG_FILE_PATH"))
	os.Setenv("LOG_FILE_PATH", "")
	expected := getGeneratedLogFilePath("my-scenario")

	actual := getLogFilePath("my-scenario")

	assert.NotEqual(t, "", actual)
	assert.Equal(t, expected, actual)
}
