package plugin

import (
	"github.com/form3tech-oss/f1/pkg/common_plugin"
	"github.com/form3tech-oss/f1/pkg/f1/testing"
)

var (
	plugins []common_plugin.F1PluginInterface
)

func RegisterPlugin(p common_plugin.F1PluginInterface) {
	plugins = append(plugins, p)
	registerScenarios()
}

func ActivePlugins() []common_plugin.F1PluginInterface {
	return plugins
}

func GetPlugin() common_plugin.F1PluginInterface {
	return plugins[0]
}

func registerScenarios() {
	setupFn := func(t *testing.T) (testing.RunFn, testing.TeardownFn) {
		p := GetPlugin()
		p.SetupScenario("dummy")

		runFn := func(t *testing.T) {
			p.RunScenarioIteration("dummy")
		}

		teardownFn := func(t *testing.T) {
			p.StopScenario("dummy")
		}

		return runFn, teardownFn
	}

	testing.Add("dummy", setupFn)
}
