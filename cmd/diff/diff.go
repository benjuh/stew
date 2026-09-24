package diff

import (
	"fmt"

	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/vcs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var DiffCmd = &cobra.Command{
	Use:   "diff <name_of_stew>",
	Short: "Show changes available for a Git-backed template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		stews := types.Stews{}
		if err := stews.Load(viper.GetString("stewsPath")); err != nil {
			return err
		}
		stew, err := stews.GetByName(args[0])
		if err != nil {
			return err
		}
		if stew.SourceType != "git" {
			return fmt.Errorf("template %q is not Git-backed", stew.Name)
		}
		if err := vcs.Fetch(stew.Path); err != nil {
			return err
		}
		from := stew.Revision
		if from == "" {
			from, err = vcs.Revision(stew.Path)
			if err != nil {
				return err
			}
		}
		to, err := vcs.RemoteRevision(stew.Path, stew.Version)
		if err != nil {
			return err
		}
		if from == to {
			fmt.Printf("%q has no available changes\n", stew.Name)
			return nil
		}
		changes, err := vcs.Diff(stew.Path, from, to)
		if err != nil {
			return err
		}
		fmt.Print(changes)
		return nil
	},
}
