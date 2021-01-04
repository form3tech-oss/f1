package file

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFileRate_SimpleStages(t *testing.T) {
	for _, test := range []testData{
		{
			testName: "Constant mode single stage",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
stages:
- duration: 5s
  mode: constant
  rate: 6/s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
`,
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedTotalDuration:     5 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{6, 6, 6, 6, 6, 6},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Staged mode single stage",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
stages:
- duration: 10s
  mode: stage
  start_rate: 0
  end_rate: 10
  iteration-frequency: 1s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
`,
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Gaussian mode single stage",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
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
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
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
			testName: "Users mode single stage",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
stages:
- duration: 10s
  mode: users
  users: 100
  parameters:
    SOP: 1
`,
			expectedMaxDuration:   1 * time.Minute,
			expectedConcurrency:   50,
			expectedMaxIterations: 100,
			expectedTotalDuration: 10 * time.Second,
			expectedUsers:         100,
			expectedParameters:    map[string]string{"SOP": "1"},
		},
		{
			testName: "Start with the stage corresponding to a given time",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T09:00:00+00:00"
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
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedTotalDuration:     5 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{2, 2, 2, 2, 2},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Constant mode single stage using default values",
			fileContent: `
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
stages:
- duration: 5s
`,
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedTotalDuration:     5 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{6, 6, 6, 6, 6, 6},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Staged mode single stage using default values",
			fileContent: `
default:
  mode: stage
  start_rate: 0
  end_rate: 10
  iteration-frequency: 1s
  jitter: 0
  distribution: none
  parameters:
    SOP: 1
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
stages:
- duration: 10s
`,
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
			expectedTotalDuration:     10 * time.Second,
			expectedIterationDuration: 1 * time.Second,
			expectedRates:             []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expectedParameters:        map[string]string{"SOP": "1"},
		},
		{
			testName: "Gaussian mode single stage using default values",
			fileContent: `
default:
  mode: gaussian
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
stages:
- duration: 10s
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
			expectedMaxDuration:       1 * time.Minute,
			expectedConcurrency:       50,
			expectedMaxIterations:     100,
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
			testName: "Users mode single stage using default values",
			fileContent: `
default:
  duration: 10s
  mode: users
  users: 100
  parameters:
    SOP: 1
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
stages:
- mode: users
`,
			expectedMaxDuration:   1 * time.Minute,
			expectedConcurrency:   50,
			expectedMaxIterations: 100,
			expectedTotalDuration: 10 * time.Second,
			expectedUsers:         100,
			expectedParameters:    map[string]string{"SOP": "1"},
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, "2020-12-10T10:00:00+00:00")

			stagesToRun, err := parseConfigFile([]byte(test.fileContent), now)

			require.NoError(t, err)
			require.Equal(t, 1, len(stagesToRun.stages))
			require.Equal(t, test.expectedMaxDuration, stagesToRun.maxDuration)
			require.Equal(t, test.expectedConcurrency, stagesToRun.concurrency)
			require.Equal(t, test.expectedMaxIterations, stagesToRun.maxIterations)
			require.Equal(t, test.expectedTotalDuration, stagesToRun.stages[0].stageDuration)
			require.Equal(t, test.expectedIterationDuration, stagesToRun.stages[0].iterationDuration)
			require.Equal(t, test.expectedParameters, stagesToRun.stages[0].params)
			require.Equal(t, test.expectedUsers, stagesToRun.stages[0].users)

			if len(test.expectedRates) > 0 {
				var rates []int
				for _, _ = range test.expectedRates {
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
		testName      string
		fileContent   string
		expectedError string
	}{
		{
			testName: "missing max-duration",
			fileContent: `
limits:
  concurrency: 50
  max-iterations: 100
`,
			expectedError: "missing max-duration",
		},
		{
			testName: "missing concurrency",
			fileContent: `
limits:
  max-duration: 1m
  max-iterations: 100
`,
			expectedError: "missing concurrency",
		},
		{
			testName: "missing max-iterations",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
`,
			expectedError: "missing max-iterations",
		},
		{
			testName: "missing constant rate",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 1h
  mode: constant
`,
			expectedError: "missing rate at stage 0",
		},
		{
			testName: "missing constant distribution",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 1h
  mode: constant
  rate: 6/s
`,
			expectedError: "missing distribution at stage 0",
		},
		{
			testName: "missing staged start rate",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: stage
`,
			expectedError: "missing start rate at stage 0",
		},
		{
			testName: "missing staged end rate",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: stage
  start_rate: 0
`,
			expectedError: "missing end rate at stage 0",
		},
		{
			testName: "missing staged iteration-frequency",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: stage
  start_rate: 0
  end_rate: 10
`,
			expectedError: "missing iteration-frequency at stage 0",
		},
		{
			testName: "missing staged distribution",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: stage
  start_rate: 0
  end_rate: 10
  iteration-frequency: 1s
`,
			expectedError: "missing distribution at stage 0",
		},
		{
			testName: "missing users stage users value",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: users
`,
			expectedError: "missing users at stage 0",
		},
		{
			testName: "missing gaussian volume",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: gaussian
`,
			expectedError: "missing volume at stage 0",
		},
		{
			testName: "missing gaussian repeat",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: gaussian
  volume: 100
`,
			expectedError: "missing repeat at stage 0",
		},
		{
			testName: "missing gaussian iteration-frequency",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
  mode: gaussian
  volume: 100
  repeat: 20s
`,
			expectedError: "missing iteration-frequency at stage 0",
		},
		{
			testName: "missing stage mode",
			fileContent: `
limits:
  max-duration: 1m
  concurrency: 50
  max-iterations: 100
schedule:
  stage_start: "2020-12-10T10:00:00+00:00"
stages:
- duration: 10s
`,
			expectedError: "missing stage mode at stage 0",
		},
		{
			testName: "invalid file content",
			fileContent: `
invalid file content
`,
			expectedError: "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `invalid...` into file.ConfigFile",
		},
	} {
		t.Run(test.testName, func(t *testing.T) {
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
	expectedTotalDuration     time.Duration
	expectedIterationDuration time.Duration
	expectedRates             []int
	expectedMaxDuration       time.Duration
	expectedConcurrency       int
	expectedMaxIterations     int32
	expectedParameters        map[string]string
	expectedUsers             int
}
