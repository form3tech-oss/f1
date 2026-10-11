package progress_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/progress"
)

func TestSnapshot_FailedIterationsRate(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name                         string
		successful, failed, dropped  uint64
		expectedFailedIterationsRate uint64
	}{
		{name: "no iterations", expectedFailedIterationsRate: 0},
		{name: "all failed", failed: 4, expectedFailedIterationsRate: 100},
		{name: "quarter failed", successful: 2, failed: 1, dropped: 1, expectedFailedIterationsRate: 25},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			snapshot := progress.Snapshot{
				SuccessfulIterationDurations: progress.IterationDurationsSnapshot{Count: test.successful},
				FailedIterationDurations:     progress.IterationDurationsSnapshot{Count: test.failed},
				DroppedIterationCount:        test.dropped,
			}

			require.Equal(t, test.expectedFailedIterationsRate, snapshot.FailedIterationsRate())
		})
	}
}
