package run

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

type ScenarioLogger struct {
	Logger *slog.Logger
	output *ui.Output

	logFile *os.File
}

func NewScenarioLogger(output *ui.Output) *ScenarioLogger {
	return &ScenarioLogger{
		output: output,
	}
}

func (s *ScenarioLogger) Open(logFilePath string, logConfig *log.Config, runName string, logToFile bool) string {
	if !logToFile {
		s.Logger = s.output.Logger
		return ""
	}

	logFile, err := s.openLogFile(logFilePath)
	if err != nil {
		s.Logger = s.output.Logger
		s.output.Display(ui.ErrorMessage{Message: "Error opening log file. Using default logger", Error: err})
		return ""
	}

	s.Logger = log.NewLogger(logFile, logConfig).With(log.ScenarioAttr(runName))
	s.logFile = logFile
	s.output.Display(ui.InfoMessage{Message: "Saving logs to " + logFilePath})

	return logFilePath
}

func (s *ScenarioLogger) Close() error {
	if s.logFile != nil {
		if err := s.logFile.Close(); err != nil {
			return fmt.Errorf("closing log file: %w", err)
		}
	}

	return nil
}

func (*ScenarioLogger) openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("opening log file '%s': %w", path, err)
	}

	return logFile, nil
}
