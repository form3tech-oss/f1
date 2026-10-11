package rate_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v2/internal/trigger/rate"
)

func TestParseRate(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		rateArg      string
		expectedRate int
		expectedUnit time.Duration
	}{
		{rateArg: "10", expectedRate: 10, expectedUnit: time.Second},
		{rateArg: "10/s", expectedRate: 10, expectedUnit: time.Second},
		{rateArg: "10/100ms", expectedRate: 10, expectedUnit: 100 * time.Millisecond},
		{rateArg: "0/1m", expectedRate: 0, expectedUnit: time.Minute},
	} {
		t.Run(test.rateArg, func(t *testing.T) {
			t.Parallel()

			parsedRate, unit, err := rate.ParseRate(test.rateArg)

			require.NoError(t, err)
			require.Equal(t, test.expectedRate, parsedRate)
			require.Equal(t, test.expectedUnit, unit)
		})
	}
}

func TestParseRate_Errors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		rateArg, expectedError string
	}{
		{rateArg: "abc", expectedError: "unable to parse rate abc"},
		{rateArg: "-1", expectedError: "rate -1 can't be negative"},
		{rateArg: "-1/s", expectedError: "rate -1/s can't be negative"},
		{rateArg: "10/", expectedError: "unable to parse unit 10/: missing unit"},
		{rateArg: "10/xyz", expectedError: "unable to parse unit 10/xyz"},
		{rateArg: "10/0", expectedError: "unit 10/0 must be positive"},
		{rateArg: "10/0s", expectedError: "unit 10/0s must be positive"},
	} {
		t.Run(test.rateArg, func(t *testing.T) {
			t.Parallel()

			_, _, err := rate.ParseRate(test.rateArg)

			require.ErrorContains(t, err, test.expectedError)
		})
	}
}
