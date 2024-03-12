package run

import (
	"strings"
	"text/template"
)

func renderTemplate(t *template.Template, r interface{}) string {
	var builder strings.Builder
	err := t.Execute(&builder, r)
	if err != nil {
		panic(err)
	}
	return builder.String()
}
