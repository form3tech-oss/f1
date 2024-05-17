package main

import (
	"github.com/form3tech-oss/f1/v2/pkg/f1"
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

func main() {
	f1.New().
		Add("emptyScenario", emptyScenario).
		Add("failingScenario", failingScenario).
		Execute()
}

func emptyScenario(*testing.T) testing.RunFn {
	runFn := func(t *testing.T) {
		t.Require().True(true)
	}

	return runFn
}

func failingScenario(*testing.T) testing.RunFn {
	runFn := func(t *testing.T) {
		t.Require().True(false)
	}

	return runFn
}
