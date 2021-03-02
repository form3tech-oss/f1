package main

import (
	"log"
	"time"

	"github.com/form3tech-oss/f1/pkg/common_plugin"
	"github.com/form3tech-oss/f1/pkg/f1/testing"
	"github.com/hashicorp/go-plugin"
)

var (
	runFunc      testing.RunFn
	teardownFunc testing.TeardownFn
)

// Interface implementation
type F1PluginFpsGateway struct{}

func (g *F1PluginFpsGateway) GetScenarios() []string {
	return []string{"scenario 1", "scenario 2"}
}

func (g *F1PluginFpsGateway) SetupScenario(name string) {
	runFunc, teardownFunc = setupFpsGatewayScenario(nil)
}

func (g *F1PluginFpsGateway) RunScenarioIteration(name string) {
	runFunc(nil)
}

func (g *F1PluginFpsGateway) StopScenario(name string) {
	teardownFunc(nil)
}

func setupFpsGatewayScenario(t *testing.T) (testing.RunFn, testing.TeardownFn) {
	log.Println("setting up scenario inside plugin")

	runFunc := func(t *testing.T) {
		// log.Println("starting plugin iteration")
		time.Sleep(50 * time.Millisecond)
		// log.Println("finished plugin iteration")
	}

	teardownFunc := func(t *testing.T) {
		log.Println("tearing down scenario inside plugin")
	}

	return runFunc, teardownFunc
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
