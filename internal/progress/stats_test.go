package progress_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/form3tech-oss/f1/v2/internal/progress"
)

func TestSnapshot_FailedIterationsRate_NoIterations(t *testing.T) {
	t.Parallel()

	snapshot := progress.Snapshot{}

	assert.Zero(t, snapshot.FailedIterationsRate())
}
