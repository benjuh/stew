package validate

import (
	"fmt"

	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/render"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/benjuh/stew/inherit"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate <name_of_stew>",
	Short: "Validate a template and its manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
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
		path, _ := cmd.Flags().GetString("path")
		if path != "" {
			path = util.ResolvePath(path)
		} else {
			resolved, resolveErr := inherit.Resolve(viper.GetString("stewsPath"), stew.Name)
			if resolveErr != nil {
				return resolveErr
			}
			defer resolved.Cleanup()
			path = resolved.Root
		}
		metadata, err := manifest.Load(path)
		if err != nil {
			return fmt.Errorf("load template manifest: %w", err)
		}
		if err := metadata.Validate(); err != nil {
			return err
		}
		names := make([]string, 0, len(metadata.Variables))
		for _, variable := range metadata.Variables {
			names = append(names, variable.Name)
		}
		if err := render.Validate(path, names); err != nil {
			return err
		}
		fmt.Printf("Template %q is valid\n", stew.Name)
		return nil
	},
}

func init() {
	ValidateCmd.Flags().StringP("path", "p", "", "Validate this directory instead of the saved template path")
}
