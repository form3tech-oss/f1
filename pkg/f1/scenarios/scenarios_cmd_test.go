package scenarios_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/form3tech-oss/f1/v3/pkg/f1/f1testing"
	"github.com/form3tech-oss/f1/v3/pkg/f1/scenarios"
)

func TestScenariosLsOutput(t *testing.T) {
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

	cmd := scenarios.Cmd(s)
	cmd.SetArgs([]string{"ls"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	require.Equal(t, []string{"alpha", "beta", "zebra"}, lines, "output should be newline-delimited and sorted")
}

func TestScenariosLsOutputEmpty(t *testing.T) {
	t.Parallel()

	s := scenarios.New()
	cmd := scenarios.Cmd(s)
	cmd.SetArgs([]string{"ls"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	require.Empty(t, strings.TrimSpace(buf.String()), "empty registry should produce no output")
}

func TestScenariosLsHelpText(t *testing.T) {
	t.Parallel()

	s := scenarios.New()
	cmd := scenarios.Cmd(s)
	lsCmd := cmd.Commands()[0]

	require.NotEmpty(t, lsCmd.Short, "ls command should have Short help")
	require.NotEmpty(t, lsCmd.Long, "ls command should have Long help")
}
