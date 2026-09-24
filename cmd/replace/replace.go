package replace

import (
	"fmt"

	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
)

const (
	green     = "\033[32m"
	red       = "\033[31m"
	pathColor = "\033[33m"
	quoted    = "\033[35m"
	reset     = "\033[0m"
)

// ReplaceCmd represents the replace command
var ReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace a project name in a stew",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) < 2 {
			return cmd.Help()
		}
		oldString := args[0]
		newString := args[1]

		path, _ := cmd.Flags().GetString("path")
		ignoreCase, _ := cmd.Flags().GetBool("ignore-case")

		path, err := util.GetPath(path)
		if err != nil {
			return err
		}

		if path == "" {
			path = util.GetCurrentDir()
		}

		count, err := util.UpdateProjectName(path, oldString, newString, ignoreCase)
		if err != nil {
			return err
		}

		fmt.Printf("\nReplaced %v%v%v instances of %v%v%v with %v%v%v\n", green, count, reset, red, oldString, reset, quoted, newString, reset)
		return nil

	},
}

func flags() {
	ReplaceCmd.Flags().StringP("path", "p", "", "The path to the stew (defaults to current directory)")
	ReplaceCmd.Flags().BoolP("ignore-case", "i", false, "Whether to replace the string case sensitively")
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// replaceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// replaceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	flags()
}
