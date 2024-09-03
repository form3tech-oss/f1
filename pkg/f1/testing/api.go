package testing

// ScenarioFunc initialises a scenario and returns the iteration function (RunFn) to be invoked for every iteration
// of the tests.
type ScenarioFunc func(t TF) RunFunc

// RunFunc performs a single iteration of the scenario. 't' may be used for asserting
// results or failing the scenario.
type RunFunc func(t TF)

// ScenarioFn initialises a scenario and returns the iteration function (RunFn) to be invoked for every iteration
// of the tests.
// Provided for backwards compatibility.
// Deprecated: Use ScenarioFunc instead.
type ScenarioFn func(t *T) RunFn

// RunFn performs a single iteration of the scenario. 't' may be used for asserting
// results or failing the scenario.
// Provided for backwards compatibility.
// Deprecated: Use RunFunc instead.
type RunFn func(t *T)
