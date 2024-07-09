package views

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/ui"
)

//nolint:lll // templates read better with long lines
const startTemplate = `{u}{bold}{intensive_blue}F1 Load Tester{-}
Running {yellow}{{.Scenario}}{-} scenario for {{if .MaxIterations}}up to {{.MaxIterations}} iterations or up to {{end}}{{duration .MaxDuration}} at a rate of {{.RateDescription}}.
`

var _ ui.Outputable = (*ViewContext[StartData])(nil)

type StartData struct {
	Scenario        string
	RateDescription string
	MaxIterations   uint64
	MaxDuration     time.Duration
}

func (c StartData) Log(logger *slog.Logger) {
	message := "Running " + c.Scenario + " for "
	if c.MaxIterations > 0 {
		message += "up to " + strconv.FormatUint(c.MaxIterations, 10) + " iterations or up to "
	}

	message += c.MaxDuration.String()
	message += " at a rate of " + c.RateDescription

	logger.Info(message)
}

func (v *Views) Start(data StartData) *ViewContext[StartData] {
	return &ViewContext[StartData]{
		view: v.start,
		data: data,
	}
}
