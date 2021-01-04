package file

import (
	"fmt"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1/trigger/constant"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/gaussian"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/staged"
	"gopkg.in/yaml.v2"
)

type ConfigFile struct {
	Default  Stage `yaml:"default"`
	Limits   Limits
	Schedule Schedule `yaml:"schedule"`
	Stages   []Stage  `yaml:"stages"`
}

type Schedule struct {
	StageStart *time.Time `yaml:"stage_start"`
}

type Limits struct {
	MaxDuration   time.Duration `yaml:"max-duration"`
	Concurrency   int           `yaml:"concurrency"`
	MaxIterations int32         `yaml:"max-iterations"`
}

type Stage struct {
	Mode               *string            `yaml:"mode"`
	StartRate          *string            `yaml:"start_rate"`
	EndRate            *string            `yaml:"end_rate"`
	Rate               *string            `yaml:"rate"`
	Distribution       *string            `yaml:"distribution"`
	Weights            *string            `yaml:"weights"`
	Users              *int               `yaml:"users"`
	Jitter             *float64           `yaml:"jitter"`
	Volume             *float64           `yaml:"volume"`
	Duration           *time.Duration     `yaml:"duration"`
	IterationFrequency *time.Duration     `yaml:"iteration-frequency"`
	Repeat             *time.Duration     `yaml:"repeat"`
	Peak               *time.Duration     `yaml:"peak"`
	StandardDeviation  *time.Duration     `yaml:"standard-deviation"`
	Parameters         *map[string]string `yaml:"parameters"`
}

func parseConfigFile(fileContent []byte, now time.Time) (*runnableStages, error) {
	configFile := ConfigFile{}
	err := yaml.Unmarshal(fileContent, &configFile)
	if err != nil {
		return nil, err
	}
	err = configFile.validateCommonFields()
	if err != nil {
		return nil, err
	}

	var stages []runnableStage
	stagesTotalDuration := 0 * time.Second
	for i, stageConfig := range configFile.Stages {
		if stageConfig.Duration == nil {
			if configFile.Default.Duration == nil {
				return nil, fmt.Errorf("missing duration at stage %d", i)
			} else {
				stageConfig.Duration = configFile.Default.Duration
			}
		}
		if stageConfig.Mode == nil {
			if configFile.Default.Mode == nil {
				return nil, fmt.Errorf("missing stage mode at stage %d", i)
			} else {
				stageConfig.Mode = configFile.Default.Mode
			}
		}
		stagesTotalDuration += *stageConfig.Duration

		if configFile.Schedule.StageStart == nil || configFile.Schedule.StageStart.Add(stagesTotalDuration).After(now) {

			var stage runnableStage

			switch *stageConfig.Mode {
			case "constant":
				validatedStage, err := stageConfig.validateConstantStage(i, configFile.Default)
				if err != nil {
					return nil, err
				}
				rates, err := constant.CalculateConstantRate(*validatedStage.Jitter, *validatedStage.Rate, *validatedStage.Distribution)
				if err != nil {
					return nil, err
				}

				stage = runnableStage{
					stageDuration:     *validatedStage.Duration,
					iterationDuration: rates.DistributedIterationDuration,
					rate:              rates.DistributedRate,
					params:            *validatedStage.Parameters,
				}
			case "stage":
				validatedStage, err := stageConfig.validateStagedStage(i, configFile.Default)
				if err != nil {
					return nil, err
				}
				stg := fmt.Sprintf("0s:%s, %s:%s", *validatedStage.StartRate, *validatedStage.Duration, *validatedStage.EndRate)
				rates, err := staged.CalculateStagedRate(*validatedStage.Jitter, *validatedStage.IterationFrequency, stg, *validatedStage.Distribution)
				if err != nil {
					return nil, err
				}

				stage = runnableStage{
					stageDuration:     *stageConfig.Duration,
					iterationDuration: rates.DistributedIterationDuration,
					rate:              rates.DistributedRate,
					params:            *stageConfig.Parameters,
				}
			case "gaussian":
				validatedStage, err := stageConfig.validateGaussianStage(i, configFile.Default)
				if err != nil {
					return nil, err
				}
				rates, err := gaussian.CalculateGaussianRate(
					*validatedStage.Volume, *validatedStage.Jitter, *validatedStage.Repeat, *validatedStage.IterationFrequency, *validatedStage.Peak,
					*validatedStage.StandardDeviation, *validatedStage.Weights, *validatedStage.Distribution,
				)
				if err != nil {
					return nil, err
				}

				stage = runnableStage{
					stageDuration:     *stageConfig.Duration,
					iterationDuration: rates.DistributedIterationDuration,
					rate:              rates.DistributedRate,
					params:            *stageConfig.Parameters,
				}
			case "users":
				validatedStage, err := stageConfig.validateUsersStage(i, configFile.Default)
				if err != nil {
					return nil, err
				}
				stage = runnableStage{
					stageDuration: *validatedStage.Duration,
					params:        *validatedStage.Parameters,
					users:         *validatedStage.Users,
				}
			default:
				return nil, fmt.Errorf("missing stage mode at stage %d", i)
			}

			stages = append(stages, stage)
		}
	}

	return &runnableStages{
		stages:              stages,
		stagesTotalDuration: stagesTotalDuration,
		maxDuration:         configFile.Limits.MaxDuration,
		concurrency:         configFile.Limits.Concurrency,
		maxIterations:       configFile.Limits.MaxIterations,
	}, nil
}

