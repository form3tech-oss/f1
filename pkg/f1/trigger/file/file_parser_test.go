package file

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFileRate_SingleStages(t *testing.T) {
	for _, test := range []testData{
		{
			testName: "Constant mode",
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 5s
  mode: constant
  rate: 6/s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     5 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{6, 6, 6, 6, 6, 6},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Ramp mode",
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: ramp
  start-rate: 0/s
  end-rate: 10/s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Staged mode",
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: staged
  stages: 0s:0,10s:10
  iteration-frequency: 1s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Gaussian mode",
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: gaussian
  volume: 100
  repeat: 20s
  iteration-frequency: 1s
  peak: 10s
  weights: "1.0,1.0"
  standard-deviation: 3s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates: []int{
				0, 0, 1, 2, 3, 6, 8, 10, 13, 13, 13, 10, 9, 5, 3, 2, 1, 0, 1, 0,
				0, 0, 1, 2, 3, 6, 8, 10, 13, 13, 13, 11, 8, 5, 3, 2, 1, 0, 1, 0,
				0, 0, 1, 2, 3, 6, 8, 10, 13, 13, 13, 11, 8, 5, 3, 2, 1, 1, 0, 0,
			},
			expectedParameters: map[string]string{"SOP": "1"},
		},
		{
			testName: "Users mode",
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: users
  concurrency: 100
  parameters:
    SOP: 1
`,
			expectedScenario:         "template",
			expectedMaxDuration:      1 * time.Minute,
			expectedConcurrency:      50,
			expectedMaxIterations:    100,
			expectedIgnoreDropped:    true,
			expectedTotalDuration:    10 * time.Second,
			expectedUsersConcurrency: 100,
			expectedParameters:       map[string]string{"SOP": "1"},
		},
		{
			testName: "Constant mode using default values",
			fileContent: `
scenario: template
default:
  mode: constant
  rate: 6/s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 5s
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     5 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{6, 6, 6, 6, 6, 6},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Ramp mode using default values",
			fileContent: `
scenario: template
default:
  mode: ramp
  start-rate: 0
  end-rate: 10
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Staged mode using default values",
			fileContent: `
scenario: template
default:
  mode: staged
  stages: 0s:0,10s:10
  iteration-frequency: 1s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Gaussian mode using default values",
			fileContent: `
scenario: template
default:
  mode: gaussian
  volume: 100
  repeat: 20s
  iteration-frequency: 1s
  peak: 10s
  weights: "1.0,1.0"
  standard-deviation: 3s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates: []int{
				0, 0, 1, 2, 3, 6, 8, 10, 13, 13, 13, 10, 9, 5, 3, 2, 1, 0, 1, 0,
				0, 0, 1, 2, 3, 6, 8, 10, 13, 13, 13, 11, 8, 5, 3, 2, 1, 0, 1, 0,
				0, 0, 1, 2, 3, 6, 8, 10, 13, 13, 13, 11, 8, 5, 3, 2, 1, 1, 0, 0,
			},
			expectedParameters: map[string]string{"SOP": "1"},
		},
		{
			testName: "Users mode using default values",
			fileContent: `
scenario: template
default:
  duration: 10s
  mode: users
  parameters:
    SOP: 1
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- mode: users
`,
			expectedScenario:         "template",
			expectedMaxDuration:      1 * time.Minute,
			expectedConcurrency:      50,
			expectedMaxIterations:    100,
			expectedIgnoreDropped:    true,
			expectedTotalDuration:    10 * time.Second,
			expectedUsersConcurrency: 50,
			expectedParameters:       map[string]string{"SOP": "1"},
		},
		{
			testName: "Skip completed stages when stage-start is provided",
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
schedule:
  stage-start: "2020-12-10T09:00:00+00:00"
stages:
- duration: 1h
  mode: constant
  rate: 1/s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
- duration: 5s
  mode: constant
  rate: 2/s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
`,
			expectedScenario:          "template",
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedIgnoreDropped:     true,
			expectedTotalDuration:     5 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{2, 2, 2, 2, 2},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, "2020-12-10T10:00:00+00:00")

			stagesToRun, err := parseConfigFile([]byte(test.fileContent), now)

			require.NoError(t, err)
			require.Equal(t, 1, len(stagesToRun.stages))
			require.Equal(t, test.expectedScenario, stagesToRun.scenario)
			require.Equal(t, test.expectedMaxDuration, stagesToRun.maxDuration)
			require.Equal(t, test.expectedConcurrency, stagesToRun.concurrency)
			require.Equal(t, test.expectedMaxIterations, stagesToRun.maxIterations)
			require.Equal(t, test.expectedIgnoreDropped, stagesToRun.ignoreDropped)
			require.Equal(t, test.expectedTotalDuration, stagesToRun.stages[0].stageDuration)
			require.Equal(t, test.expectedIterationDuration, stagesToRun.stages[0].iterationDuration)
			require.Equal(t, test.expectedParameters, stagesToRun.stages[0].params)
			require.Equal(t, test.expectedUsersConcurrency, stagesToRun.stages[0].usersConcurrency)

			if len(test.expectedRates) > 0 {
				var rates []int
				for range test.expectedRates {
					now = now.Add(test.expectedIterationDuration)
					rates = append(rates, stagesToRun.stages[0].rate(now))
				}
				require.Equal(t, test.expectedRates, rates)
			}
		})
	}
}

