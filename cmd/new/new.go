package new

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/benjuh/stew/inherit"
	"github.com/benjuh/stew/render"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
)

const (
	green     = "\033[32m"
	red       = "\033[31m"
	pathColor = "\033[33m"
	quoted    = "\033[35m"
	reset     = "\033[0m"
)

// CreateCmd creates a project from a saved template.
var CreateCmd = &cobra.Command{
	Use:     "create <name_of_stew> [destination]",
	Aliases: []string{"new"},
	Short:   "Create a project from a template",
	Long:    `stew create <name_of_stew> [destination] [flags]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s := types.Stews{}
		err := s.Load(viper.GetString("stewsPath"))
		if err != nil {
			return err
		}
		if len(args) == 0 {
			return cmd.Help()
		}

		name := args[0]
		path, _ := cmd.Flags().GetString("path")
		force, _ := cmd.Flags().GetBool("force")
		variant, _ := cmd.Flags().GetString("variant")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		valuesPath, _ := cmd.Flags().GetString("values")
		nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
		variables, _ := cmd.Flags().GetStringArray("var")
		if len(args) > 2 {
			return fmt.Errorf("expected a template name and optional destination")
		}
		if len(args) == 2 {
			if path != "" {
				return fmt.Errorf("destination provided both as an argument and with --path")
			}
			path = args[1]
		}
		if variant != "" {
			name += "-" + variant
		}

		// get Stew by name
		stew, err := s.GetByName(name)
		if err != nil {
			return err
		}

		if !util.CheckIfDirExists(stew.Path) {
			fmt.Printf("\n%v%v%v\n", red, stew.Path, reset)
			fmt.Println("This stew location has been deleted")
			fmt.Println("Either edit the path or remove the stew")
			return fmt.Errorf("stew source directory does not exist: %s", stew.Path)
		}

		path = util.ResolvePath(path)

		values := make(map[string]string, len(variables))
		for _, value := range variables {
			key, val, ok := strings.Cut(value, "=")
			if !ok || strings.TrimSpace(key) == "" {
				return fmt.Errorf("invalid variable %q; expected key=value", value)
			}
			values[strings.TrimSpace(key)] = val
		}
		resolved, err := inherit.Resolve(viper.GetString("stewsPath"), name)
		if err != nil {
			return err
		}
		defer resolved.Cleanup()
		metadata := resolved.Manifest
		values, err = metadata.Resolve(values, valuesPath, cmd.InOrStdin(), cmd.ErrOrStderr(), nonInteractive)
		if err != nil {
			return err
		}
		changes, err := render.Generate(resolved.Root, path, render.Options{
			DryRun:    dryRun,
			Overwrite: force,
			Variables: values,
		})
		if err != nil {
			return err
		}
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(changes)
		}
		if dryRun {
			fmt.Printf("Would change %d file(s)\n", len(changes))
		} else {
			fmt.Printf("🎉 Project created successfully (%d file(s))\n", len(changes))
		}
		return nil

	},
}

// NewCmd is retained as a source-level compatibility alias.
var NewCmd = CreateCmd

func flags() {
	CreateCmd.Flags().StringP("path", "p", "", "Destination directory (defaults to current directory)")
	CreateCmd.Flags().BoolP("force", "f", false, "Force the creation of the project")
	CreateCmd.Flags().String("variant", "", "Catalog variant previously installed for this template")
	CreateCmd.Flags().StringArray("var", nil, "Template variable in key=value form (repeatable)")
	CreateCmd.Flags().String("values", "", "YAML file containing template variables")
	CreateCmd.Flags().Bool("non-interactive", false, "Fail instead of prompting for missing required variables")
	CreateCmd.Flags().Bool("dry-run", false, "Show planned changes without writing files")
	CreateCmd.Flags().Bool("json", false, "Output the planned or completed changes as JSON")
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	flags()
}
