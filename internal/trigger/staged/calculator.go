package staged

import (
	"time"
)

type RateCalculator struct {
	start   time.Time
	stages  []Stage
	current int
}

func NewRateCalculator(stages []Stage, start *time.Time) *RateCalculator {
	calculator := RateCalculator{
		current: -1,
	}
	calculator.addRange(stages)
	if start != nil {
		calculator.start = *start
	}
	return &calculator
}

func (s *RateCalculator) addRange(stages []Stage) {
	for _, stage := range stages {
		s.add(stage)
	}
}

func (s *RateCalculator) add(newStage Stage) {
	if len(s.stages) == 0 {
		newStage.StartTarget = 0
	} else {
		newStage.StartTarget = s.stages[len(s.stages)-1].EndTarget
	}
	s.stages = append(s.stages, newStage)
}

func (s *RateCalculator) Rate(now time.Time) int {
	if s.current < 0 {
		s.current = 0
		if s.start.IsZero() {
			s.start = now
		}
	}
	if s.current > len(s.stages)-1 {
		return 0
	}
	for s.current < len(s.stages) && now.Sub(s.start)+1 > s.stages[s.current].Duration {
		s.start = s.start.Add(s.stages[s.current].Duration)
		s.current++
	}
	if s.current > len(s.stages)-1 {
		return 0
	}
	// interpolate
	offset := now.Sub(s.start)
	position := float64(offset) / float64(s.stages[s.current].Duration)
	rate := s.stages[s.current].StartTarget +
		int(position*float64(s.stages[s.current].EndTarget-s.stages[s.current].StartTarget))
	return rate
}

func (s *RateCalculator) MaxDuration() time.Duration {
	max := 0 * time.Second
	for _, stage := range s.stages {
		max += stage.Duration
	}
	return max
}
