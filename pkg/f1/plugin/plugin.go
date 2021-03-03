package plugin

import (
	"os"
	"os/exec"

	"github.com/form3tech-oss/f1/pkg/common_plugin"
	"github.com/form3tech-oss/f1/pkg/f1/testing"
	"github.com/hashicorp/go-plugin"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"scenarioplugin": &common_plugin.F1Plugin{},
}

func RegisterPlugin(p common_plugin.F1PluginInterface) {
	for _, scenarioName := range p.GetScenarios() {
		copyScenario := scenarioName
		setupFn := func(t *testing.T) (testing.RunFn, testing.TeardownFn) {
			p.SetupScenario(copyScenario)

			runFn := func(t *testing.T) {
				p.RunScenarioIteration(copyScenario)
			}

			teardownFn := func(t *testing.T) {
				p.StopScenario(copyScenario)
			}

			return runFn, teardownFn
		}

		testing.Add(copyScenario, setupFn)
	}
}

func DiscoverPlugins() ([]string, error) {
	return plugin.Discover("*-plugin", "~/.f1/plugins")
}

func Launch(pluginPath string) (*plugin.Client, common_plugin.F1PluginInterface, error) {
	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(pluginPath),
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("scenarioplugin")
	if err != nil {
		return nil, nil, err
	}

	// We should have a GetScenarios now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	return client, raw.(common_plugin.F1PluginInterface), nil
}

func LaunchAll() (func(), error) {
	var clients []*plugin.Client

	paths, err := plugin.Discover("*-plugin", pluginDir())
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		c, p, err := Launch(path)
		if err != nil {
			return nil, err
		}

		clients = append(clients, c)
		RegisterPlugin(p)
	}

	fn := func() {
		for _, c := range clients {
			c.Kill()
		}
	}

	return fn, nil
}

func pluginDir() string {
	return os.ExpandEnv("${HOME}/.f1/plugins")
}
