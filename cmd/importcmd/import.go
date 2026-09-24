package importcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/benjuh/stew/bundle"
	"github.com/benjuh/stew/catalog"
	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
	"github.com/benjuh/stew/vcs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ImportCmd = &cobra.Command{
	Use:   "import <archive-or-git-url>",
	Short: "Import an archive or curate a Git project into the local catalog",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		source := args[0]
		if vcs.IsGitSource(source) {
			return importGit(cmd, source)
		}
		archive := util.ResolvePath(source)
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
	ImportCmd.Flags().String("profile", "", "Curation profile: auto, generic, react, node, or go")
	ImportCmd.Flags().StringSlice("exclude", nil, "Additional file or directory names to exclude")
	ImportCmd.Flags().Bool("dry-run", false, "Show what would be imported without saving it")
	ImportCmd.Flags().String("description", "", "Description for an imported Git template")
}

type curationProfile struct {
	Name     string
	Excludes []string
	Tags     []string
}

var profiles = map[string]curationProfile{
	"generic": {Name: "generic", Excludes: []string{".git", ".DS_Store", "node_modules", "dist", "build", "coverage", ".next", "out", ".turbo", ".env", ".env.local", ".env.development", ".env.production"}},
	"react":   {Name: "react", Excludes: []string{".git", ".DS_Store", "node_modules", "dist", "build", "coverage", ".next", "out", ".turbo", ".env", ".env.local", ".env.development", ".env.production"}, Tags: []string{"react", "typescript"}},
	"node":    {Name: "node", Excludes: []string{".git", ".DS_Store", "node_modules", "dist", "build", "coverage", ".next", "out", ".turbo", ".env", ".env.local", ".env.development", ".env.production"}, Tags: []string{"node"}},
	"go":      {Name: "go", Excludes: []string{".git", ".DS_Store", "bin", "coverage", "tmp", ".env", ".env.local", ".env.development", ".env.production"}, Tags: []string{"go"}},
}

func importGit(cmd *cobra.Command, source string) error {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		name = catalog.NameFromGitURL(source)
	}
	profileName, _ := cmd.Flags().GetString("profile")
	excludes, _ := cmd.Flags().GetStringSlice("exclude")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	description, _ := cmd.Flags().GetString("description")

	checkout, err := os.MkdirTemp("", "stew-import-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(checkout)
	if err := vcs.Clone(source, checkout, ""); err != nil {
		return err
	}
	profile, err := selectProfile(checkout, profileName)
	if err != nil {
		return err
	}
	profile.Excludes = append(profile.Excludes, excludes...)

	files, err := curatedFiles(checkout, profile.Excludes)
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "Profile: %s\nFiles to import: %d\n", profile.Name, len(files))
		for _, file := range files {
			fmt.Fprintln(cmd.OutOrStdout(), file)
		}
		return nil
	}

	cachePath, err := catalog.CachePath(viper.GetString("templatesPath"), name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(cachePath); err == nil {
		return fmt.Errorf("template cache path already exists: %s", cachePath)
	} else if !os.IsNotExist(err) {
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
	if err := copyCurated(checkout, cachePath, profile.Excludes); err != nil {
		return err
	}
	if err := writeGeneratedManifest(cachePath, profile, description); err != nil {
		return err
	}
	revision, err := vcs.Revision(checkout)
	if err != nil {
		return err
	}
	if err := registerImportedGit(name, cachePath, source, revision); err != nil {
		return err
	}
	keepCache = true
	return nil
}

func selectProfile(root, requested string) (curationProfile, error) {
	if requested == "" || requested == "auto" {
		if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
			if strings.Contains(strings.ToLower(string(data)), "react") {
				return profiles["react"], nil
			}
			return profiles["node"], nil
		}
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return profiles["go"], nil
		}
		return profiles["generic"], nil
	}
	profile, ok := profiles[strings.ToLower(requested)]
	if !ok {
		return curationProfile{}, fmt.Errorf("unknown curation profile %q; choose auto, generic, react, node, or go", requested)
	}
	return profile, nil
}

func curatedFiles(root string, excludes []string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if excluded(rel, info.IsDir(), excludes) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func copyCurated(source, destination string, excludes []string) error {
	files, err := curatedFiles(source, excludes)
	if err != nil {
		return err
	}
	for _, relative := range files {
		sourcePath := filepath.Join(source, filepath.FromSlash(relative))
		destinationPath := filepath.Join(destination, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
			return err
		}
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(destinationPath, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

func excluded(relative string, directory bool, excludes []string) bool {
	for _, part := range strings.Split(filepath.ToSlash(relative), "/") {
		for _, pattern := range excludes {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}
			if part == pattern || strings.HasPrefix(part, pattern+".") {
				return true
			}
		}
	}
	return false
}

func writeGeneratedManifest(path string, profile curationProfile, description string) error {
	if _, err := os.Stat(filepath.Join(path, manifest.Filename)); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if description == "" {
		description = "Imported " + profile.Name + " project template"
	}
	data, err := yaml.Marshal(manifest.Manifest{Description: description, Tags: profile.Tags})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, manifest.Filename), data, 0644)
}

func registerImportedGit(name, path, source, revision string) error {
	metadata, err := manifest.Load(path)
	if err != nil {
		return err
	}
	description := metadata.Description
	if description == "" {
		description = "imported Git template"
	}
	entry := types.Stew{
		Name:        name,
		Description: description,
		Path:        path,
		CreatedAt:   time.Now(),
		Source:      source,
		SourceType:  "git",
		Revision:    revision,
		UpdatedAt:   time.Now(),
	}
	if err := catalog.Register(viper.GetString("stewsPath"), entry); err != nil {
		return err
	}
	fmt.Printf("Imported %q from %s\n", name, source)
	return nil
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
