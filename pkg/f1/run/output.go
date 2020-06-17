package run

import (
	"strings"
	"text/template"
	"time"

	"github.com/hako/durafmt"
	"github.com/workanator/go-ataman"
)

var (
	templateFunctions = template.FuncMap{
		"add": func(i, j uint64) uint64 {
			return i + j
		},
		"rate": func(duration time.Duration, count uint64) uint64 {
			if uint64(duration/time.Second) == 0 {
				return 0
			}
			return count / uint64(duration/time.Second)
		},
		"durationSeconds": func(t time.Duration) time.Duration {
			return t.Round(time.Second)
		},
		"duration": func(t time.Duration) string {
			return durafmt.Parse(t).String()
		},
		"percent": func(val, total uint64) float64 {
			return 100.0 * float64(val) / float64(total)
		},
	}
)

func renderTemplate(t *template.Template, r interface{}) string {
	var builder strings.Builder
	err := t.Execute(&builder, r)
	if err != nil {
		panic(err)
	}
	rndr := ataman.NewRenderer(ataman.CurlyStyle())
	return rndr.MustRenderf(builder.String())
}
