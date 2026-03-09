package f1

import (
	"github.com/form3tech-oss/f1/v2/pkg/f1/f1testing"
)

// CombineScenarios creates a single scenario that will call each ScenarioFn
// sequentially and return a f1testing.RunFn that will call each scenario's RunFn
// every iteration.
func CombineScenarios(scenarios ...f1testing.ScenarioFn) f1testing.ScenarioFn {
	return func(t *f1testing.T) f1testing.RunFn {
		run := make([]f1testing.RunFn, 0, len(scenarios))
		for _, s := range scenarios {
			run = append(run, s(t))
		}

		return func(t *f1testing.T) {
			for _, r := range run {
				r(t)
			}
		}
	}
}
