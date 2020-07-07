package staged

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/pflag"
)

type stage struct {
	startTarget int
	endTarget   int
	duration    time.Duration
}

type stages struct {
	current int
	stages  []stage
	start   time.Time
}

func (s *stages) add(newStage stage) {
	if len(s.stages) == 0 {
		newStage.startTarget = 0
	} else {
		newStage.startTarget = s.stages[len(s.stages)-1].endTarget
	}
	s.stages = append(s.stages, newStage)
}

func (s *stages) Rate(now time.Time) int {
	if s.current < 0 {
		s.current = 0
		s.start = now
	}
	if s.current > len(s.stages)-1 {
		return 0
	}
	if now.Sub(s.start)+1 > s.stages[s.current].duration {
		s.start = s.start.Add(s.stages[s.current].duration)
		s.current++
	}
	if s.current > len(s.stages)-1 {
		return 0
	}
	// interpolate
	offset := now.Sub(s.start)
	position := float64(offset) / float64(s.stages[s.current].duration)
	rate := s.stages[s.current].startTarget + int(position*float64(s.stages[s.current].endTarget-s.stages[s.current].startTarget))
	log.Debugf("Stage %d; Triggering %d. Offset: %d, Duration: %d, Position: %v\n", s.current, rate, offset, s.stages[s.current].duration, position)
	return rate
}

func (s *stages) MaxDuration() time.Duration {
	max := 0 * time.Second
	for _, stage := range s.stages {
		max += stage.duration
	}
	return max
}

func StagedRate() api.Builder {
	flags := pflag.NewFlagSet("staged", pflag.ContinueOnError)
	flags.StringP("stages", "s", "0s:1; 10s:1", "Semicolon separated list of <stage_duration>:<target_concurrent_iterations>. During the stage, the number of concurrent iterations will ramp up or down to the target. ")
	flags.DurationP("iterationFrequency", "f", 1*time.Second, "How frequently iterations should be started")
	flags.Float64P("jitter", "j", 0.0, "vary the rate randomly by up to jitter percent")

	return api.Builder{
		Name:        "staged",
		Description: "triggers iterations at varying rates",
		Flags:       flags,
		New: func(params *pflag.FlagSet) (*api.Trigger, error) {

			jitterArg, err := params.GetFloat64("jitter")
			if err != nil {
				return nil, err
			}
			stg, err := params.GetString("stages")
			if err != nil {
				return nil, err
			}
			frequency, err := params.GetDuration("iterationFrequency")
			if err != nil {
				return nil, err
			}

			stages := stages{
				current: -1,
			}
			for i, stageDefinition := range strings.Split(stg, ",") {
				stageDefinition = strings.TrimSpace(stageDefinition)
				split := strings.Split(stageDefinition, ":")
				if len(split) != 2 {
					return nil, fmt.Errorf("unable to parse stage %d: `%s` from `%s`", i, stageDefinition, stg)
				}
				duration, err := time.ParseDuration(strings.TrimSpace(split[0]))
				if err != nil {
					return nil, fmt.Errorf("unable to parse duration %s in stage %d: %s", split[0], i, stageDefinition)
				}
				target, err := strconv.Atoi(strings.TrimSpace(split[1]))
				if err != nil {
					return nil, fmt.Errorf("unable to parse target %s in stage %d: %s", split[1], i, stageDefinition)
				}
				stages.add(stage{
					endTarget: target,
					duration:  duration,
				})
			}

			for i, s := range stages.stages {
				log.Debugf("Stage %d: %s, %d -> %d\n", i, s.duration.String(), s.startTarget, s.endTarget)
			}
			return &api.Trigger{
					Trigger:     api.NewIterationWorker(frequency, api.WithJitter(stages.Rate, jitterArg)),
					DryRun:      api.WithJitter(stages.Rate, jitterArg),
					Description: fmt.Sprintf("Starting iterations every %s in numbers varying by time: %s,", frequency, stg),
					Duration:    stages.MaxDuration(),
				},
				nil
		},
	}
}
