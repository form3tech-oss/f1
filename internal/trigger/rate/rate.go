package rate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseRate parses a rate such as "10", "10/s", "10/100ms" or "10/.5s" into the number of
// iterations and the period they are spread over. The period defaults to one second.
func ParseRate(rateArg string) (int, time.Duration, error) {
	rateValue, unitArg, hasUnit := strings.Cut(rateArg, "/")

	rate, err := strconv.Atoi(rateValue)
	if err != nil {
		return rate, 0, fmt.Errorf("unable to parse rate %s: %w", rateArg, err)
	}
	if rate < 0 {
		return rate, 0, fmt.Errorf("rate %s can't be negative", rateArg)
	}

	if !hasUnit {
		return rate, time.Second, nil
	}

	if unitArg == "" {
		return rate, 0, fmt.Errorf("unable to parse unit %s: missing unit", rateArg)
	}
	// allow a bare unit such as "s" or "ms" to mean one of that unit
	if first := unitArg[0]; first != '.' && (first < '0' || first > '9') {
		unitArg = "1" + unitArg
	}
	unit, err := time.ParseDuration(unitArg)
	if err != nil {
		return rate, unit, fmt.Errorf("unable to parse unit %s: %w", rateArg, err)
	}
	if unit <= 0 {
		return rate, unit, fmt.Errorf("unit %s must be positive", rateArg)
	}

	return rate, unit, nil
}
