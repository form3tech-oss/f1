package staged

import (
	"time"

	log "github.com/sirupsen/logrus"
)

type rateCalculator struct {
	current int
	stages  []stage
	start   time.Time
}

func newRateCalculator(stages []stage) *rateCalculator {
	calculator := rateCalculator{
		current: -1,
	}
	calculator.addRange(stages)
	return &calculator
}

func (s *rateCalculator) addRange(stages []stage) {
	for _, stage := range stages {
		log.Debugf("Stage %d: %s, %d -> %d\n", stage, stage.duration.String(), stage.startTarget, stage.endTarget)
		s.add(stage)
	}
}

func (s *rateCalculator) add(newStage stage) {
	if len(s.stages) == 0 {
		newStage.startTarget = 0
	} else {
		newStage.startTarget = s.stages[len(s.stages)-1].endTarget
	}
	s.stages = append(s.stages, newStage)
}

func (s *rateCalculator) Rate(now time.Time) int {
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

func (s *rateCalculator) MaxDuration() time.Duration {
	max := 0 * time.Second
	for _, stage := range s.stages {
		max += stage.duration
	}
	return max
}
