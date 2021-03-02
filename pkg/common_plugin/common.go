package common_plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Interface
type F1PluginInterface interface {
	GetScenarios() []string
	SetupScenario(name string)        // Setup pool of go workers and run SetupFn
	RunScenarioIteration(name string) // Run RunFn inside of go worker
	StopScenario(name string)
}

// F1 plugin
type F1Plugin struct {
	Impl F1PluginInterface
}

// Server implementation
type F1PluginRpcServer struct {
	// This is the real implementation
	Impl F1PluginInterface
}

func (s *F1PluginRpcServer) GetScenarios(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetScenarios()
	return nil
}

func (s *F1PluginRpcServer) SetupScenario(args interface{}, resp *[]string) error {
	s.Impl.SetupScenario("setupFpsGatewayScenario")
	return nil
}

func (s *F1PluginRpcServer) RunScenarioIteration(args interface{}, resp *[]string) error {
	s.Impl.RunScenarioIteration("setupFpsGatewayScenario")
	return nil
}

func (s *F1PluginRpcServer) StopScenario(args interface{}, resp *[]string) error {
	s.Impl.StopScenario("setupFpsGatewayScenario")
	return nil
}

func (p *F1Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &F1PluginRpcServer{Impl: p.Impl}, nil
}

// Client implementation
type F1PluginRpcClient struct{ client *rpc.Client }

func (g *F1PluginRpcClient) GetScenarios() []string {
	var resp []string
	err := g.client.Call("Plugin.GetScenarios", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (g *F1PluginRpcClient) SetupScenario(name string) {
	var resp []string
	err := g.client.Call("Plugin.SetupScenario", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}
}

func (g *F1PluginRpcClient) RunScenarioIteration(name string) {
	var resp []string
	err := g.client.Call("Plugin.RunScenarioIteration", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}
}

func (g *F1PluginRpcClient) StopScenario(name string) {
	var resp []string
	err := g.client.Call("Plugin.StopScenario", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}
}

func (F1Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &F1PluginRpcClient{client: c}, nil
}
