package main

import (
	"errors"
	"github.com/form3tech-oss/f1/pkg/common_plugin"
	"github.com/form3tech-oss/f1/pkg/f1/run"
	"github.com/form3tech-oss/f1/pkg/f1/testing"
	"github.com/form3tech-oss/f1/pkg/paymentapi_plugin/scenarios"
	"github.com/hashicorp/go-plugin"
)

type ScenarioPlugin struct {
	scenarios map[string]*scenario
}

type scenario struct {
	setupFn    testing.SetupFn
	runFn      testing.RunFn
	teardownFn testing.TeardownFn
	t          *testing.T
}

func (g *ScenarioPlugin) GetScenarios() []string {
	var result []string
	for name := range g.scenarios {
		result = append(result, name)
	}
	return result
}

func (g *ScenarioPlugin) SetupScenario(name string) error {
	s := g.getScenarioByName(name)
	s.t = testing.NewT(run.LoadEnvironment(), "0", "0", name)

	s.runFn, s.teardownFn = s.setupFn(s.t)

	if s.t.HasFailed() {
		return errors.New("setup scenario failed")
	}

	return nil
}

func (g *ScenarioPlugin) RunScenarioIteration(name string) error {
	s := g.getScenarioByName(name)
	s.t = testing.NewT(run.LoadEnvironment(), "0", "0", name)

	s.runFn(s.t)

	if s.t.HasFailed() {
		return errors.New("iteration failed")
	}

	return nil
}

func (g *ScenarioPlugin) StopScenario(name string) error {
	s := g.getScenarioByName(name)
	s.teardownFn(s.t)

	if s.t.HasFailed() {
		return errors.New("stop scenario failed")
	}

	return nil
}

func (g *ScenarioPlugin) getScenarioByName(name string) *scenario {
	return g.scenarios[name]
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// Serve the f1 scenario plugin
func main() {
	p := &ScenarioPlugin{}
	p.scenarios = make(map[string]*scenario)
	p.scenarios["createPayment"] = &scenario{
		setupFn: scenarios.CreatePaymentScenario,
	}

	pluginMap := map[string]plugin.Plugin{
		"scenarioplugin": &common_plugin.F1Plugin{Impl: p},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
