package views

import (
	"math"
	"strings"
	"text/template"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/termcolor"
)

type renderTermColorsType bool

const (
	renderTermColorsEnabled  renderTermColorsType = true
	renderTermColorsDisabled renderTermColorsType = false
)

type templates struct {
	start                *template.Template
	result               *template.Template
	setup                *template.Template
	progress             *template.Template
	teardown             *template.Template
	timeout              *template.Template
	maxIterationsReached *template.Template
	interrupt            *template.Template
}

func parseTemplates(renderTermColors renderTermColorsType) *templates {
	templateFunctions := template.FuncMap{
		"rate": func(duration time.Duration, count uint64) uint64 {
			durationInSeconds := duration.Round(time.Second).Seconds()

			if durationInSeconds == 0 {
				return 0
			}

			return uint64(math.Round(float64(count) / durationInSeconds))
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

	replacements := termReplacements(renderTermColors)

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

	return &templates{
		start:                start,
		result:               result,
		setup:                setup,
		progress:             progress,
		teardown:             teardown,
		timeout:              timeout,
		maxIterationsReached: maxIterationsReached,
		interrupt:            interrupt,
	}
}

func render(t *template.Template, data any) string {
	var builder strings.Builder
	err := t.Execute(&builder, data)
	if err != nil {
		panic(err)
	}
	return builder.String()
}

func applyReplacements(templateString string, replacements map[string]string) string {
	res := templateString
	for replacement, value := range replacements {
		res = strings.ReplaceAll(res, replacement, value)
	}
	return res
}

func termReplacements(renderFlag renderTermColorsType) map[string]string {
	if renderFlag == renderTermColorsEnabled {
		return map[string]string{
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
	}

	return map[string]string{
		"{-}":              "",
		"{u}":              "",
		"{bold}":           "",
		"{cyan}":           "",
		"{yellow}":         "",
		"{red}":            "",
		"{green}":          "",
		"{intensive_blue}": "",
		"{light_black}":    "",
	}
}
