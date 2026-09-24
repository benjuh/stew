package add

import (
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"os"
)

// SaveCmd stores a directory as a named template.
var SaveCmd = &cobra.Command{
	Use:     "save <name_of_stew>",
	Aliases: []string{"add"},
	Short:   "Save a directory as a template",
	Long:    `stew save <name_of_stew> [flags]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		name := args[0]
		description, _ := cmd.Flags().GetString("description")
		path, _ := cmd.Flags().GetString("path")

		st := types.Stews{}

		if err := st.Load(viper.GetString("stewsPath")); err != nil {
			return err
		}

		// get absolute path
		path, err := util.GetPath(path)
		if err != nil {
			return err
		}

		// if description is empty, set it to no description provided
		if description == "" {
			description = "no description provided"
		}

		if err = st.Add(name, description, path); err != nil {
			return err
		}
		err = st.Save(viper.GetString("stewsPath"))
		if err != nil {
			return err
		}
		return nil
	},
}

// AddCmd is retained as a source-level compatibility alias.
var AddCmd = SaveCmd

func getCWD() string {
	cwd, _ := os.Getwd()
	return cwd
}

func addFlag() {
	SaveCmd.Flags().StringP("description", "d", "no description provided", "Description of the template")
	SaveCmd.Flags().StringP("path", "p", getCWD(), "Path to the template")

}

func init() {
	addFlag()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
