package f1

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func completionsCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generates shell completions",
	}
	cmd.AddCommand(bashCmd(rootCmd))
	cmd.AddCommand(zshCmd(rootCmd))
	cmd.AddCommand(fishCmd(rootCmd))
	return cmd
}

func bashCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generates bash completion scripts",
		Long: `To load completion run

. <(f1 completion)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(f1 completion)
`,
		RunE: func(*cobra.Command, []string) error {
			if err := rootCmd.GenBashCompletionV2(os.Stdout, true); err != nil {
				return fmt.Errorf("generating bash completion: %w", err)
			}

			return nil
		},
	}
}

func zshCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generates zsh completion scripts",
		Long: `To load completion run

. <(f1 completion)
`,
		RunE: func(*cobra.Command, []string) error {
			if err := rootCmd.GenZshCompletion(os.Stdout); err != nil {
				return fmt.Errorf("generating zsh completion: %w", err)
			}
			return nil
		},
	}
}

func fishCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generates fish completion scripts",
		Long: `To define completions run
./f1 completions fish >  ~/.config/fish/completions/f1.fish`,
		RunE: func(*cobra.Command, []string) error {
			if err := rootCmd.GenFishCompletion(os.Stdout, true); err != nil {
				return fmt.Errorf("generating fish completion: %w", err)
			}
			return nil
		},
	}
}
