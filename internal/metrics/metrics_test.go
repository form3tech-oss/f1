package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
)

func TestMetrics_Init_IsSafe(t *testing.T) {
	t.Parallel()

	metrics.Init(true)

	// race detector assertion
	for range 10 {
		go func() {
			metrics.Init(false)
		}()
	}

	assert.True(t, metrics.Instance().IterationMetricsEnabled)
}
