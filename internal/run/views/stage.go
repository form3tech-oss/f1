package views

import (
	"log/slog"

	"github.com/form3tech-oss/f1/v2/internal/log"
)

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

func (d SetupData) Log(logger *slog.Logger) {
	if d.Error != nil {
		logger.Error("setup failed", log.ErrorAttr(d.Error))
	} else {
		logger.Info("setup completed")
	}
}

func (d TeardownData) Log(logger *slog.Logger) {
	if d.Error != nil {
		logger.Error("teardown failed", log.ErrorAttr(d.Error))
	} else {
		logger.Info("teardown completed")
	}
}

func (v *Views) Setup(data SetupData) *ViewContext[SetupData] {
	return &ViewContext[SetupData]{
		view: v.setup,
		data: data,
	}
}

func (v *Views) Teardown(data TeardownData) *ViewContext[TeardownData] {
	return &ViewContext[TeardownData]{
		view: v.teardown,
		data: data,
	}
}
