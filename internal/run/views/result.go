package views

import (
	"log/slog"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

//nolint:lll // templates read better with long lines
const resultTemplate = `
{{if .Failed -}}
{red}{bold}{u}Load Test Failed{-}
{{- else -}}
{green}{bold}{u}Load Test Passed{-}
{{- end}}
{{- if .Error}}
{red}Error: {{.Error}}{-}
{{- end}}
{{.IterationsStarted}} iterations started in {{duration .Duration}} ({{rate .Duration .IterationsStarted}}/second)
{{- if .SuccessfulIterationCount}}
{bold}Successful Iterations:{-} {green}{{.SuccessfulIterationCount}} ({{percent .SuccessfulIterationCount .Iterations | printf "%0.2f"}}%, {{rate .Duration .SuccessfulIterationCount}}/second){-} {{.SuccessfulIterationDurations}}
{{- end}}
{{- if .FailedIterationCount}}
{bold}Failed Iterations:{-} {red}{{.FailedIterationCount}} ({{percent .FailedIterationCount .Iterations | printf "%0.2f"}}%, {{rate .Duration .FailedIterationCount}}){-} {{.FailedIterationDurations}}
{{- end}}
{{- if .DroppedIterationCount}}
{bold}Dropped Iterations:{-} {yellow}{{.DroppedIterationCount}} ({{percent .DroppedIterationCount .Iterations | printf "%0.2f"}}%, {{rate .Duration .DroppedIterationCount}}){-} (consider increasing --concurrency setting)
{{- end}}
{bold}Full logs:{-} {{.LogFilePath}}
`

var _ ui.Outputable = (*ViewContext[ResultData])(nil)

type ResultData struct {
	Error                        error
	LogFilePath                  string
	SuccessfulIterationDurations progress.IterationDurationsSnapshot
	FailedIterationDurations     progress.IterationDurationsSnapshot
	IterationsStarted            uint64
	Duration                     time.Duration
	SuccessfulIterationCount     uint64
	Iterations                   uint64
	FailedIterationCount         uint64
	DroppedIterationCount        uint64
	Failed                       bool
}

func (d ResultData) Log(logger *slog.Logger) {
	stats := log.IterationStatsGroup(
		d.IterationsStarted,
		d.SuccessfulIterationCount,
		d.FailedIterationCount,
		d.DroppedIterationCount,
		d.Duration,
	)

	if d.Failed {
		if d.Error != nil {
			logger.Error("Load Test Failed", log.ErrorAttr(d.Error), stats)
		} else {
			logger.Error("Load Test Failed", stats)
		}
	} else {
		logger.Info("Load Test Passed", stats)
	}
}

func (v *Views) Result(data ResultData) *ViewContext[ResultData] {
	return &ViewContext[ResultData]{
		view: v.result,
		data: data,
	}
}
