package update

import (
	"fmt"
	"time"

	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/render"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/vcs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var UpdateCmd = &cobra.Command{
	Use:   "update <name_of_stew>",
	Short: "Update a Git-backed template",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		stews := types.Stews{}
		registryPath := viper.GetString("stewsPath")
		if err := stews.Load(registryPath); err != nil {
			return err
		}
		stew, err := stews.GetByName(args[0])
		if err != nil {
			return err
		}
		if stew.SourceType != "git" {
			return fmt.Errorf("template %q is not Git-backed", stew.Name)
		}
		dirty, err := vcs.Dirty(stew.Path)
		if err != nil {
			return err
		}
		if dirty {
			return fmt.Errorf("template %q has local modifications; refusing to update", stew.Name)
		}
		if err := vcs.Fetch(stew.Path); err != nil {
			return err
		}
		ref, _ := cmd.Flags().GetString("ref")
		if ref == "" {
			ref = stew.Version
		}
		remoteRevision, err := vcs.RemoteRevision(stew.Path, ref)
		if err != nil {
			return err
		}
		currentRevision, err := vcs.Revision(stew.Path)
		if err != nil {
			return err
		}
		if currentRevision == remoteRevision {
			fmt.Printf("%q is already up to date (%s)\n", stew.Name, shortRevision(currentRevision))
			return nil
		}

		check, _ := cmd.Flags().GetBool("check")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if check || dryRun {
			fmt.Printf("%q can update from %s to %s\n", stew.Name, shortRevision(currentRevision), shortRevision(remoteRevision))
			return nil
		}
		if err := vcs.FastForward(stew.Path, remoteRevision); err != nil {
			return err
		}
		if err := validateUpdatedTemplate(stew.Path); err != nil {
			_ = vcs.Reset(stew.Path, currentRevision)
			return err
		}
		stew.Revision = remoteRevision
		stew.UpdatedAt = time.Now()
		if ref != "" {
			stew.Version = ref
		}
		if err := stews.Save(registryPath); err != nil {
			return err
		}
		fmt.Printf("Updated %q to %s\n", stew.Name, shortRevision(remoteRevision))
		return nil
	},
}

func init() {
	UpdateCmd.Flags().String("ref", "", "Git branch, tag, or commit to update to")
	UpdateCmd.Flags().Bool("check", false, "Check for updates without changing files")
	UpdateCmd.Flags().Bool("dry-run", false, "Show the update without changing files")
}

func validateUpdatedTemplate(path string) error {
	metadata, err := manifest.Load(path)
	if err != nil {
		return fmt.Errorf("load updated template manifest: %w", err)
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

func shortRevision(revision string) string {
	if len(revision) > 12 {
		return revision[:12]
	}
	return revision
}
