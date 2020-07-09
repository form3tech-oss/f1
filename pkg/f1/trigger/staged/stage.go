package staged

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type stage struct {
	startTarget int
	endTarget   int
	duration    time.Duration
}

func parseStages(value string) ([]stage, error) {

	var stages []stage

	stageElements := strings.Split(value, ",")

	for i, stageElements := range stageElements {

		stageElement := strings.Split(strings.TrimSpace(stageElements), ":")
		if len(stageElement) != 2 {
			return nil, fmt.Errorf("unable to parse stage %d: `%s` from `%s`", i, stageElements, value)
		}

		duration, err := time.ParseDuration(strings.TrimSpace(stageElement[0]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse duration %s in stage %d: %s", stageElement[0], i, stageElements)
		}

		target, err := strconv.Atoi(strings.TrimSpace(stageElement[1]))
		if err != nil {
			return nil, fmt.Errorf("unable to parse target %s in stage %d: %s", stageElement[1], i, stageElements)
		}

		stages = append(stages, stage{
			endTarget: target,
			duration:  duration,
		})
	}

	return stages, nil
}
