package main

import (
	"errors"
	"log"
	"time"

	"github.com/form3tech-oss/f1/pkg/common_plugin"
	"github.com/form3tech-oss/f1/pkg/f1/testing"
	"github.com/hashicorp/go-plugin"
)

// Interface implementation
type F1PluginFpsGateway struct {
	scenarios map[string]*scenario
}

type scenario struct {
	setupFn    testing.SetupFn
	runFn      testing.RunFn
	teardownFn testing.TeardownFn
	t          *testing.T
}

func (g *F1PluginFpsGateway) getScenarioByName(name string) *scenario {
	return g.scenarios[name]
}

func (g *F1PluginFpsGateway) GetScenarios() []string {
	var result []string
	for name := range g.scenarios {
		result = append(result, name)
	}
	return result
}

func (g *F1PluginFpsGateway) SetupScenario(name string) error {
	s := g.getScenarioByName(name)
	s.t = testing.NewT(make(map[string]string), "virtual user", "iter", name)

	s.runFn, s.teardownFn = setupFpsGatewayScenario(s.t)

	if s.t.HasFailed() {
		return errors.New("setup scenario failed")
	}

	return nil
}

func (g *F1PluginFpsGateway) RunScenarioIteration(name string) error {
	s := g.getScenarioByName(name)
	s.runFn(s.t)

	if s.t.HasFailed() {
		return errors.New("iteration failed")
	}

	return nil
}

func (g *F1PluginFpsGateway) StopScenario(name string) error {
	s := g.getScenarioByName(name)
	s.teardownFn(s.t)

	if s.t.HasFailed() {
		return errors.New("stop scenario failed")
	}

	return nil
}

func setupFpsGatewayScenario(t *testing.T) (testing.RunFn, testing.TeardownFn) {
	log.Println("setting up scenario inside plugin")

	runFunc := func(t *testing.T) {
		// assert.Fail(t, "I'm failing")
		time.Sleep(50 * time.Millisecond)
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
	f1PluginFpsGateway.scenarios = make(map[string]*scenario)
	f1PluginFpsGateway.scenarios["fpsGateway"] = &scenario{
		setupFn: setupFpsGatewayScenario,
	}

	pluginMap := map[string]plugin.Plugin{
		"fpsgateway": &common_plugin.F1Plugin{Impl: f1PluginFpsGateway},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
