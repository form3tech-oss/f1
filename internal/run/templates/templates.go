package templates

import (
	"strings"
	"text/template"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/termcolor"
)

//nolint:lll // templates read better with long lines
const (
	startTemplate = `{u}{bold}{intensive_blue}F1 Load Tester{-}
Running {yellow}{{.Options.Scenario}}{-} scenario for {{if .Options.MaxIterations}}up to {{.Options.MaxIterations}} iterations or up to {{end}}{{duration .Options.MaxDuration}} at a rate of {{.RateDescription}}.
`

	resultTemplate = `
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

	progressTemplate = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]{-}  {green}✔ {{printf "%5d" .SuccessfulIterationCount}}{-}  {{if .DroppedIterationCount}}{yellow}⦸ {{printf "%5d" .DroppedIterationCount}}{-}  {{end}}{red}✘ {{printf "%5d" .FailedIterationCount}}{-} {light_black}({{rate .RecentDuration .RecentSuccessfulIterations}}/s){-}
{{- with .SuccessfulIterationDurations}}   p(50): {{.Get 0.5}},  p(95): {{.Get 0.95}}, p(100): {{.Get 1.0}}{{end}}`

	setupTemplate    = `{cyan}[Setup]{-}    {{if .Error}}{red}✘ {{.Error}}{-}{{else}}{green}✔{-}{{end}}`
	teardownTemplate = `{cyan}[Teardown]{-} {{if .Error}}{red}✘ {{.Error}}{-}{{else}}{green}✔{-}{{end}}`

	timeoutTemplate              = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Max Duration Elapsed - waiting for active tests to complete{-}`
	maxIterationsReachedTemplate = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Max Iterations Reached - waiting for active tests to complete{-}`
	interruptTemplate            = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]  Interrupted - waiting for active tests to complete{-}`
)

type Templates struct {
	Start                *template.Template
	Result               *template.Template
	Setup                *template.Template
	Progress             *template.Template
	Teardown             *template.Template
	Timeout              *template.Template
	MaxIterationsReached *template.Template
	Interrupt            *template.Template
}

func applyReplacements(templateString string, replacements map[string]string) string {
	res := templateString
	for replacement, value := range replacements {
		res = strings.ReplaceAll(res, replacement, value)
	}
	return res
}

func Parse() *Templates {
	templateFunctions := template.FuncMap{
		"add": func(i, j uint64) uint64 {
			return i + j
		},
		"rate": func(duration time.Duration, count uint64) uint64 {
			if uint64(duration/time.Second) == 0 {
				return 0
			}
			return count / uint64(duration/time.Second)
		},
		"durationSeconds": func(d time.Duration) time.Duration {
			return d.Round(time.Second)
		},
		"duration": func(d time.Duration) string {
			return d.String()
		},
		"percent": func(val, total uint64) float64 {
			return 100.0 * float64(val) / float64(total)
		},
	}

	replacements := map[string]string{
		"{-}":              termcolor.Reset,
		"{u}":              termcolor.Underline,
		"{bold}":           termcolor.Bold,
		"{cyan}":           termcolor.Cyan,
		"{yellow}":         termcolor.Yellow,
		"{red}":            termcolor.Red,
		"{green}":          termcolor.Green,
		"{intensive_blue}": termcolor.BrightBlue,
		"{light_black}":    termcolor.BrightBlack,
	}

	start := template.Must(template.New("start").
		Funcs(templateFunctions).
		Parse(applyReplacements(startTemplate, replacements)))

	result := template.Must(template.New("result").
		Funcs(templateFunctions).
		Parse(applyReplacements(resultTemplate, replacements)))

	progress := template.Must(template.New("progress").
		Funcs(templateFunctions).
		Parse(applyReplacements(progressTemplate, replacements)))

	setup := template.Must(template.New("setup").
		Funcs(templateFunctions).
		Parse(applyReplacements(setupTemplate, replacements)))

	teardown := template.Must(template.New("teardown").
		Funcs(templateFunctions).
		Parse(applyReplacements(teardownTemplate, replacements)))

	timeout := template.Must(template.New("timeout").
		Funcs(templateFunctions).
		Parse(applyReplacements(timeoutTemplate, replacements)))

	maxIterationsReached := template.Must(template.New("maxIterationsReached").
		Funcs(templateFunctions).
		Parse(applyReplacements(maxIterationsReachedTemplate, replacements)))

	interrupt := template.Must(template.New("interrupt").
		Funcs(templateFunctions).
		Parse(applyReplacements(interruptTemplate, replacements)))

	return &Templates{
		Start:                start,
		Result:               result,
		Setup:                setup,
		Progress:             progress,
		Teardown:             teardown,
		Timeout:              timeout,
		MaxIterationsReached: maxIterationsReached,
		Interrupt:            interrupt,
	}
}
