package common_plugin

import "github.com/hashicorp/go-plugin"

// Server implementation
type F1PluginRpcServer struct {
	// This is the real implementation
	Impl F1PluginInterface
}

func (s *F1PluginRpcServer) GetScenarios(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetScenarios()
	return nil
}

func (s *F1PluginRpcServer) SetupScenario(name string, resp *[]string) error {
	return s.Impl.SetupScenario(name)
}

func (s *F1PluginRpcServer) RunScenarioIteration(name string, resp *[]string) error {
	return s.Impl.RunScenarioIteration(name)
}

func (s *F1PluginRpcServer) StopScenario(name string, resp *[]string) error {
	return s.Impl.StopScenario(name)
}

func (p *F1Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &F1PluginRpcServer{Impl: p.Impl}, nil
}
