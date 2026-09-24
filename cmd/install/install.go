package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/benjuh/stew/bundle"
	"github.com/benjuh/stew/catalog"
	"github.com/benjuh/stew/cmd/catalogsource"
	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
	"github.com/benjuh/stew/vcs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var InstallCmd = &cobra.Command{
	Use:   "install <source> [name]",
	Short: "Install a template from a directory, archive, or Git URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			return cmd.Help()
		}
		source := args[0]
		name := ""
		if len(args) == 2 {
			name = args[1]
		}
		variant, _ := cmd.Flags().GetString("variant")
		variantPath := ""
		if variant != "" {
			catalogURL, _ := cmd.Flags().GetString("catalog-url")
			index, err := catalogsource.Load(catalogURL)
			if err != nil {
				return err
			}
			template, err := catalog.Find(index, source)
			if err != nil {
				return err
			}
			selected, err := template.FindVariant(variant)
			if err != nil {
				return err
			}
			source = selected.Source
			variantPath = selected.Path
			if name == "" {
				name = template.ID + "-" + selected.ID
			}
			if ref, _ := cmd.Flags().GetString("ref"); ref == "" {
				_ = cmd.Flags().Set("ref", selected.Ref)
			}
		}
		if name == "" {
			if info, err := os.Stat(util.ResolvePath(source)); err == nil && info.IsDir() {
				name = filepath.Base(filepath.Clean(source))
			} else if vcs.IsGitSource(source) {
				name = catalog.NameFromGitURL(source)
			} else {
				name = catalog.NameFromArchive(source)
			}
		}
		cachePath, err := catalog.CachePath(viper.GetString("templatesPath"), name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
			return err
		}
		if _, err := os.Stat(cachePath); err == nil {
			return fmt.Errorf("template cache path already exists: %s", cachePath)
		} else if !os.IsNotExist(err) {
			return err
		}
		keepCache := false
		defer func() {
			if !keepCache {
				_ = os.RemoveAll(cachePath)
			}
		}()

		var sourceType, version, checksum, revision string
		resolvedSource := util.ResolvePath(source)
		if info, statErr := os.Stat(resolvedSource); statErr == nil && info.IsDir() {
			if err := util.CopyDir(resolvedSource, cachePath); err != nil {
				return err
			}
			sourceType = "local"
		} else if info, statErr := os.Stat(resolvedSource); statErr == nil && !info.IsDir() {
			if err := bundle.Extract(resolvedSource, cachePath); err != nil {
				return err
			}
			sourceType = "archive"
			checksum, err = bundle.Checksum(resolvedSource)
			if err != nil {
				return err
			}
		} else {
			if !vcs.IsGitSource(source) {
				return fmt.Errorf("source does not exist or is not a supported Git URL: %s", source)
			}
			ref, _ := cmd.Flags().GetString("ref")
			clonePath := cachePath
			if variantPath != "" {
				clonePath, err = os.MkdirTemp(filepath.Dir(cachePath), ".stew-source-")
				if err != nil {
					return err
				}
				defer os.RemoveAll(clonePath)
			}
			if err := vcs.Clone(source, clonePath, ref); err != nil {
				return err
			}
			if variantPath != "" {
				selectedPath, pathErr := safeVariantPath(clonePath, variantPath)
				if pathErr != nil {
					return pathErr
				}
				if err := util.CopyDir(selectedPath, cachePath); err != nil {
					return err
				}
			}
			sourceType = "git"
			version = ref
			revision, err = vcs.Revision(clonePath)
			if err != nil {
				return err
			}
		}

		metadata, err := manifest.Load(cachePath)
		if err != nil {
			return err
		}
		description := metadata.Description
		if description == "" {
			description = "installed template"
		}
		entry := types.Stew{
			Name:        name,
			Description: description,
			Path:        cachePath,
			CreatedAt:   time.Now(),
			Source:      source,
			SourceType:  sourceType,
			Version:     version,
			Checksum:    checksum,
			Revision:    revision,
			UpdatedAt:   time.Now(),
		}
		if err := catalog.Register(viper.GetString("stewsPath"), entry); err != nil {
			return err
		}
		keepCache = true
		fmt.Printf("Installed %q\n", name)
		return nil
	},
}

func safeVariantPath(root, relative string) (string, error) {
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("catalog variant path must be relative: %s", relative)
	}
	path := filepath.Join(root, filepath.Clean(relative))
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("catalog variant path escapes source repository: %s", relative)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("catalog variant path %q: %w", relative, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("catalog variant path is not a directory: %s", relative)
	}
	return path, nil
}

func init() {
	InstallCmd.Flags().String("ref", "", "Git branch, tag, or commit to install")
	InstallCmd.Flags().String("variant", "", "Catalog variant to install")
	InstallCmd.Flags().String("catalog-url", "", "Catalog URL or local index file")
}
