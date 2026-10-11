package staged

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Stage struct {
	StartTarget int
	EndTarget   int
	Duration    time.Duration
}

func ParseStages(value string) ([]Stage, error) {
	stageElements := strings.Split(value, ",")
	stages := make([]Stage, len(stageElements))

	for i, stageElements := range stageElements {
		stageElement := strings.Split(strings.TrimSpace(stageElements), ":")
		if len(stageElement) != 2 {
			return nil, fmt.Errorf("unable to parse stage %d: `%s` from `%s`", i, stageElements, value)
		}

		duration, err := time.ParseDuration(strings.TrimSpace(stageElement[0]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse duration %s in stage %d: %s", stageElement[0], i, stageElements)
		}
		if duration < 0 {
			return nil, fmt.Errorf("duration %s in stage %d can't be negative", duration, i)
		}

		target, err := strconv.Atoi(strings.TrimSpace(stageElement[1]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse target %s in stage %d: %s", stageElement[1], i, stageElements)
		}
		if target < 0 {
			return nil, fmt.Errorf("target %d in stage %d can't be negative", target, i)
		}

		stages[i] = Stage{
			EndTarget: target,
			Duration:  duration,
		}
	}

	return stages, nil
}
