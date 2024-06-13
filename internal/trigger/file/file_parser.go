package file

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/form3tech-oss/f1/v2/internal/trigger/constant"
	"github.com/form3tech-oss/f1/v2/internal/trigger/gaussian"
	"github.com/form3tech-oss/f1/v2/internal/trigger/ramp"
	"github.com/form3tech-oss/f1/v2/internal/trigger/staged"
)

type ConfigFile struct {
	Scenario *string  `yaml:"scenario"`
	Default  Stage    `yaml:"default"`
	Limits   Limits   `yaml:"limits"`
	Schedule Schedule `yaml:"schedule"`
	Stages   []Stage  `yaml:"stages"`
}

type Schedule struct {
	StageStart *time.Time `yaml:"stage-start"`
}

type Limits struct {
	MaxDuration     *time.Duration `yaml:"max-duration"`
	Concurrency     *int           `yaml:"concurrency"`
	MaxIterations   *uint64        `yaml:"max-iterations"`
	MaxFailures     *uint64        `yaml:"max-failures"`
	MaxFailuresRate *int           `yaml:"max-failures-rate"`
	IgnoreDropped   *bool          `yaml:"ignore-dropped"`
}

type Stage struct {
	Mode               *string            `yaml:"mode"`
	StartRate          *string            `yaml:"start-rate"`
	EndRate            *string            `yaml:"end-rate"`
	Rate               *string            `yaml:"rate"`
	Distribution       *string            `yaml:"distribution"`
	Weights            *string            `yaml:"weights"`
	Stages             *string            `yaml:"stages"`
	Concurrency        *int               `yaml:"concurrency"`
	Jitter             *float64           `yaml:"jitter"`
	Volume             *float64           `yaml:"volume"`
	Duration           *time.Duration     `yaml:"duration"`
	IterationFrequency *time.Duration     `yaml:"iteration-frequency"`
	Repeat             *time.Duration     `yaml:"repeat"`
	Peak               *time.Duration     `yaml:"peak"`
	StandardDeviation  *time.Duration     `yaml:"standard-deviation"`
	Parameters         *map[string]string `yaml:"parameters"`
}

func ParseConfigFile(fileContent []byte, now time.Time) (*RunnableStages, error) {
	configFile := ConfigFile{}
	err := yaml.Unmarshal(fileContent, &configFile)
	if err != nil {
		return nil, fmt.Errorf("parsing config file as yaml: %w", err)
	}
	validatedConfigFile, err := configFile.validateCommonFields()
	if err != nil {
		return nil, err
	}

	var stages []runnableStage
	stagesTotalDuration := 0 * time.Second
	for idx, stageConfig := range validatedConfigFile.Stages {
		validatedStage, err := stageConfig.validateCommonFieldsOfStage(idx, validatedConfigFile.Default)
		if err != nil {
			return nil, err
		}
		stagesTotalDuration += *validatedStage.Duration

		stageStart := validatedConfigFile.Schedule.StageStart
		if stageStart == nil || stageStart.Add(stagesTotalDuration).After(now) {
			parsedStage, err := validatedStage.parseStage(idx, validatedConfigFile.Default)
			if err != nil {
				return nil, err
			}
			stages = append(stages, *parsedStage)
		}
	}

	return &RunnableStages{
		Scenario:            *validatedConfigFile.Scenario,
		Stages:              stages,
		stagesTotalDuration: stagesTotalDuration,
		MaxDuration:         *validatedConfigFile.Limits.MaxDuration,
		Concurrency:         *validatedConfigFile.Limits.Concurrency,
		MaxIterations:       *validatedConfigFile.Limits.MaxIterations,
		maxFailures:         *validatedConfigFile.Limits.MaxFailures,
		maxFailuresRate:     *validatedConfigFile.Limits.MaxFailuresRate,
		IgnoreDropped:       *validatedConfigFile.Limits.IgnoreDropped,
	}, nil
}

