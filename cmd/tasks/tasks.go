package tasks

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/benjuh/stew/tasks"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
)

var TasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List available project tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := tasks.FindProjectRoot(util.GetCurrentDir())
		if err != nil {
			return err
		}
		available, err := tasks.Available(root)
		if err != nil {
			return err
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(available)
		}
		if len(available) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No tasks found")
			return nil
		}
		for _, task := range available {
			command := []string{strconv.Quote(task.Task.Command)}
			for _, arg := range task.Task.Args {
				command = append(command, strconv.Quote(arg))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-10s %-10s %-32s %s\n", task.Name, task.Source, task.Description, strings.Join(command, " "))
		}
		return nil
	},
}

func init() {
	TasksCmd.Flags().Bool("json", false, "Output tasks as JSON")
}
