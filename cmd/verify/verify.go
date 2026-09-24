package verify

import (
	"fmt"
	"os"

	"github.com/benjuh/stew/bundle"
	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/render"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/vcs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/benjuh/stew/inherit"
)

var VerifyCmd = &cobra.Command{
	Use:   "verify <name_of_stew>",
	Short: "Verify a template source, checksum, and manifest",
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
		resolved, err := inherit.Resolve(viper.GetString("stewsPath"), stew.Name)
		if err != nil {
			return err
		}
		defer resolved.Cleanup()
		if err := validateTemplate(resolved.Root); err != nil {
			return err
		}
		switch stew.SourceType {
		case "archive":
			if stew.Source == "" || stew.Checksum == "" {
				return fmt.Errorf("template %q has no archive checksum metadata", stew.Name)
			}
			if _, err := os.Stat(stew.Source); err != nil {
				return fmt.Errorf("archive source is unavailable: %w", err)
			}
			checksum, err := bundle.Checksum(stew.Source)
			if err != nil {
				return err
			}
			if checksum != stew.Checksum {
				return fmt.Errorf("archive checksum mismatch: expected %s, got %s", stew.Checksum, checksum)
			}
		case "git":
			revision, err := vcs.Revision(stew.Path)
			if err != nil {
				return err
			}
			if stew.Revision != "" && revision != stew.Revision {
				return fmt.Errorf("Git revision mismatch: expected %s, got %s", stew.Revision, revision)
			}
		}
		fmt.Printf("Template %q is verified\n", stew.Name)
		return nil
	},
}

func validateTemplate(path string) error {
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
	return render.Validate(path, names)
}
