package search

import (
	"encoding/json"
	"fmt"

	"github.com/benjuh/stew/catalog"
	"github.com/benjuh/stew/cmd/catalogsource"
	"github.com/spf13/cobra"
)

var SearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the remote template catalog",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return cmd.Help()
		}
		url, _ := cmd.Flags().GetString("catalog-url")
		index, err := catalogsource.Load(url)
		if err != nil {
			return err
		}
		query := ""
		if len(args) == 1 {
			query = args[0]
		}
		matches := catalog.Search(index, query)
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(matches)
		}
		for _, template := range matches {
			fmt.Fprintf(cmd.OutOrStdout(), "%-36s %-14s %s\n", template.ID, template.Language, template.Description)
		}
		return nil
	},
}

func init() {
	SearchCmd.Flags().String("catalog-url", "", "Catalog URL or local index file")
	SearchCmd.Flags().Bool("json", false, "Output templates as JSON")
}
