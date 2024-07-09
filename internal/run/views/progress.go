package views

import (
	"log/slog"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/log"
	"github.com/form3tech-oss/f1/v2/internal/progress"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

//nolint:lll // templates read better with long lines
const progressTemplate = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]{-}  {green}✔ {{printf "%5d" .SuccessfulIterationCount}}{-}  {{if .DroppedIterationCount}}{yellow}⦸ {{printf "%5d" .DroppedIterationCount}}{-}  {{end}}{red}✘ {{printf "%5d" .FailedIterationCount}}{-} {light_black}({{rate .Period .SuccessfulIterationDurationsForPeriod.Count}}/s){-}   {{.SuccessfulIterationDurationsForPeriod}}`

var _ ui.Outputable = (*ViewContext[ProgressData])(nil)

type ProgressData struct {
	SuccessfulIterationDurationsForPeriod progress.IterationDurationsSnapshot
	Duration                              time.Duration
	SuccessfulIterationCount              uint64
	DroppedIterationCount                 uint64
	FailedIterationCount                  uint64
	Period                                time.Duration
}

func (d ProgressData) Log(logger *slog.Logger) {
	logger.Info("progress", log.IterationStatsGroup(
		0,
		d.SuccessfulIterationCount,
		d.FailedIterationCount,
		d.DroppedIterationCount,
		d.Period,
	))
}

func (v *Views) Progress(data ProgressData) *ViewContext[ProgressData] {
	return &ViewContext[ProgressData]{
		view: v.progress,
		data: data,
	}
}
