package plugin

import "github.com/form3tech-oss/f1/pkg/common_plugin"

var (
	plugins []common_plugin.F1PluginInterface
)

func RegisterPlugin(p common_plugin.F1PluginInterface) {
	plugins = append(plugins, p)
}

func ActivePlugins() []common_plugin.F1PluginInterface {
	return plugins
}
