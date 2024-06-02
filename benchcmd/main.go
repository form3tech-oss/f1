package main

import (
	"os"
	"strconv"
	"time"

	"github.com/form3tech-oss/f1/v2/pkg/f1"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

func main() {
	f1.New().
		Add("emptyScenario", emptyScenario).
		Add("failingScenario", failingScenario).
		Add("sleepScenario", sleepScenario).
		Execute()
}

func emptyScenario(*testing.T) testing.RunFn {
	runFn := func(t *testing.T) {
		t.Require().True(true)
	}

	return runFn
}

func sleepScenario(t *testing.T) testing.RunFn {
	msString := os.Getenv("MS_SLEEP")
	ms, err := strconv.ParseInt(msString, 10, 64)
	t.Require().NoError(err)

	runFn := func(*testing.T) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}

	return runFn
}

func failingScenario(*testing.T) testing.RunFn {
	runFn := func(t *testing.T) {
		t.Require().True(false)
	}

	return runFn
}
