package scenarios

import (
	"github.com/form3tech-oss/f1/pkg/f1/testing"
	log "github.com/sirupsen/logrus"
	"time"
)

func CreatePaymentScenario(t *testing.T) (testing.RunFn, testing.TeardownFn) {
	runFunc := func(t *testing.T) {
		// assert.Fail(t, "I'm failing")
		time.Sleep(100 * time.Millisecond)

		dummyEnv := t.Environment["FOO"]
		log.Printf("Received environment variable FOO = %+v\n", dummyEnv)
	}

	teardownFunc := func(t *testing.T) {
	}

	return runFunc, teardownFunc
}