func (s *Stage) parseStage(stageIdx int, defaults Stage) (*runnableStage, error) {
	switch *s.Mode {
	case "constant":
		validatedConstantStage, err := s.validateConstantStage(stageIdx, defaults)
		if err != nil {
			return nil, fmt.Errorf("validating constant stage: %w", err)
		}
		rates, err := constant.CalculateConstantRate(
			*validatedConstantStage.Jitter,
			*validatedConstantStage.Rate,
			*validatedConstantStage.Distribution,
		)
		if err != nil {
			return nil, fmt.Errorf("calculating constant rate: %w", err)
		}

		return &runnableStage{
			StageDuration:     *validatedConstantStage.Duration,
			IterationDuration: rates.IterationDuration,
			Rate:              rates.Rate,
			Params:            *validatedConstantStage.Parameters,
		}, nil
	case "ramp":
		validatedRampStage, err := s.validateRampStage(stageIdx, defaults)
		if err != nil {
			return nil, fmt.Errorf("validating ramp stage: %w", err)
		}
		rates, err := ramp.CalculateRampRate(
			*validatedRampStage.StartRate,
			*validatedRampStage.EndRate,
			*validatedRampStage.Distribution,
			*validatedRampStage.Duration,
			*validatedRampStage.Jitter,
		)
		if err != nil {
			return nil, fmt.Errorf("calculating ramp rate: %w", err)
		}

		return &runnableStage{
			StageDuration:     *validatedRampStage.Duration,
			IterationDuration: rates.IterationDuration,
			Rate:              rates.Rate,
			Params:            *validatedRampStage.Parameters,
		}, nil
	case "staged":
		validatedStagedStage, err := s.validateStagedStage(stageIdx, defaults)
		if err != nil {
			return nil, fmt.Errorf("validating staged stage: %w", err)
		}
		rates, err := staged.CalculateStagedRate(
			*validatedStagedStage.Jitter,
			*validatedStagedStage.IterationFrequency,
			*validatedStagedStage.Stages,
			*validatedStagedStage.Distribution,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("calculating staged rate: %w", err)
		}

		return &runnableStage{
			StageDuration:     *validatedStagedStage.Duration,
			IterationDuration: rates.IterationDuration,
			Rate:              rates.Rate,
			Params:            *validatedStagedStage.Parameters,
		}, nil
	case "gaussian":
		validatedGaussianStage, err := s.validateGaussianStage(stageIdx, defaults)
		if err != nil {
			return nil, fmt.Errorf("validating gaussian stage: %w", err)
		}
		rates, err := gaussian.CalculateGaussianRate(
			*validatedGaussianStage.Volume, *validatedGaussianStage.Jitter, *validatedGaussianStage.Repeat,
			*validatedGaussianStage.IterationFrequency, *validatedGaussianStage.Peak, *validatedGaussianStage.StandardDeviation,
			*validatedGaussianStage.Weights, *validatedGaussianStage.Distribution,
		)
		if err != nil {
			return nil, fmt.Errorf("calculating gaussian rate: %w", err)
		}

		return &runnableStage{
			StageDuration:     *validatedGaussianStage.Duration,
			IterationDuration: rates.IterationDuration,
			Rate:              rates.Rate,
			Params:            *validatedGaussianStage.Parameters,
		}, nil
	case "users":
		validatedUsersStage, err := s.validateUsersStage(stageIdx, defaults)
		if err != nil {
			return nil, err
		}
		return &runnableStage{
			StageDuration:    *validatedUsersStage.Duration,
			Params:           *validatedUsersStage.Parameters,
			UsersConcurrency: *validatedUsersStage.Concurrency,
		}, nil
	default:
		return nil, fmt.Errorf("invalid stage mode at stage %d", stageIdx)
	}
}

func (c *ConfigFile) validateCommonFields() (*ConfigFile, error) {
	if c.Scenario == nil {
		return nil, errors.New("missing scenario")
	}
	if c.Limits.MaxDuration == nil {
		return nil, errors.New("missing max-duration")
	}
	if c.Limits.Concurrency == nil {
		return nil, errors.New("missing concurrency")
	}
	if c.Limits.MaxIterations == nil {
		return nil, errors.New("missing max-iterations")
	}
	if c.Limits.IgnoreDropped == nil {
		return nil, errors.New("missing ignore-dropped")
	}
	if len(c.Stages) == 0 {
		return nil, errors.New("missing stages")
	}

	if c.Limits.MaxFailures == nil {
		maxFailures := uint64(0)
		c.Limits.MaxFailures = &maxFailures
	}
	if c.Limits.MaxFailuresRate == nil {
		maxFailuresRate := 0
		c.Limits.MaxFailuresRate = &maxFailuresRate
	}
	if c.Default.Concurrency == nil {
		c.Default.Concurrency = c.Limits.Concurrency
	}

	return c, nil
}

func (s *Stage) validateCommonFieldsOfStage(idx int, defaults Stage) (*Stage, error) {
	if s.Duration == nil {
		if defaults.Duration == nil {
			return nil, fmt.Errorf("missing duration at stage %d", idx)
		}
		s.Duration = defaults.Duration
	}
	if s.Mode == nil {
		if defaults.Mode == nil {
			return nil, fmt.Errorf("missing stage mode at stage %d", idx)
		}
		s.Mode = defaults.Mode
	}

	return s, nil
}

