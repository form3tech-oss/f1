package run_test

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	result := m.Run()

	if result == 0 {
		if err := goleak.Find(); err != nil {
			log.Errorf("goleak: Errors on successful test run: %v\n", err)
			result = 1
		}
	}

	os.Exit(result)
}
