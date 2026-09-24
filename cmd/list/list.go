package list

import (
	"encoding/json"
	"strings"

	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List or search saved templates",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		st := types.Stews{}
		if err := st.Load(viper.GetString("stewsPath")); err != nil {
			return err
		}
		search, _ := cmd.Flags().GetString("search")
		tag, _ := cmd.Flags().GetString("tag")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if len(args) > 0 {
			search = args[0]
		}
		filtered := make(types.Stews, 0, len(st))
		entries := make([]listEntry, 0, len(st))
		for _, stew := range st {
			metadata, err := manifest.Load(stew.Path)
			if err != nil {
				return err
			}
			if search != "" && !strings.Contains(strings.ToLower(stew.Name+" "+stew.Description+" "+strings.Join(metadata.Tags, " ")), strings.ToLower(search)) {
				continue
			}
			if tag != "" && !containsTag(metadata.Tags, tag) {
				continue
			}
			filtered = append(filtered, stew)
			entries = append(entries, listEntry{Stew: stew, Manifest: metadata})
		}
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(entries)
		}
		filtered.List()
		return nil
	},
}

type listEntry struct {
	Stew     types.Stew        `json:"stew"`
	Manifest manifest.Manifest `json:"manifest"`
}

func containsTag(tags []string, wanted string) bool {
	for _, tag := range tags {
		if strings.EqualFold(tag, wanted) {
			return true
		}
	}
	return false
}

func init() {
	ListCmd.Flags().String("search", "", "Search template names, descriptions, and tags")
	ListCmd.Flags().String("tag", "", "Only show templates with this tag")
	ListCmd.Flags().Bool("json", false, "Output templates as JSON")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
