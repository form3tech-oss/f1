package run

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type logFilePath struct {
	customPath string
	runName    string
}

func LogFilePathOrDefault(customPath, runName string) string {
	return logFilePath{
		customPath: customPath,
		runName:    runName,
	}.Path()
}

func (o logFilePath) Path() string {
	if o.isCustomPathValid() {
		return o.customPath
	}

	return o.defaultFilePath()
}

func (o logFilePath) isCustomPathValid() bool {
	if o.customPath == "" {
		return false
	}

	directoryPath := filepath.Dir(o.customPath)
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (o logFilePath) defaultFilePath() string {
	var uniqPart string

	data := make([]byte, 2)
	if _, err := rand.Read(data); err == nil {
		uniqPart = hex.EncodeToString(data)
	}

	return filepath.Join(
		os.TempDir(),
		fmt.Sprintf("f1-%s-%s-%s.log", o.runName, uniqPart, time.Now().Format("2006-01-02_15-04-05")),
	)
}
