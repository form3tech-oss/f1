package views

import (
	"log/slog"
	"os"
	"text/template"

	"github.com/mattn/go-isatty"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

type ViewContext[T log.Loggable] struct {
	view *View
	data T
}

func (vc *ViewContext[T]) Render() string {
	return render(vc.view.Template(), vc.data)
}

func (vc *ViewContext[T]) Print(printer *ui.Printer) {
	printer.Println(vc.Render())
}

func (vc *ViewContext[T]) Log(logger *slog.Logger) {
	vc.data.Log(logger)
}

type Views struct {
	start                *View
	result               *View
	setup                *View
	progress             *View
	teardown             *View
	timeout              *View
	maxIterationsReached *View
	interrupt            *View
}

type View struct {
	tty   *template.Template
	notty *template.Template
}

func (v *View) Template() *template.Template {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		return v.tty
	}

	return v.notty
}

func New() *Views {
	tty := parseTemplates(renderTermColorsEnabled)
	notty := parseTemplates(renderTermColorsDisabled)

	return &Views{
		start: &View{
			tty:   tty.start,
			notty: notty.start,
		},
		result: &View{
			tty:   tty.result,
			notty: notty.result,
		},
		setup: &View{
			tty:   tty.setup,
			notty: notty.setup,
		},
		timeout: &View{
			tty:   tty.timeout,
			notty: notty.timeout,
		},
		progress: &View{
			tty:   tty.progress,
			notty: notty.progress,
		},
		teardown: &View{
			tty:   tty.teardown,
			notty: notty.teardown,
		},
		maxIterationsReached: &View{
			tty:   tty.maxIterationsReached,
			notty: notty.maxIterationsReached,
		},
		interrupt: &View{
			tty:   tty.interrupt,
			notty: notty.interrupt,
		},
	}
}