func (s *Stage) validateConstantStage(idx int, defaults Stage) (*Stage, error) {
	if s.Rate == nil {
		if defaults.Rate == nil {
			return nil, fmt.Errorf("missing rate at stage %d", idx)
		}
		s.Rate = defaults.Rate
	}
	if s.Distribution == nil {
		if defaults.Distribution == nil {
			return nil, fmt.Errorf("missing distribution at stage %d", idx)
		}
		s.Distribution = defaults.Distribution
	}
	if s.Jitter == nil {
		s.Jitter = defaults.Jitter
	}
	if s.Parameters == nil {
		if defaults.Parameters == nil {
			s.Parameters = &map[string]string{}
		} else {
			s.Parameters = defaults.Parameters
		}
	}

	return s, nil
}

func (s *Stage) validateRampStage(idx int, defaults Stage) (*Stage, error) {
	if s.StartRate == nil {
		if defaults.StartRate == nil {
			return nil, fmt.Errorf("missing start-rate at stage %d", idx)
		}
		s.StartRate = defaults.StartRate
	}
	if s.EndRate == nil {
		if defaults.EndRate == nil {
			return nil, fmt.Errorf("missing end-rate at stage %d", idx)
		}
		s.EndRate = defaults.EndRate
	}
	if s.Distribution == nil {
		if defaults.Distribution == nil {
			return nil, fmt.Errorf("missing distribution at stage %d", idx)
		}
		s.Distribution = defaults.Distribution
	}
	if s.Jitter == nil {
		s.Jitter = defaults.Jitter
	}
	if s.Parameters == nil {
		if defaults.Parameters == nil {
			s.Parameters = &map[string]string{}
		} else {
			s.Parameters = defaults.Parameters
		}
	}

	return s, nil
}

func (s *Stage) validateStagedStage(idx int, defaults Stage) (*Stage, error) {
	if s.Stages == nil {
		if defaults.Stages == nil {
			return nil, fmt.Errorf("missing stages at stage %d", idx)
		}
		s.Stages = defaults.Stages
	}
	if s.IterationFrequency == nil {
		if defaults.IterationFrequency == nil {
			return nil, fmt.Errorf("missing iteration-frequency at stage %d", idx)
		}
		s.IterationFrequency = defaults.IterationFrequency
	}
	if s.Distribution == nil {
		if defaults.Distribution == nil {
			return nil, fmt.Errorf("missing distribution at stage %d", idx)
		}
		s.Distribution = defaults.Distribution
	}
	if s.Jitter == nil {
		s.Jitter = defaults.Jitter
	}
	if s.Parameters == nil {
		if defaults.Parameters == nil {
			s.Parameters = &map[string]string{}
		} else {
			s.Parameters = defaults.Parameters
		}
	}

	return s, nil
}

func (s *Stage) validateGaussianStage(idx int, defaults Stage) (*Stage, error) {
	if s.Volume == nil {
		if defaults.Volume == nil {
			return nil, fmt.Errorf("missing volume at stage %d", idx)
		}
		s.Volume = defaults.Volume
	}
	if s.Repeat == nil {
		if defaults.Repeat == nil {
			return nil, fmt.Errorf("missing repeat at stage %d", idx)
		}
		s.Repeat = defaults.Repeat
	}
	if s.IterationFrequency == nil {
		if defaults.IterationFrequency == nil {
			return nil, fmt.Errorf("missing iteration-frequency at stage %d", idx)
		}
		s.IterationFrequency = defaults.IterationFrequency
	}
	if s.Peak == nil {
		if defaults.Peak == nil {
			return nil, fmt.Errorf("missing peak at stage %d", idx)
		}
		s.Peak = defaults.Peak
	}
	if s.Weights == nil {
		if defaults.Weights == nil {
			return nil, fmt.Errorf("missing weights at stage %d", idx)
		}
		s.Weights = defaults.Weights
	}
	if s.StandardDeviation == nil {
		if defaults.StandardDeviation == nil {
			return nil, fmt.Errorf("missing standard-deviation at stage %d", idx)
		}
		s.StandardDeviation = defaults.StandardDeviation
	}
	if s.Distribution == nil {
		if defaults.Distribution == nil {
			return nil, fmt.Errorf("missing distribution at stage %d", idx)
		}
		s.Distribution = defaults.Distribution
	}
	if s.Jitter == nil {
		s.Jitter = defaults.Jitter
	}
	if s.Parameters == nil {
		if defaults.Parameters == nil {
			s.Parameters = &map[string]string{}
		} else {
			s.Parameters = defaults.Parameters
		}
	}

	return s, nil
}

func (s *Stage) validateUsersStage(idx int, defaults Stage) (*Stage, error) {
	if s.Concurrency == nil {
		if defaults.Concurrency == nil {
			return nil, fmt.Errorf("missing users at stage %d", idx)
		}

		s.Concurrency = defaults.Concurrency
	}
	if s.Parameters == nil {
		if defaults.Parameters == nil {
			s.Parameters = &map[string]string{}
		} else {
			s.Parameters = defaults.Parameters
		}
	}

	return s, nil
}