func TestFileRate_FileErrors(t *testing.T) {
	for _, test := range []struct {
		fileContent, expectedError string
	}{
		{
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
`,
			expectedError: "missing scenario",
		},
		{
			fileContent: `
scenario: template
limits:
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
`,
			expectedError: "missing max-duration",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  max-iterations: 100
  ignore-dropped: true
`,
			expectedError: "missing concurrency",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  ignore-dropped: true
`,
			expectedError: "missing max-iterations",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
`,
			expectedError: "missing ignore-dropped",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
`,
			expectedError: "missing stages",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 1h
  mode: constant
`,
			expectedError: "missing rate at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 1h
  mode: constant
  rate: 6/s
`,
			expectedError: "missing distribution at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: ramp
`,
			expectedError: "missing start-rate at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: ramp
  start-rate: 0
`,
			expectedError: "missing end-rate at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: ramp
  start-rate: 0
  end-rate: 10
`,
			expectedError: "missing distribution at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: staged
  iteration-frequency: 1s
`,
			expectedError: "missing stages at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: staged
  stages: 0s:0,10s:10
`,
			expectedError: "missing iteration-frequency at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: gaussian
`,
			expectedError: "missing volume at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: gaussian
  volume: 100
`,
			expectedError: "missing repeat at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
  mode: gaussian
  volume: 100
  repeat: 20s
`,
			expectedError: "missing iteration-frequency at stage 0",
		},
		{
			fileContent: `
scenario: template
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
  ignore-dropped: true
stages:
- duration: 10s
`,
			expectedError: "missing stage mode at stage 0",
		},
		{
			fileContent: `
invalid file content
`,
			expectedError: "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `invalid...` into file.ConfigFile",
		},
	} {
		t.Run(test.expectedError, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, "2020-12-10T10:00:00+00:00")

			runnableStages, err := parseConfigFile([]byte(test.fileContent), now)

			require.Nil(t, runnableStages)
			require.EqualError(t, err, test.expectedError)
		})
	}
}

type testData struct {
	testName                  string
	fileContent               string
	expectedScenario          string
	expectedTotalDuration     time.Duration
	expectedIterationDuration time.Duration
	expectedMaxDuration       time.Duration
	expectedIgnoreDropped     bool
	expectedMaxIterations     int32
	expectedConcurrency       int
	expectedUsersConcurrency  int
	expectedRates             []int
	expectedParameters        map[string]string
}
