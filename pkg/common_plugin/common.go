package common_plugin

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

// Interface
type F1PluginInterface interface {
	GetScenarios() []string
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

func (F1Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &F1PluginRpcClient{client: c}, nil
}
