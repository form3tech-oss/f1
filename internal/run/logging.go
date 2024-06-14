package run

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func directoryExists(path string) bool {
	directoryPath := filepath.Dir(path)
	if _, err := os.Stat(directoryPath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func GetGeneratedLogFilePath(scenario string) string {
	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf("f1-%s-%s.log", scenario, time.Now().Format("2006-01-02_15-04-05")),
	)
}

func GetLogFilePath(scenario string, logPath string) string {
	if logPath != "" && directoryExists(logPath) {
		return logPath
	}
	return GetGeneratedLogFilePath(scenario)
}

func redirectLoggingToFile(scenario string, logPath string, logger *logrus.Logger) string {
	logFilePath := GetLogFilePath(scenario, logPath)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err == nil {
		logger.SetOutput(file)
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	return logFilePath
}
