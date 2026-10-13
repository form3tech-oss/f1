package rate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func ParseRate(rateArg string) (int, time.Duration, error) {
	var rate int
	var unit time.Duration

	if strings.Contains(rateArg, "/") {
		var err error
		rate, err = strconv.Atoi(rateArg[0:strings.Index(rateArg, "/")])
		if err != nil {
			return rate, unit, fmt.Errorf("unable to parse rate %s: %w", rateArg, err)
		}
		if rate < 0 {
			return rate, unit, fmt.Errorf("rate %s can't be negative", rateArg)
		}
		unitArg := rateArg[strings.Index(rateArg, "/")+1:]
		// allow a bare unit such as "s" or "ms" to mean one of that unit
		if first, _ := utf8.DecodeRuneInString(unitArg); unicode.IsLetter(first) {
			unitArg = "1" + unitArg
		}
		unit, err = time.ParseDuration(unitArg)
		if err != nil {
			return rate, unit, fmt.Errorf("unable to parse unit %s: %w", rateArg, err)
		}
		if unit <= 0 {
			return rate, unit, fmt.Errorf("unit %s must be positive", rateArg)
		}
	} else {
		var err error
		rate, err = strconv.Atoi(rateArg)
		if err != nil {
			return rate, unit, fmt.Errorf("unable to parse rate %s: %w", rateArg, err)
		}
		if rate < 0 {
			return rate, unit, fmt.Errorf("rate %s can't be negative", rateArg)
		}
		unit = 1 * time.Second
	}

	return rate, unit, nil
}
