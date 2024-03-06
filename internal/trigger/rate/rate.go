package rate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseRate(rateArg string) (int, time.Duration, error) {
	var rate int
	var unit time.Duration

	if strings.Contains(rateArg, "/") {
		var err error
		rate, err = strconv.Atoi((rateArg)[0:strings.Index(rateArg, "/")])
		if err != nil {
			return rate, unit, fmt.Errorf("unable to parse rate %s: %w", rateArg, err)
		}
		if rate < 0 {
			return rate, unit, fmt.Errorf("invalid rate arg %s", rateArg)
		}
		unitArg := (rateArg)[strings.Index(rateArg, "/")+1:]
		if !isNumeric(unitArg[0:1]) {
			unitArg = "1" + unitArg
		}
		unit, err = time.ParseDuration(unitArg)
		if err != nil {
			return rate, unit, fmt.Errorf("unable to parse unit %s: %w", rateArg, err)
		}
	} else {
		var err error
		rate, err = strconv.Atoi(rateArg)
		if err != nil {
			return rate, unit, fmt.Errorf("unable to parse rate %s: %w", rateArg, err)
		}
		if rate < 0 {
			return rate, unit, fmt.Errorf("invalid rate arg %s", rateArg)
		}
		unit = 1 * time.Second
	}

	return rate, unit, nil
}

func isNumeric(value string) bool {
	re := regexp.MustCompile("^[0-9]+$")
	return re.MatchString(value)
}
