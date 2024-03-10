package run_test

import (
	"os"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	result := m.Run()

	goleak.VerifyTestMain(m)

	os.Exit(result)
}
