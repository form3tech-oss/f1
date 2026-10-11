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
		Long: fmt.Sprintf(`To load completion run

. <(%[1]s generate bash)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(%[1]s generate bash)
`, rootCmd.Name()),
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
		Long: fmt.Sprintf(`To load completion run

source <(%s generate zsh)
`, rootCmd.Name()),
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
		Long: fmt.Sprintf(`To define completions run
%[1]s generate fish > ~/.config/fish/completions/%[1]s.fish`, rootCmd.Name()),
		RunE: func(*cobra.Command, []string) error {
			if err := rootCmd.GenFishCompletion(os.Stdout, true); err != nil {
				return fmt.Errorf("generating fish completion: %w", err)
			}
			return nil
		},
	}
}
