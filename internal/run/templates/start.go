package templates

import (
	"time"
)

//nolint:lll // templates read better with long lines
const startTemplate = `{u}{bold}{intensive_blue}F1 Load Tester{-}
Running {yellow}{{.Scenario}}{-} scenario for {{if .MaxIterations}}up to {{.MaxIterations}} iterations or up to {{end}}{{duration .MaxDuration}} at a rate of {{.RateDescription}}.
`

type StartData struct {
	Scenario        string
	RateDescription string
	MaxIterations   uint64
	MaxDuration     time.Duration
}

func (t *Templates) Start(data StartData) string {
	return render(t.start, data)
}
