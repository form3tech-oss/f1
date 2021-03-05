package scenarios

import (
	"fmt"
	"sort"

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
	lsCmd := &cobra.Command{
		Use: "ls",
		Run: lsCmdExecute(s),
	}
	return lsCmd
}

func lsCmdExecute(s *Scenarios) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		scenarios := s.GetScenarioNames()
		sort.Strings(scenarios)
		for _, scenario := range scenarios {
			fmt.Println(scenario)
		}
	}
}
