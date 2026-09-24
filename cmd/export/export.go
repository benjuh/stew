package export

import (
	"fmt"

	"github.com/benjuh/stew/bundle"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/benjuh/stew/inherit"
)

var ExportCmd = &cobra.Command{
	Use:   "export <name_of_stew> <archive>",
	Short: "Export a template as a portable archive",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
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
		archivePath := util.ResolvePath(args[1])
		resolved, err := inherit.Resolve(viper.GetString("stewsPath"), args[0])
		if err != nil {
			return err
		}
		defer resolved.Cleanup()
		if err := bundle.Create(resolved.Root, archivePath); err != nil {
			return err
		}
		checksum, err := bundle.Checksum(archivePath)
		if err != nil {
			return err
		}
		fmt.Printf("Exported %q to %s (%s)\n", stew.Name, archivePath, checksum)
		return nil
	},
}
