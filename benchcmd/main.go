package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/form3tech-oss/f1/v3/pkg/f1"
	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
)

func main() {
	f1.New().
		AddScenario("emptyScenario", emptyScenario).
		AddScenario("failingScenario", failingScenario).
		AddScenario("sleepScenario", sleepScenario).
		AddScenario("logScenario", logScenario).
		Execute()
}

func emptyScenario(context.Context, *f1testing.T) f1testing.RunFn {
	return func(_ context.Context, t *f1testing.T) {
		t.Require().True(true)
	}
}

func sleepScenario(_ context.Context, t *f1testing.T) f1testing.RunFn {
	msString := os.Getenv("MS_SLEEP")
	ms, err := strconv.ParseInt(msString, 10, 64)
	t.Require().NoError(err)

	runFn := func(_ context.Context, _ *f1testing.T) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}

	return runFn
}

func failingScenario(context.Context, *f1testing.T) f1testing.RunFn {
	return func(_ context.Context, t *f1testing.T) {
		t.Require().True(false)
	}
}

func logScenario(_ context.Context, t *f1testing.T) f1testing.RunFn {
	t.Log("Setup")
	runFn := func(_ context.Context, t *f1testing.T) {
		t.Logf("Iteration: %d", t.Iteration)
		t.Logger().With("iteration", t.Iteration).Debug("Trace log")
		t.Logger().With("iteration", t.Iteration).Debug("Debug log")
		t.Logger().With("iteration", t.Iteration).Info("Info log")
		t.Logger().With("iteration", t.Iteration).Warn("Warn log")
		t.Logger().With("iteration", t.Iteration).Error("Error log")

		panic("panic message")
	}

	return runFn
}