func (r *ConfigFile) validateCommonFields() error {
	if r.Limits.MaxDuration == 0*time.Second {
		return fmt.Errorf("missing max-duration")
	} else if r.Limits.Concurrency == 0 {
		return fmt.Errorf("missing concurrency")
	} else if r.Limits.MaxIterations == 0 {
		return fmt.Errorf("missing max-iterations")
	}

	return nil
}

func (r *Stage) validateConstantStage(idx int, defaults Stage) (*Stage, error) {
	if r.Rate == nil {
		if defaults.Rate == nil {
			return nil, fmt.Errorf("missing rate at stage %d", idx)
		} else {
			r.Rate = defaults.Rate
		}
	}
	if r.Distribution == nil {
		if defaults.Distribution == nil {
			return nil, fmt.Errorf("missing distribution at stage %d", idx)
		} else {
			r.Distribution = defaults.Distribution
		}
	}
	if r.Jitter == nil {
		r.Jitter = defaults.Jitter
	}
	if r.Parameters == nil {
		if defaults.Parameters == nil {
			r.Parameters = &map[string]string{}
		} else {
			r.Parameters = defaults.Parameters
		}
	}

	return r, nil
}

func (r *Stage) validateStagedStage(idx int, defaults Stage) (*Stage, error) {
	if r.StartRate == nil {
		if defaults.StartRate == nil {
			return nil, fmt.Errorf("missing start rate at stage %d", idx)
		} else {
			r.StartRate = defaults.StartRate
		}
	}
	if r.EndRate == nil {
		if defaults.EndRate == nil {
			return nil, fmt.Errorf("missing end rate at stage %d", idx)
		} else {
			r.EndRate = defaults.EndRate
		}
	}
	if r.IterationFrequency == nil {
		if defaults.IterationFrequency == nil {
			return nil, fmt.Errorf("missing iteration-frequency at stage %d", idx)
		} else {
			r.IterationFrequency = defaults.IterationFrequency
		}
	}
	if r.Distribution == nil {
		if defaults.Distribution == nil {
			return nil, fmt.Errorf("missing distribution at stage %d", idx)
		} else {
			r.Distribution = defaults.Distribution
		}
	}
	if r.Jitter == nil {
		r.Jitter = defaults.Jitter
	}
	if r.Parameters == nil {
		if defaults.Parameters == nil {
			r.Parameters = &map[string]string{}
		} else {
			r.Parameters = defaults.Parameters
		}
	}

	return r, nil
}

func (r *Stage) validateGaussianStage(idx int, defaults Stage) (*Stage, error) {
	if r.Volume == nil {
		if defaults.Volume == nil {
			return nil, fmt.Errorf("missing volume at stage %d", idx)
		} else {
			r.Volume = defaults.Volume
		}
	}
	if r.Repeat == nil {
		if defaults.Repeat == nil {
			return nil, fmt.Errorf("missing repeat at stage %d", idx)
		} else {
			r.Repeat = defaults.Repeat
		}
	}
	if r.IterationFrequency == nil {
		if defaults.IterationFrequency == nil {
			return nil, fmt.Errorf("missing iteration-frequency at stage %d", idx)
		} else {
			r.IterationFrequency = defaults.IterationFrequency
		}
	}
	if r.Peak == nil {
		if defaults.Peak == nil {
			return nil, fmt.Errorf("missing peak at stage %d", idx)
		} else {
			r.Peak = defaults.Peak
		}
	}
	if r.Weights == nil {
		if defaults.Weights == nil {
			return nil, fmt.Errorf("missing weights at stage %d", idx)
		} else {
			r.Weights = defaults.Weights
		}
	}
	if r.StandardDeviation == nil {
		if defaults.StandardDeviation == nil {
			return nil, fmt.Errorf("missing standard-deviation at stage %d", idx)
		} else {
			r.StandardDeviation = defaults.StandardDeviation
		}
	}
	if r.Distribution == nil {
		if defaults.Distribution == nil {
			return nil, fmt.Errorf("missing distribution at stage %d", idx)
		} else {
			r.Distribution = defaults.Distribution
		}
	}
	if r.Jitter == nil {
		r.Jitter = defaults.Jitter
	}
	if r.Parameters == nil {
		if defaults.Parameters == nil {
			r.Parameters = &map[string]string{}
		} else {
			r.Parameters = defaults.Parameters
		}
	}

	return r, nil
}

func (r *Stage) validateUsersStage(idx int, defaults Stage) (*Stage, error) {
	if r.Users == nil {
		if defaults.Users == nil {
			return nil, fmt.Errorf("missing users at stage %d", idx)
		} else {
			r.Users = defaults.Users
		}
	}
	if r.Parameters == nil {
		if defaults.Parameters == nil {
			r.Parameters = &map[string]string{}
		} else {
			r.Parameters = defaults.Parameters
		}
	}

	return r, nil
}
