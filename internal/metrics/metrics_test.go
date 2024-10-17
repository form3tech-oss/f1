package metrics_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/metrics"
)

func TestMetrics_Init_IsSafe(t *testing.T) {
	t.Parallel()
	labels := map[string]string{
		"product":  "fps",
		"customer": "fake-customer",
		"f1_id":    "myid",
		"labelx":   "vx",
	}
	metrics.InitWithStaticMetrics(true, labels) // race detector assertion
	for range 10 {
		go func() {
			metrics.InitWithStaticMetrics(true, labels)
		}()
	}
	assert.True(t, metrics.Instance().IterationMetricsEnabled)
	metrics.Instance().RecordIterationResult("test1", metrics.SuccessResult, 1)
	assert.Equal(t, 1, testutil.CollectAndCount(metrics.Instance().Iteration, "form3_loadtest_iteration"))

	expected := `
        	     # HELP form3_loadtest_iteration Duration of iteration functions.
        	     # TYPE form3_loadtest_iteration summary
				`
	quantileFormat := `
				form3_loadtest_iteration{customer="fake-customer",f1_id="myid",labelx="vx",product="fps",result="success",stage="iteration",test="test1",quantile="%s"} 1
				`
	for _, quantile := range []string{"0.5", "0.75", "0.9", "0.95", "0.99", "0.9999", "1.0"} {
		expected += fmt.Sprintf(quantileFormat, quantile)
	}

	expected += `
        	      form3_loadtest_iteration_sum{customer="fake-customer",f1_id="myid",labelx="vx",product="fps",result="success",stage="iteration",test="test1"} 1
        	      form3_loadtest_iteration_count{customer="fake-customer",f1_id="myid",labelx="vx",product="fps",result="success",stage="iteration",test="test1"} 1
				`
	r := bytes.NewReader([]byte(expected))
	require.NoError(t, testutil.CollectAndCompare(metrics.Instance().Iteration, r))
}
