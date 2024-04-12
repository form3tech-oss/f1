package run

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

func directoryExists(path string) bool {
	directoryPath := filepath.Dir(path)
	if _, err := os.Stat(directoryPath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func getGeneratedLogFilePath(scenario string) string {
	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf("f1-%s-%s.log", scenario, time.Now().Format("2006-01-02_15-04-05")),
	)
}

func getLogFilePath(scenario string, logPath string) string {
	if logPath != "" && directoryExists(logPath) {
		return logPath
	}
	return getGeneratedLogFilePath(scenario)
}

func redirectLoggingToFile(scenario string, logPath string, output io.Writer) string {
	logFilePath := getLogFilePath(scenario, logPath)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err == nil {
		log.StandardLogger().SetOutput(file)
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	stdlog.SetOutput(output)
	return logFilePath
}
