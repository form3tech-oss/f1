package templates

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
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
{bold}Successful Iterations:{-} {green}{{.SuccessfulIterationCount}} ({{percent .SuccessfulIterationCount .Iterations | printf "%0.2f"}}%, {{rate .Duration .SuccessfulIterationCount}}/second){-} {{.SuccessfulIterationDurations.String}}
{{- end}}
{{- if .FailedIterationCount}}
{bold}Failed Iterations:{-} {red}{{.FailedIterationCount}} ({{percent .FailedIterationCount .Iterations | printf "%0.2f"}}%, {{rate .Duration .FailedIterationCount}}){-} {{.FailedIterationDurations.String}}
{{- end}}
{{- if .DroppedIterationCount}}
{bold}Dropped Iterations:{-} {yellow}{{.DroppedIterationCount}} ({{percent .DroppedIterationCount .Iterations | printf "%0.2f"}}%, {{rate .Duration .DroppedIterationCount}}){-} (consider increasing --concurrency setting)
{{- end}}
{bold}Full logs:{-} {{.LogFile}}
`

type ResultData struct {
	SuccessfulIterationDurations metrics.DurationPercentileMap
	FailedIterationDurations     metrics.DurationPercentileMap
	Error                        error
	LogFile                      string
	IterationsStarted            uint64
	Duration                     time.Duration
	SuccessfulIterationCount     uint64
	Iterations                   uint64
	FailedIterationCount         uint64
	DroppedIterationCount        uint64
	Failed                       bool
}

func (t *Templates) Result(data ResultData) string {
	return render(t.result, data)
}
