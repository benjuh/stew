package outdated

import (
	"encoding/json"
	"fmt"

	"github.com/benjuh/stew/types"
	"github.com/benjuh/stew/vcs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var OutdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "Check installed Git templates for updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		stews := types.Stews{}
		if err := stews.Load(viper.GetString("stewsPath")); err != nil {
			return err
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")
		results := make([]result, 0)
		for _, stew := range stews {
			if stew.SourceType != "git" {
				continue
			}
			if err := vcs.Fetch(stew.Path); err != nil {
				return err
			}
			current, err := vcs.Revision(stew.Path)
			if err != nil {
				return err
			}
			remote, err := vcs.RemoteRevision(stew.Path, stew.Version)
			if err != nil {
				return err
			}
			results = append(results, result{Name: stew.Name, Current: current, Remote: remote, Outdated: current != remote})
		}
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(results)
		}
		for _, item := range results {
			if item.Outdated {
				fmt.Printf("%s: update available (%s -> %s)\n", item.Name, short(item.Current), short(item.Remote))
			} else {
				fmt.Printf("%s: up to date (%s)\n", item.Name, short(item.Current))
			}
		}
		return nil
	},
}

type result struct {
	Name     string `json:"name"`
	Current  string `json:"current"`
	Remote   string `json:"remote"`
	Outdated bool   `json:"outdated"`
}

func init() {
	OutdatedCmd.Flags().Bool("json", false, "Output update status as JSON")
}

func short(revision string) string {
	if len(revision) > 12 {
		return revision[:12]
	}
	return revision
}
