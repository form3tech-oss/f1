package templates

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/progress"
)

//nolint:lll // templates read better with long lines
const progressTemplate = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]{-}  {green}✔ {{printf "%5d" .SuccessfulIterationCount}}{-}  {{if .DroppedIterationCount}}{yellow}⦸ {{printf "%5d" .DroppedIterationCount}}{-}  {{end}}{red}✘ {{printf "%5d" .FailedIterationCount}}{-} {light_black}({{rate .Period .SuccessfulIterationDurationsForPeriod.Count}}/s){-}   {{.SuccessfulIterationDurationsForPeriod}}`

type ProgressData struct {
	SuccessfulIterationDurationsForPeriod progress.IterationDurationsSnapshot
	Duration                              time.Duration
	SuccessfulIterationCount              uint64
	DroppedIterationCount                 uint64
	FailedIterationCount                  uint64
	Period                                time.Duration
}

func (t *Templates) Progress(data ProgressData) string {
	return render(t.progress, data)
}
