package scenarios

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/form3tech-oss/f1/v2/internal/ui"
)

func Cmd(s *Scenarios) *cobra.Command {
	scenariosCmd := &cobra.Command{
		Use:   "scenarios",
		Short: "Prints information about available test scenarios",
	}

	// this should be injected, but it's a breaking change
	outputer := ui.NewConoleOnlyOutput()
	scenariosCmd.AddCommand(lsCmd(s, outputer))
	return scenariosCmd
}

func lsCmd(s *Scenarios, outputer ui.Outputer) *cobra.Command {
	lsCmd := &cobra.Command{
		Use: "ls",
		Run: lsCmdExecute(s, outputer),
	}
	return lsCmd
}

func lsCmdExecute(s *Scenarios, outputer ui.Outputer) func(*cobra.Command, []string) {
	return func(*cobra.Command, []string) {
		scenarios := s.GetScenarioNames()
		sort.Strings(scenarios)
		for _, scenario := range scenarios {
			outputer.Display(ui.InteractiveMessage{Message: scenario})
		}
	}
}
