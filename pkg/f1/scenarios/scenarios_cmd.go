package scenarios

import (
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/form3tech-oss/f1/v2/internal/console"
)

func Cmd(s *Scenarios) *cobra.Command {
	scenariosCmd := &cobra.Command{
		Use:   "scenarios",
		Short: "Prints information about available test scenarios",
	}

	// this should be injected, but it's a breaking changes for v2
	printer := console.NewPrinter(os.Stdout, os.Stderr)
	scenariosCmd.AddCommand(lsCmd(s, printer))
	return scenariosCmd
}

func lsCmd(s *Scenarios, printer *console.Printer) *cobra.Command {
	lsCmd := &cobra.Command{
		Use: "ls",
		Run: lsCmdExecute(s, printer),
	}
	return lsCmd
}

func lsCmdExecute(s *Scenarios, printer *console.Printer) func(*cobra.Command, []string) {
	return func(*cobra.Command, []string) {
		scenarios := s.GetScenarioNames()
		sort.Strings(scenarios)
		for _, scenario := range scenarios {
			printer.Println(scenario)
		}
	}
}
