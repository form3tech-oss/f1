package scenarios_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
	"github.com/form3tech-oss/f1/v3/pkg/f1/scenarios"
)

func TestWithDescription(t *testing.T) {
	t.Parallel()

	info := &scenarios.Scenario{Name: "test"}
	scenarios.WithDescription("a load test")(info)
	require.Equal(t, "a load test", info.Description)
}

func TestWithParameter(t *testing.T) {
	t.Parallel()

	info := &scenarios.Scenario{Name: "test"}
	param := scenarios.ScenarioParameter{Name: "rate", Description: "requests per second", Default: "1/s"}
	scenarios.WithParameter(param)(info)
	require.Len(t, info.Parameters, 1)
	require.Equal(t, "rate", info.Parameters[0].Name)
	require.Equal(t, "requests per second", info.Parameters[0].Description)
	require.Equal(t, "1/s", info.Parameters[0].Default)

	scenarios.WithParameter(scenarios.ScenarioParameter{Name: "duration", Default: "10s"})(info)
	require.Len(t, info.Parameters, 2)
	require.Equal(t, "duration", info.Parameters[1].Name)
}

func TestAddScenarioAndGetScenario(t *testing.T) {
	t.Parallel()

	s := scenarios.New()
	scenario := &scenarios.Scenario{
		Name:       "myScenario",
		ScenarioFn: func(context.Context, *f1testing.T) f1testing.RunFn { return func(context.Context, *f1testing.T) {} },
	}

	s.AddScenario(scenario)

	got := s.GetScenario("myScenario")
	require.NotNil(t, got)
	require.Equal(t, "myScenario", got.Name)
	require.Nil(t, s.GetScenario("nonexistent"))
}

func TestGetScenarioNames(t *testing.T) {
	t.Parallel()

	s := scenarios.New()
	mkScenario := func(name string) *scenarios.Scenario {
		return &scenarios.Scenario{
			Name:       name,
			ScenarioFn: func(context.Context, *f1testing.T) f1testing.RunFn { return func(context.Context, *f1testing.T) {} },
		}
	}

	s.AddScenario(mkScenario("zebra")).
		AddScenario(mkScenario("alpha")).
		AddScenario(mkScenario("beta"))

	names := s.GetScenarioNames()
	require.Equal(t, []string{"alpha", "beta", "zebra"}, names)
}

func TestGetScenarioNamesEmpty(t *testing.T) {
	t.Parallel()

	s := scenarios.New()
	names := s.GetScenarioNames()
	require.Empty(t, names)
}
