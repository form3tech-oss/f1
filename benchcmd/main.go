package main

import (
	"os"
	"strconv"
	"time"

	"github.com/form3tech-oss/f1/v2/pkg/f1"
	"github.com/form3tech-oss/f1/v2/pkg/f1/f1testing"
)

func main() {
	f1.New().
		Add("emptyScenario", emptyScenario).
		Add("failingScenario", failingScenario).
		Add("sleepScenario", sleepScenario).
		Add("logScenario", logScenario).
		Execute()
}

func emptyScenario(*f1testing.T) f1testing.RunFn {
	runFn := func(t *f1testing.T) {
		t.Require().True(true)
	}

	return runFn
}

func sleepScenario(t *f1testing.T) f1testing.RunFn {
	msString := os.Getenv("MS_SLEEP")
	ms, err := strconv.ParseInt(msString, 10, 64)
	t.Require().NoError(err)

	runFn := func(*f1testing.T) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}

	return runFn
}

func failingScenario(*f1testing.T) f1testing.RunFn {
	runFn := func(t *f1testing.T) {
		t.Require().True(false)
	}

	return runFn
}

func logScenario(t *f1testing.T) f1testing.RunFn {
	t.Log("Setup")
	runFn := func(t *f1testing.T) {
		t.Logf("Iteration: %s", t.Iteration)
		t.Logger().With("iteration", t.Iteration).Debug("Trace log")
		t.Logger().With("iteration", t.Iteration).Debug("Debug log")
		t.Logger().With("iteration", t.Iteration).Info("Info log")
		t.Logger().With("iteration", t.Iteration).Warn("Warn log")
		t.Logger().With("iteration", t.Iteration).Error("Error log")

		panic("panic message")
	}

	return runFn
}
