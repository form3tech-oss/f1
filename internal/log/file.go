package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func directoryExists(path string) bool {
	directoryPath := filepath.Dir(path)
	if _, err := os.Stat(directoryPath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func defaultLogFilePath(scenario string) string {
	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf("f1-%s-%s.log", scenario, time.Now().Format("2006-01-02_15-04-05")),
	)
}

func getLogFilePath(scenario string, logPath string) string {
	if logPath != "" && directoryExists(logPath) {
		return logPath
	}
	return defaultLogFilePath(scenario)
}





