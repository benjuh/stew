package completion

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	Use:       "completion <shell>",
	Short:     "Generate shell completion scripts",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := strings.ToLower(args[0])
		root := cmd.Root()
		out := cmd.OutOrStdout()
		switch shell {
		case "bash":
			return root.GenBashCompletionV2(out, true)
		case "zsh":
			return root.GenZshCompletion(out)
		case "fish":
			return root.GenFishCompletion(out, true)
		case "powershell":
			return root.GenPowerShellCompletion(out)
		default:
			return fmt.Errorf("unsupported shell %q; choose bash, zsh, fish, or powershell", args[0])
		}
	},
}
