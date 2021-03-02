package plugin

import (
	"github.com/form3tech-oss/f1/pkg/common_plugin"
	"github.com/form3tech-oss/f1/pkg/f1/testing"
)

func RegisterPlugin(p common_plugin.F1PluginInterface) {
	for _, scenarioName := range p.GetScenarios() {
		setupFn := func(t *testing.T) (testing.RunFn, testing.TeardownFn) {
			p.SetupScenario(scenarioName)

			runFn := func(t *testing.T) {
				p.RunScenarioIteration(scenarioName)
			}

			teardownFn := func(t *testing.T) {
				p.StopScenario(scenarioName)
			}

			return runFn, teardownFn
		}

		testing.Add(scenarioName, setupFn)
	}
}
