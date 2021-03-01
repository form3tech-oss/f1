package main

import (
	"github.com/form3tech-oss/f1/pkg/common_plugin"
	"github.com/hashicorp/go-plugin"
)

// Interface implementation
type F1PluginFpsGateway struct{}

func (g *F1PluginFpsGateway) GetScenarios() []string {
	return []string{"scenario 1", "scenario 2"}
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// Serve the FPS gateway plugin
func main() {
	f1PluginFpsGateway := &F1PluginFpsGateway{}

	pluginMap := map[string]plugin.Plugin{
		"fpsgateway": &common_plugin.F1Plugin{Impl: f1PluginFpsGateway},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
