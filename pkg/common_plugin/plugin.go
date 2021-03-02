package common_plugin

import (
	"os/exec"

	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"fpsgateway": &F1Plugin{},
}

func Launch() (*plugin.Client, F1PluginInterface) {
	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("./pkg/fpsgateway_plugin/fpsgateway"),
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("fpsgateway")
	if err != nil {
		log.Fatal(err)
	}

	// We should have a GetScenarios now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	return client, raw.(F1PluginInterface)
}