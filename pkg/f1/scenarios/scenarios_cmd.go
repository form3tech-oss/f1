package scenarios

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Cmd(s *Scenarios) *cobra.Command {
	scenariosCmd := &cobra.Command{
		Use:   "scenarios",
		Short: "Prints information about available test scenarios",
	}

	scenariosCmd.AddCommand(lsCmd(s))
	return scenariosCmd
}

func lsCmd(s *Scenarios) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List available scenario names",
		Long:  "List all registered scenario names, one per line, sorted alphabetically.",
		Run:   lsCmdExecute(s),
	}
}

func lsCmdExecute(s *Scenarios) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, _ []string) {
		names := s.GetScenarioNames()
		out := cmd.OutOrStdout()
		for _, name := range names {
			fmt.Fprintln(out, name)
		}
	}
}
