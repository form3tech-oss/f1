package f1

import (
	"github.com/form3tech-oss/f1/v2/pkg/f1/testing"
)

// CombineScenarios creates a single scenario that will call each ScenarioFn
// sequentially and return a testing.RunFn that will call each scenario's RunFn
// every iteration.
func CombineDeprecatedScenarios(scenarios ...testing.ScenarioFn) testing.ScenarioFn {
	return func(t *testing.T) testing.RunFn {
		var run []testing.RunFn
		for _, s := range scenarios {
			run = append(run, s(t))
		}

		return func(t *testing.T) {
			for _, r := range run {
				r(t)
			}
		}
	}
}

func CombineScenarios(scenarios ...testing.ScenarioFunc) testing.ScenarioFunc {
	return func(t testing.TF) testing.RunFunc {
		var run []testing.RunFunc
		for _, s := range scenarios {
			run = append(run, s(t))
		}

		return func(t testing.TF) {
			for _, r := range run {
				r(t)
			}
		}
	}
}
