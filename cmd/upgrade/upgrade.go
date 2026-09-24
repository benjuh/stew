package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	modulePath = "github.com/benjuh/stew"
	releaseAPI = "https://api.github.com/repos/benjuh/stew/releases/latest"
	releaseURL = "https://github.com/benjuh/stew/releases/download"
)

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
	version, err := resolveVersion(cmd.Context(), target)
	if err != nil {
		return err
	}

	executable, executableErr := os.Executable()
	if executableErr == nil && isGoInstallation(cmd.Context(), executable) {
		return installWithGo(cmd, version)
	}
	return installBinary(cmd, version, executable, executableErr)
}

func installWithGo(cmd *cobra.Command, version string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return installBinary(cmd, version, "", fmt.Errorf("Go is not installed or is not on PATH"))
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Upgrading Stew to %s with go install...\n", version)
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

func installBinary(cmd *cobra.Command, version, executable string, executableErr error) error {
	if executableErr != nil || executable == "" {
		return fmt.Errorf("cannot locate the Stew executable; install the release archive manually")
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("automatic binary upgrades are not supported on Windows while Stew is running; download the %s release from %s/v%s", version, releaseURL, strings.TrimPrefix(version, "v"))
	}
	if isPackageManagerPath(executable) {
		return fmt.Errorf("Stew appears to be managed by a package manager at %s; upgrade it through that package manager", executable)
	}

	asset := assetName(version, runtime.GOOS, runtime.GOARCH)
	archiveURL := fmt.Sprintf("%s/v%s/%s", releaseURL, strings.TrimPrefix(version, "v"), asset)
	checksumURL := fmt.Sprintf("%s/v%s/checksums.txt", releaseURL, strings.TrimPrefix(version, "v"))
	fmt.Fprintf(cmd.ErrOrStderr(), "Downloading Stew %s for %s/%s...\n", version, runtime.GOOS, runtime.GOARCH)
	archiveData, err := download(cmd.Context(), archiveURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	checksums, err := download(cmd.Context(), checksumURL)
	if err != nil {
		return fmt.Errorf("checksum download failed: %w", err)
	}
	expected, err := checksumFor(checksums, asset)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(archiveData)
	if !strings.EqualFold(expected, hex.EncodeToString(digest[:])) {
		return fmt.Errorf("checksum verification failed for %s", asset)
	}
	binary, err := extractTarGz(archiveData, executableName(executable))
	if err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}
	if err := replaceExecutable(executable, binary); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Stew upgrade complete.")
	return nil
}

func checkLatest(cmd *cobra.Command) error {
	version, err := latestVersion(cmd.Context())
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Latest version: %s\nRun `stew upgrade` to install it.\n", version)
	return nil
}

func resolveVersion(ctx context.Context, target string) (string, error) {
	if target == "latest" {
		return latestVersion(ctx)
	}
	if !strings.HasPrefix(target, "v") {
		target = "v" + target
	}
	return target, nil
}

func latestVersion(ctx context.Context) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseAPI, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "stew-upgrade")
	response, err := (&http.Client{Timeout: 15 * time.Second}).Do(request)
	if err != nil {
		return "", fmt.Errorf("could not check the latest version: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("could not check the latest version: GitHub returned %s", response.Status)
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("could not parse the latest version: %w", err)
	}
	if release.TagName == "" {
		return "", fmt.Errorf("GitHub did not return a latest release")
	}
	return release.TagName, nil
}

func download(ctx context.Context, source string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "stew-upgrade")
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("request returned %s", response.Status)
	}
	return io.ReadAll(response.Body)
}

func assetName(version, goos, goarch string) string {
	extension := "tar.gz"
	if goos == "windows" {
		extension = "zip"
	}
	return fmt.Sprintf("stew_%s_%s_%s.%s", strings.TrimPrefix(version, "v"), goos, goarch, extension)
}

func checksumFor(data []byte, filename string) (string, error) {
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum for %s was not found", filename)
}

func extractTarGz(data []byte, filename string) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if filepath.Base(header.Name) == filename {
			return io.ReadAll(tarReader)
		}
	}
	return nil, fmt.Errorf("%s was not found in archive", filename)
}

func replaceExecutable(path string, binary []byte) error {
	mode := os.FileMode(0755)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode()
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".stew-upgrade-*")
	if err != nil {
		return fmt.Errorf("could not prepare executable replacement: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(binary); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("could not replace executable: %w", err)
	}
	return nil
}

func executableName(path string) string {
	name := filepath.Base(path)
	if runtime.GOOS == "windows" && !strings.HasSuffix(name, ".exe") {
		return name + ".exe"
	}
	return name
}

func isPackageManagerPath(path string) bool {
	clean := filepath.ToSlash(path)
	return strings.Contains(clean, "/Cellar/") || strings.Contains(clean, "/Homebrew/")
}

func isGoInstallation(ctx context.Context, executable string) bool {
	if _, err := exec.LookPath("go"); err != nil {
		return false
	}
	output, err := exec.CommandContext(ctx, "go", "env", "GOBIN", "GOPATH").Output()
	if err != nil {
		return false
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return false
	}
	goBin := strings.TrimSpace(lines[0])
	if goBin == "" {
		goBin = filepath.Join(strings.TrimSpace(lines[1]), "bin")
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		resolved = executable
	}
	expected, err := filepath.EvalSymlinks(filepath.Join(goBin, executableName(executable)))
	if err != nil {
		expected = filepath.Join(goBin, executableName(executable))
	}
	return filepath.Clean(resolved) == filepath.Clean(expected)
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
