package importcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/benjuh/stew/bundle"
	"github.com/benjuh/stew/catalog"
	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ImportCmd = &cobra.Command{
	Use:   "import <archive>",
	Short: "Import a template archive into the local catalog",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		archive := util.ResolvePath(args[0])
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = catalog.NameFromArchive(archive)
		}
		cachePath, err := catalog.CachePath(viper.GetString("templatesPath"), name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
			return err
		}
		keepCache := false
		defer func() {
			if !keepCache {
				_ = os.RemoveAll(cachePath)
			}
		}()
		if err := bundle.Extract(archive, cachePath); err != nil {
			return err
		}
		if err := registerImported(name, cachePath, archive); err != nil {
			return err
		}
		keepCache = true
		return nil
	},
}

func init() {
	ImportCmd.Flags().String("name", "", "Name to register the imported template under")
}

func registerImported(name, path, source string) error {
	metadata, err := manifest.Load(path)
	if err != nil {
		return err
	}
	description := metadata.Description
	if description == "" {
		description = "imported template"
	}
	checksum, err := bundle.Checksum(source)
	if err != nil {
		return err
	}
	entry := types.Stew{
		Name:        name,
		Description: description,
		Path:        path,
		CreatedAt:   time.Now(),
		Source:      source,
		SourceType:  "archive",
		Checksum:    checksum,
	}
	if err := catalog.Register(viper.GetString("stewsPath"), entry); err != nil {
		return err
	}
	fmt.Printf("Imported %q from %s\n", name, filepath.Clean(source))
	return nil
}
