package info

import (
	"encoding/json"
	"fmt"

	"github.com/benjuh/stew/catalog"
	"github.com/benjuh/stew/cmd/catalogsource"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:   "info <template>",
	Short: "Show remote template catalog details",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		url, _ := cmd.Flags().GetString("catalog-url")
		index, err := catalogsource.Load(url)
		if err != nil {
			return err
		}
		template, err := catalog.Find(index, args[0])
		if err != nil {
			return err
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(template)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s\n", template.Name, template.Description)
		fmt.Fprintf(cmd.OutOrStdout(), "Family: %s\nLanguage: %s\n", template.Family, template.Language)
		if len(template.Tags) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Tags: %v\n", template.Tags)
		}
		if len(template.Variants) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "Variants:")
			for _, variant := range template.Variants {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-12s %s\n", variant.ID, variant.Description)
			}
		}
		return nil
	},
}

func init() {
	InfoCmd.Flags().String("catalog-url", "", "Catalog URL or local index file")
	InfoCmd.Flags().Bool("json", false, "Output template details as JSON")
}
