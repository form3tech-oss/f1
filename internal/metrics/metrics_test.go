package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
)

func TestMetrics_Init_IsSafe(t *testing.T) {
	t.Parallel()
	metrics.InitWithStaticMetrics(true, map[string]string{
		"product": "fps",
		"f1_id":   "myid",
	}) // race detector assertion
	for range 10 {
		go func() {
			metrics.InitWithStaticMetrics(true, map[string]string{
				"product": "fps",
				"f1_id":   "myid",
			})
		}()
	}
	assert.True(t, metrics.Instance().IterationMetricsEnabled)
	metrics.Instance().RecordIterationResult("test1", metrics.SuccessResult, 1)
	assert.Equal(t, 1, testutil.CollectAndCount(metrics.Instance().Iteration, "form3_loadtest_iteration"))
	o, err := metrics.Instance().Iteration.MetricVec.GetMetricWith(prometheus.Labels{
		metrics.TestNameLabel: "test1",
		metrics.StageLabel:    metrics.IterationStage,
		metrics.ResultLabel:   metrics.SuccessResult.String(),
		"product":             "fps",
		"f1_id":               "myid",
	})
	require.NoError(t, err)
	assert.Contains(t, o.Desc().String(), "product")
	assert.Contains(t, o.Desc().String(), "f1_id")
}
