package scenarios

import (
	"fmt"
	"github.com/form3tech-oss/f1/pkg/f1/testing"
	"sort"

	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	scenariosCmd := &cobra.Command{
		Use:   "scenarios",
		Short: "Prints information about available test scenarios",
	}
	scenariosCmd.AddCommand(lsCmd())
	return scenariosCmd
}

func lsCmd() *cobra.Command {
	lsCmd := &cobra.Command{
		Use: "ls",
		Run: lsCmdExecute(),
	}
	return lsCmd
}

func lsCmdExecute() func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		scenarios := testing.GetScenarioNames()
		sort.Strings(scenarios)
		for _, scenario := range scenarios {
			fmt.Println(scenario)
		}
	}
}
