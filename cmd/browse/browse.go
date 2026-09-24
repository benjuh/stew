package browse

import (
	"github.com/benjuh/stew/cmd/search"
	"github.com/spf13/cobra"
)

var BrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse the remote template catalog",
	RunE:  search.SearchCmd.RunE,
}

func init() {
	BrowseCmd.Flags().String("catalog-url", "", "Catalog URL or local index file")
	BrowseCmd.Flags().Bool("json", false, "Output templates as JSON")
}
