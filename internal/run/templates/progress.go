package templates

import (
	"time"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
)

//nolint:lll // templates read better with long lines
const progressTemplate = `{cyan}[{{durationSeconds .Duration | printf "%5s"}}]{-}  {green}✔ {{printf "%5d" .SuccessfulIterationCount}}{-}  {{if .DroppedIterationCount}}{yellow}⦸ {{printf "%5d" .DroppedIterationCount}}{-}  {{end}}{red}✘ {{printf "%5d" .FailedIterationCount}}{-} {light_black}({{rate .RecentDuration .RecentSuccessfulIterations}}/s){-}
{{- with .SuccessfulIterationDurations}}   p(50): {{.Get 0.5}},  p(95): {{.Get 0.95}}, p(100): {{.Get 1.0}}{{end}}`

type ProgressData struct {
	SuccessfulIterationDurations metrics.DurationPercentileMap
	Duration                     time.Duration
	SuccessfulIterationCount     uint64
	DroppedIterationCount        uint64
	FailedIterationCount         uint64
	RecentDuration               time.Duration
	RecentSuccessfulIterations   uint64
}

func (t *Templates) Progress(data ProgressData) string {
	return render(t.progress, data)
}
