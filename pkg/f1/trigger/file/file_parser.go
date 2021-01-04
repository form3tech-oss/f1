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
	Mode               string            `yaml:"mode"`
	StartRate          string            `yaml:"start_rate"`
	EndRate            string            `yaml:"end_rate"`
	Rate               string            `yaml:"rate"`
	Distribution       string            `yaml:"distribution"`
	Weights            string            `yaml:"weights"`
	Users              int               `yaml:"users"`
	Jitter             float64           `yaml:"jitter"`
	Volume             float64           `yaml:"volume"`
	Duration           time.Duration     `yaml:"duration"`
	IterationFrequency time.Duration     `yaml:"iteration-frequency"`
	Repeat             time.Duration     `yaml:"repeat"`
	Peak               time.Duration     `yaml:"peak"`
	StandardDeviation  time.Duration     `yaml:"standard-deviation"`
	Parameters         map[string]string `yaml:"parameters"`
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
		stagesTotalDuration += stageConfig.Duration

		if configFile.Schedule.StageStart == nil || configFile.Schedule.StageStart.Add(stagesTotalDuration).After(now) {
			var stage runnableStage

			switch stageConfig.Mode {
			case "constant":
				err := stageConfig.validateConstantStage(i)
				if err != nil {
					return nil, err
				}
				rates, err := constant.CalculateConstantRate(stageConfig.Jitter, stageConfig.Rate, stageConfig.Distribution)
				if err != nil {
					return nil, err
				}

				stage = runnableStage{
					stageDuration:     stageConfig.Duration,
					iterationDuration: rates.DistributedIterationDuration,
					rate:              rates.DistributedRate,
					params:            stageConfig.Parameters,
				}
			case "stage":
				err := stageConfig.validateStagedStage(i)
				if err != nil {
					return nil, err
				}
				stg := fmt.Sprintf("0s:%s, %s:%s", stageConfig.StartRate, stageConfig.Duration, stageConfig.EndRate)
				rates, err := staged.CalculateStagedRate(stageConfig.Jitter, stageConfig.IterationFrequency, stg, stageConfig.Distribution)
				if err != nil {
					return nil, err
				}

				stage = runnableStage{
					stageDuration:     stageConfig.Duration,
					iterationDuration: rates.DistributedIterationDuration,
					rate:              rates.DistributedRate,
					params:            stageConfig.Parameters,
				}
			case "gaussian":
				err := stageConfig.validateGaussianStage(i)
				if err != nil {
					return nil, err
				}
				rates, err := gaussian.CalculateGaussianRate(
					stageConfig.Volume, stageConfig.Jitter, stageConfig.Repeat, stageConfig.IterationFrequency, stageConfig.Peak,
					stageConfig.StandardDeviation, stageConfig.Weights, stageConfig.Distribution,
				)
				if err != nil {
					return nil, err
				}

				stage = runnableStage{
					stageDuration:     stageConfig.Duration,
					iterationDuration: rates.DistributedIterationDuration,
					rate:              rates.DistributedRate,
					params:            stageConfig.Parameters,
				}
			case "users":
				err := stageConfig.validateUsersStage(i)
				if err != nil {
					return nil, err
				}
				stage = runnableStage{
					stageDuration: stageConfig.Duration,
					params:        stageConfig.Parameters,
					users:         stageConfig.Users,
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

func (r *Stage) validateConstantStage(idx int) error {
	if r.Rate == "" {
		return fmt.Errorf("missing rate at stage %d", idx)
	} else if r.Distribution == "" {
		return fmt.Errorf("missing distribution at stage %d", idx)
	}

	return nil
}

func (r *Stage) validateStagedStage(idx int) error {
	if r.StartRate == "" {
		return fmt.Errorf("missing start rate at stage %d", idx)
	} else if r.EndRate == "" {
		return fmt.Errorf("missing end rate at stage %d", idx)
	} else if r.IterationFrequency == 0*time.Second {
		return fmt.Errorf("missing iteration-frequency at stage %d", idx)
	} else if r.Distribution == "" {
		return fmt.Errorf("missing distribution at stage %d", idx)
	}

	return nil
}

func (r *Stage) validateGaussianStage(idx int) error {
	if r.Volume == 0 {
		return fmt.Errorf("missing volume at stage %d", idx)
	} else if r.Repeat == 0 {
		return fmt.Errorf("missing repeat at stage %d", idx)
	} else if r.IterationFrequency == 0 {
		return fmt.Errorf("missing iteration-frequency at stage %d", idx)
	} else if r.Peak == 0 {
		return fmt.Errorf("missing peak at stage %d", idx)
	} else if r.Weights == "" {
		return fmt.Errorf("missing weights at stage %d", idx)
	} else if r.StandardDeviation == 0 {
		return fmt.Errorf("missing standard-deviation at stage %d", idx)
	} else if r.Distribution == "" {
		return fmt.Errorf("missing distribution at stage %d", idx)
	}

	return nil
}

func (r *Stage) validateUsersStage(idx int) error {
	if r.Users == 0 {
		return fmt.Errorf("missing users at stage %d", idx)
	}

	return nil
}
