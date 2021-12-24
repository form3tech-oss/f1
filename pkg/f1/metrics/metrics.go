package metrics

import (
	internal_metrics "github.com/form3tech-oss/f1/v2/internal/metrics"
)

func GetMetrics() *internal_metrics.Metrics {
	return internal_metrics.Instance()
}
