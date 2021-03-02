package common_plugin

import (
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

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

func (g *F1PluginRpcClient) SetupScenario(name string) error {
	var err error

	clientErr := g.client.Call("Plugin.SetupScenario", name, err)
	if clientErr != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(clientErr)
	}

	return err
}

func (g *F1PluginRpcClient) RunScenarioIteration(name string) error {
	var err error

	clientErr := g.client.Call("Plugin.RunScenarioIteration", name, err)
	if clientErr != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(clientErr)
	}

	return err
}

func (g *F1PluginRpcClient) StopScenario(name string) error {
	var err error

	clientErr := g.client.Call("Plugin.StopScenario", name, err)
	if clientErr != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(clientErr)
	}

	return err
}

func (F1Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &F1PluginRpcClient{client: c}, nil
}
