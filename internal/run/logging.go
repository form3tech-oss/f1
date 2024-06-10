package run

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LogOutput struct {
	logPath    string
	name       string
}

func NewLogOutput(logPath, name string) LogOutput {
	l := LogOutput{
		logPath: logPath,
		name:    name,
	}

	return l
}

func (l LogOutput) LogPath() string {
	if l.isLogPathValid() {
		return l.logPath
	}

	return l.defaultLogFilePath()
}

func (l LogOutput) isLogPathValid() bool {
	if l.logPath == "" {
		return false
	}

	directoryPath := filepath.Dir(l.logPath)
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (l LogOutput) defaultLogFilePath() string {
	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf("f1-%s-%s.log", l.name, time.Now().Format("2006-01-02_15-04-05")),
	)
}
