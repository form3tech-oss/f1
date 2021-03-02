package scenarios

import (
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/testing"
)

func AdmissionScenario(t *testing.T) (testing.RunFn, testing.TeardownFn) {
	runFunc := func(t *testing.T) {
		// assert.Fail(t, "I'm failing")
		time.Sleep(50 * time.Millisecond)
	}

	teardownFunc := func(t *testing.T) {
	}

	return runFunc, teardownFunc
}
