package staged_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/trigger/staged"
)

func TestRate_RejectsNonPositiveIterationFrequency(t *testing.T) {
	t.Parallel()

	for _, frequency := range []string{"0s", "-1s"} {
		t.Run(frequency, func(t *testing.T) {
			t.Parallel()

			builder := staged.Rate()
			require.NoError(t, builder.Flags.Parse([]string{"--iterationFrequency", frequency}))

			_, err := builder.New(builder.Flags)

			require.EqualError(t, err, "--iterationFrequency "+frequency+" must be positive")
		})
	}
}
