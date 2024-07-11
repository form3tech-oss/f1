package metrics

import (
	internal_metrics "github.com/form3tech-oss/f1/v2/internal/metrics"
)

// Deprecated: internal metrics will not be exposed in future versions
func GetMetrics() *internal_metrics.Metrics {
	return internal_metrics.Instance()
}
