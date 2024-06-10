package staged_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/trigger/staged"
)

func TestParseStages_With_Valid_String(t *testing.T) {
	t.Parallel()

	value := "0s:1,10s:1,20s:20,1m:50,1h:200"
	expected := []staged.Stage{
		{
			StartTarget: 0,
			EndTarget:   1,
			Duration:    0 * time.Second,
		},
		{
			StartTarget: 0,
			EndTarget:   1,
			Duration:    10 * time.Second,
		},
		{
			StartTarget: 0,
			EndTarget:   20,
			Duration:    20 * time.Second,
		},
		{
			StartTarget: 0,
			EndTarget:   50,
			Duration:    1 * time.Minute,
		},
		{
			StartTarget: 0,
			EndTarget:   200,
			Duration:    1 * time.Hour,
		},
	}
	actual, err := staged.ParseStages(value)

	require.NoError(t, err)
	assert.ElementsMatch(t, actual, expected)
}

func TestParseStages_Error_Too_Many_Elements(t *testing.T) {
	t.Parallel()

	value := "0s:1:2"
	stages, err := staged.ParseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}

func TestParseStages_Error_Bad_Duration(t *testing.T) {
	t.Parallel()

	value := "0BB:1"
	stages, err := staged.ParseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}

func TestParseStages_Error_Bad_Target(t *testing.T) {
	t.Parallel()

	value := "1s:BB"
	stages, err := staged.ParseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}
