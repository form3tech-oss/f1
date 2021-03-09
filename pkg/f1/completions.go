package f1

import (
	"os"

	"github.com/form3tech-oss/f1/v2/internal/support/errorh"
	"github.com/form3tech-oss/f1/v2/pkg/f1/scenarios"

	"github.com/spf13/cobra"
)

func completionsCmd(s *scenarios.Scenarios, p *profiling) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generates shell completions",
	}
	cmd.AddCommand(bashCmd(s, p))
	cmd.AddCommand(zshCmd(s, p))
	cmd.AddCommand(fishCmd(s, p))
	return cmd
}

func bashCmd(s *scenarios.Scenarios, p *profiling) *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

. <(f1 completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(f1 completion)
`,
		Run: func(cmd *cobra.Command, args []string) {
			errorh.Print(buildRootCmd(s, p).GenBashCompletion(os.Stdout), "error generating bash completion")
		},
	}
}
func zshCmd(s *scenarios.Scenarios, p *profiling) *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generates zsh completion scripts",
		Long: `To load completion run

. <(f1 completion)
`,
		Run: func(cmd *cobra.Command, args []string) {
			errorh.Print(buildRootCmd(s, p).GenZshCompletion(os.Stdout), "error generating zsh completion")
		},
	}
}

func fishCmd(s *scenarios.Scenarios, p *profiling) *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generates fish completion scripts",
		Long: `To define completions run
./f1 completions fish >  ~/.config/fish/completions/f1.fish`,
		Run: func(cmd *cobra.Command, args []string) {
			errorh.Print(buildRootCmd(s, p).GenFishCompletion(os.Stdout, true), "error generating fish completion")
		},
	}
}
