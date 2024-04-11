package staged

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStages_With_Valid_String(t *testing.T) {
	t.Parallel()

	value := "0s:1,10s:1,20s:20,1m:50,1h:200"
	expected := []stage{
		{
			startTarget: 0,
			endTarget:   1,
			duration:    0 * time.Second,
		},
		{
			startTarget: 0,
			endTarget:   1,
			duration:    10 * time.Second,
		},
		{
			startTarget: 0,
			endTarget:   20,
			duration:    20 * time.Second,
		},
		{
			startTarget: 0,
			endTarget:   50,
			duration:    1 * time.Minute,
		},
		{
			startTarget: 0,
			endTarget:   200,
			duration:    1 * time.Hour,
		},
	}
	actual, err := parseStages(value)

	require.NoError(t, err)
	assert.ElementsMatch(t, actual, expected)
}

func TestParseStages_Error_Too_Many_Elements(t *testing.T) {
	t.Parallel()

	value := "0s:1:2"
	stages, err := parseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}

func TestParseStages_Error_Bad_Duration(t *testing.T) {
	t.Parallel()

	value := "0BB:1"
	stages, err := parseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}

func TestParseStages_Error_Bad_Target(t *testing.T) {
	t.Parallel()

	value := "1s:BB"
	stages, err := parseStages(value)
	require.Error(t, err)
	assert.Nil(t, stages)
}
