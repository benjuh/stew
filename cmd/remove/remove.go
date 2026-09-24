package remove

import (
	"github.com/benjuh/stew/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RemoveCmd represents the delete command
var RemoveCmd = &cobra.Command{
	Use:   "remove <name_of_stew>",
	Short: "Remove a stew",
	Long:  `stew remove <name_of_stew> [flags]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s := types.Stews{}
		err := s.Load(viper.GetString("stewsPath"))
		if err != nil {
			return err
		}

		if len(args) == 0 {
			name, _ := cmd.Flags().GetString("name")
			id, _ := cmd.Flags().GetInt("id")
			if name == "" && id < 0 {
				return cmd.Help()
			}
			if name != "" {
				args = []string{name}
			} else {
				stew, getErr := s.Get(id)
				if getErr != nil {
					return getErr
				}
				args = []string{stew.Name}
			}
		}

		name := args[0]
		_, err = s.GetByName(name)
		if err != nil {
			return err
		}

		err = s.RemoveByName(name)
		if err != nil {
			return err
		}

		err = s.Save(viper.GetString("stewsPath"))
		if err != nil {
			return err
		}
		return nil
	},
}

func flags() {
	RemoveCmd.Flags().IntP("id", "i", -1, "The id of the stew you want to remove")
	RemoveCmd.Flags().StringP("name", "n", "", "The name of the stew you want to remove")
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	flags()
}
