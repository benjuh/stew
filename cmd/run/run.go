package run

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/benjuh/stew/tasks"
	"github.com/benjuh/stew/util"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run <task>",
	Short: "Run a named project task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Help()
		}
		root, err := tasks.FindProjectRoot(util.GetCurrentDir())
		if err != nil {
			return err
		}
		plan, err := tasks.Plan(root, args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "Project root: %s\nExecution plan:\n", root)
		for i, resolved := range plan {
			fmt.Fprintf(cmd.ErrOrStderr(), "  %d. %s (%s): %s\n", i+1, resolved.Name, resolved.Source, formatCommand(resolved.Task.Command, resolved.Task.Args))
		}
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return nil
		}
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Fprint(cmd.ErrOrStderr(), "Run this task? [y/N] ")
			answer, readErr := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
			if readErr != nil {
				return fmt.Errorf("confirmation required; use --yes for non-interactive execution")
			}
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled")
				return nil
			}
		}
		for _, resolved := range plan {
			fmt.Fprintf(cmd.ErrOrStderr(), "Running %s...\n", resolved.Name)
			processContext := context.Background()
			cancel := func() {}
			timeout, timeoutErr := tasks.TaskTimeout(resolved.Task)
			if timeoutErr != nil {
				return fmt.Errorf("task %q: invalid timeout: %w", resolved.Name, timeoutErr)
			}
			if timeout > 0 {
				processContext, cancel = context.WithTimeout(processContext, timeout)
			}
			process := exec.CommandContext(processContext, resolved.Task.Command, resolved.Task.Args...)
			process.Dir, err = tasks.TaskDir(root, resolved.Task)
			if err != nil {
				cancel()
				return fmt.Errorf("task %q: %w", resolved.Name, err)
			}
			process.Env = os.Environ()
			for key, value := range resolved.Task.Env {
				process.Env = append(process.Env, key+"="+value)
			}
			process.Stdin = cmd.InOrStdin()
			process.Stdout = cmd.OutOrStdout()
			process.Stderr = cmd.ErrOrStderr()
			if err := process.Run(); err != nil {
				cancel()
				return fmt.Errorf("task %q failed: %w", resolved.Name, err)
			}
			cancel()
		}
		return nil
	},
}

func init() {
	RunCmd.Flags().Bool("dry-run", false, "Show the task without executing it")
	RunCmd.Flags().Bool("yes", false, "Skip confirmation")
}

func formatCommand(command string, args []string) string {
	parts := []string{strconv.Quote(command)}
	for _, arg := range args {
		parts = append(parts, strconv.Quote(arg))
	}
	return strings.Join(parts, " ")
}
