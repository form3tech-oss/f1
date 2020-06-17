package testing

type ScenarioInfo struct {
	Name        string
	Description string
	Parameters  []ScenarioParameter
}

type ScenarioParameter struct {
	Name        string
	Description string
	Default     string
}
