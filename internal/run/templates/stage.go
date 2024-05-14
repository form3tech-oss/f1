package templates

const (
	setupTemplate    = `{cyan}[Setup]{-}    {{if .Error}}{red}✘ {{.Error}}{-}{{else}}{green}✔{-}{{end}}`
	teardownTemplate = `{cyan}[Teardown]{-} {{if .Error}}{red}✘ {{.Error}}{-}{{else}}{green}✔{-}{{end}}`
)

type stageData struct {
	Error error
}

type (
	SetupData    stageData
	TeardownData stageData
)

func (t *Templates) Setup(data SetupData) string {
	return render(t.setup, data)
}

func (t *Templates) Teardown(data TeardownData) string {
	return render(t.teardown, data)
}
