package common_plugin

// Interface
type F1PluginInterface interface {
	GetScenarios() []string
	SetupScenario(name string) error        // Setup pool of go workers and run SetupFn
	RunScenarioIteration(name string) error // Run RunFn inside of go worker
	StopScenario(name string) error
}

// F1 plugin
type F1Plugin struct {
	Impl F1PluginInterface
}
