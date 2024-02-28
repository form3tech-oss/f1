package run_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"go.uber.org/goleak"
)

var fakePrometheus FakePrometheus

const fakePrometheusNamespace = "test-namespace"
const fakePrometheusID = "test-run-name"

func TestMain(m *testing.M) {

	var err error
	fakePrometheus.Port, err = freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("PROMETHEUS_PUSH_GATEWAY", fmt.Sprintf("http://localhost:%d/", fakePrometheus.Port))
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("PROMETHEUS_NAMESPACE", fakePrometheusNamespace)
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("PROMETHEUS_LABEL_ID", fakePrometheusID)
	if err != nil {
		log.Fatal(err)
	}

	fakePrometheus.StartServer()

	result := m.Run()

	fakePrometheus.StopServer()

	if result == 0 {
		if err := goleak.Find(); err != nil {
			log.Errorf("goleak: Errors on successful test run: %v\n", err)
			result = 1
		}
	}

	os.Exit(result)
}
