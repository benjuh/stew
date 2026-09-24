package upgrade

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

const modulePath = "github.com/benjuh/stew"

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade [version]",
	Short: "Upgrade Stew to the latest or a specific version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		check, _ := cmd.Flags().GetBool("check")
		if check {
			if len(args) > 0 {
				return fmt.Errorf("a version cannot be supplied with --check")
			}
			return checkLatest(cmd)
		}

		target := "latest"
		if len(args) == 1 {
			target = args[0]
		}
		return install(cmd, target)
	},
}

func install(cmd *cobra.Command, target string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("cannot upgrade automatically: Go is not installed or is not on PATH; install Stew through your package manager instead")
	}
	version := target
	if target == "latest" {
		var err error
		version, err = latestVersion(cmd)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "Upgrading Stew to %s...\n", version)
	command := exec.CommandContext(cmd.Context(), "go", "install", "-ldflags", "-X github.com/benjuh/stew/cmd.Version="+version, modulePath+"@"+version)
	command.Env = directGoEnv()
	command.Stdout = cmd.OutOrStdout()
	command.Stderr = cmd.ErrOrStderr()
	if err := command.Run(); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Stew upgrade complete.")
	return nil
}

func checkLatest(cmd *cobra.Command) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("cannot check for upgrades automatically: Go is not installed or is not on PATH")
	}

	version, err := latestVersion(cmd)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Latest version: %s\nRun `stew upgrade` to install it.\n", version)
	return nil
}

func latestVersion(cmd *cobra.Command) (string, error) {
	command := exec.CommandContext(cmd.Context(), "go", "list", "-m", "-f", "{{.Version}}", modulePath+"@latest")
	command.Env = directGoEnv()
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("could not check the latest version: %w", err)
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("could not determine the latest version")
	}
	return version, nil
}

func directGoEnv() []string {
	env := os.Environ()
	for i, value := range env {
		if strings.HasPrefix(value, "GOPROXY=") {
			env[i] = "GOPROXY=direct"
			return env
		}
	}
	return append(env, "GOPROXY=direct")
}

func init() {
	UpgradeCmd.Flags().Bool("check", false, "Show the latest available version without upgrading")
}
