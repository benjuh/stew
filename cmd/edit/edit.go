package edit

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
)

// EditCmd represents the edit command
var EditCmd = &cobra.Command{
	Use:   "edit <name_of_stew>",
	Short: "Edit a stew's name, description, or path",
	Long:  `stew edit <name_of_stew> [flags]`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return cmd.Help()
		}

		selectedStew := args[0]
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		path, _ := cmd.Flags().GetString("path")

		if name == "" && description == "" && path == "" {
			return cmd.Help()
		}

		stews := types.Stews{}

		err := stews.Load(viper.GetString("stewsPath"))
		if err != nil {
			return err
		}

		stew, err := stews.GetByName(selectedStew)
		if err != nil {
			return err
		}
		if name != "" && name != selectedStew {
			if _, err := stews.GetByName(name); err == nil {
				return fmt.Errorf("stew name already exists: %s", name)
			}
		}
		if path != "" {
			resolvedPath, pathErr := util.GetPath(path)
			if pathErr != nil {
				return pathErr
			}
			for _, candidate := range stews {
				if candidate.Name != selectedStew && candidate.Path == resolvedPath {
					return fmt.Errorf("stew path already exists: %s", resolvedPath)
				}
			}
		}

		err = stew.Edit(name, description, path)
		if err != nil {
			return err
		}
		err = stews.Save(viper.GetString("stewsPath"))
		if err != nil {
			return err
		}
		return nil
	},
}

func flags() {
	EditCmd.Flags().StringP("name", "n", "", "The new name of the stew")
	EditCmd.Flags().StringP("description", "d", "", "The new description of the stew")
	EditCmd.Flags().StringP("path", "p", "", "The new path of the stew")
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	flags()
}
