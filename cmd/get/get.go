package get

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/benjuh/stew/inherit"
)

// ViewCmd inspects a saved template and its source tree.
var ViewCmd = &cobra.Command{
	Use:     "view <name_of_stew>",
	Aliases: []string{"get"},
	Short:   "View a template and its project tree",
	Long:    `stew view <name_of_stew> [flags]`,
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

		noTree, _ := cmd.Flags().GetBool("no-tree")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		maxDepth, _ := cmd.Flags().GetInt("depth")
		if maxDepth < 0 {
			return fmt.Errorf("depth cannot be negative")
		}
		resolved, err := inherit.Resolve(viper.GetString("stewsPath"), args[0])
		if err != nil {
			return err
		}
		defer resolved.Cleanup()
		metadata := resolved.Manifest
		var tree *util.TreeNode
		if !noTree {
			node, treeErr := util.BuildTree(resolved.Root, maxDepth)
			if treeErr != nil {
				return treeErr
			}
			tree = &node
		}
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(viewOutput{
				Stew:     *stew,
				Manifest: metadata,
				Tree:     tree,
			})
		}

		stew.Print()
		if stew.Source != "" {
			fmt.Printf("Source: %s\n", stew.Source)
			fmt.Printf("Source type: %s\n", stew.SourceType)
			if stew.Version != "" {
				fmt.Printf("Version: %s\n", stew.Version)
			}
			if stew.Checksum != "" {
				fmt.Printf("Checksum: %s\n", stew.Checksum)
			}
			if stew.Revision != "" {
				fmt.Printf("Revision: %s\n", stew.Revision)
			}
			if !stew.UpdatedAt.IsZero() {
				fmt.Printf("Updated at: %s\n", util.FormatTime(stew.UpdatedAt))
			}
		}
		if metadata.Description != "" || len(metadata.Tags) > 0 || len(metadata.Variables) > 0 {
			fmt.Println("Manifest:")
			if metadata.Description != "" {
				fmt.Printf("  Description: %s\n", metadata.Description)
			}
			if len(metadata.Tags) > 0 {
				fmt.Printf("  Tags: %v\n", metadata.Tags)
			}
			if len(metadata.Variables) > 0 {
				fmt.Println("  Variables:")
				for _, variable := range metadata.Variables {
					details := make([]string, 0, 2)
					if variable.Required {
						details = append(details, "required")
					}
					if variable.Default != "" {
						details = append(details, "default="+variable.Default)
					}
					line := variable.Name
					if len(details) > 0 {
						line += " (" + strings.Join(details, ", ") + ")"
					}
					if variable.Description != "" {
						line += " - " + variable.Description
					}
					fmt.Printf("    - %s\n", line)
				}
			}
		}
		if tree != nil {
			fmt.Println("\nTemplate tree:")
			util.PrintTreeNode(*tree)
		}
		return nil
	},
}

type viewOutput struct {
	Stew     types.Stew        `json:"stew"`
	Manifest manifest.Manifest `json:"manifest"`
	Tree     *util.TreeNode    `json:"tree,omitempty"`
}

func flags() {
	ViewCmd.Flags().BoolP("tree", "t", false, "Print the tree (retained for compatibility)")
	ViewCmd.Flags().Bool("no-tree", false, "Do not print the template tree")
	ViewCmd.Flags().Int("depth", 0, "Limit tree depth; zero means unlimited")
	ViewCmd.Flags().Bool("json", false, "Output template details and tree as JSON")
}

func init() {
	flags()
}
