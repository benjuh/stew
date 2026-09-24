package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/render"
	"github.com/benjuh/stew/tasks"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
)

type Diagnostic struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose the current project and its tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		if path == "" {
			path = util.GetCurrentDir()
		} else {
			path = util.ResolvePath(path)
		}
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("doctor path is not a directory: %s", path)
		}

		diagnostics := diagnose(path)
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			if err := json.NewEncoder(cmd.OutOrStdout()).Encode(diagnostics); err != nil {
				return err
			}
		} else {
			for _, diagnostic := range diagnostics {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", diagnostic.Level, diagnostic.Message)
			}
		}
		for _, diagnostic := range diagnostics {
			if diagnostic.Level == "ERROR" {
				return fmt.Errorf("doctor found project issues")
			}
		}
		return nil
	},
}

func init() {
	DoctorCmd.Flags().String("path", "", "Project directory to diagnose (defaults to current directory)")
	DoctorCmd.Flags().Bool("json", false, "Output diagnostics as JSON")
}

func diagnose(root string) []Diagnostic {
	diagnostics := make([]Diagnostic, 0)
	metadata, err := manifest.Load(root)
	if err != nil {
		return []Diagnostic{{Level: "ERROR", Message: fmt.Sprintf("manifest: %v", err)}}
	}
	_, manifestErr := os.Stat(filepath.Join(root, manifest.Filename))
	manifestPresent := manifestErr == nil
	if os.IsNotExist(manifestErr) {
		diagnostics = append(diagnostics, Diagnostic{Level: "INFO", Message: "no .stew.yaml manifest found"})
	} else if manifestErr != nil {
		diagnostics = append(diagnostics, Diagnostic{Level: "ERROR", Message: fmt.Sprintf("manifest: %v", manifestErr)})
	} else if err := metadata.Validate(); err != nil {
		diagnostics = append(diagnostics, Diagnostic{Level: "ERROR", Message: fmt.Sprintf("manifest: %v", err)})
	} else {
		diagnostics = append(diagnostics, Diagnostic{Level: "OK", Message: "manifest is valid"})
	}

	if !manifestPresent {
		diagnostics = append(diagnostics, Diagnostic{Level: "INFO", Message: "template rendering check skipped"})
	} else {
		variables := make([]string, 0, len(metadata.Variables))
		for _, variable := range metadata.Variables {
			variables = append(variables, variable.Name)
		}
		if err := render.Validate(root, variables); err != nil {
			diagnostics = append(diagnostics, Diagnostic{Level: "ERROR", Message: fmt.Sprintf("template rendering: %v", err)})
		} else {
			diagnostics = append(diagnostics, Diagnostic{Level: "OK", Message: "template files are renderable"})
		}
	}

	issues := tasks.ValidateProject(root)
	for _, issue := range issues {
		diagnostics = append(diagnostics, Diagnostic{Level: "ERROR", Message: issue})
	}
	if len(issues) == 0 {
		diagnostics = append(diagnostics, Diagnostic{Level: "OK", Message: "tasks are valid"})
	}
	return diagnostics
}
