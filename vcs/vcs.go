// Package vcs contains the small Git integration used by stew sources.
package vcs

import (
	"fmt"
	"os/exec"
	"strings"
)

func IsGitSource(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "git@") || strings.HasPrefix(source, "ssh://") || strings.HasPrefix(source, "file://")
}

func Clone(source, destination, ref string) error {
	args := []string{"clone"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, source, destination)
	if _, err := run(args...); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}

func Fetch(repo string) error {
	if _, err := run("-C", repo, "fetch", "--quiet", "--prune", "origin"); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}
	return nil
}

func Revision(repo string) (string, error) {
	output, err := run("-C", repo, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("read Git revision failed: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func Ref(repo string) (string, error) {
	output, err := run("-C", repo, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("read Git branch failed: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func RemoteRevision(repo, ref string) (string, error) {
	if ref == "" {
		var err error
		ref, err = Ref(repo)
		if err != nil {
			return "", err
		}
	}
	for _, candidate := range []string{"refs/remotes/origin/" + ref, ref} {
		if output, err := run("-C", repo, "rev-parse", candidate); err == nil {
			return strings.TrimSpace(output), nil
		}
	}
	return "", fmt.Errorf("remote revision not found for ref %q", ref)
}

func Dirty(repo string) (bool, error) {
	output, err := run("-C", repo, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("check Git status failed: %w", err)
	}
	return strings.TrimSpace(output) != "", nil
}

func FastForward(repo, target string) error {
	if _, err := run("-C", repo, "merge", "--ff-only", target); err != nil {
		return fmt.Errorf("git fast-forward failed: %w", err)
	}
	return nil
}

func Reset(repo, revision string) error {
	if _, err := run("-C", repo, "reset", "--hard", revision); err != nil {
		return fmt.Errorf("git rollback failed: %w", err)
	}
	return nil
}

func Diff(repo, from, to string) (string, error) {
	output, err := run("-C", repo, "diff", "--no-ext-diff", from, to)
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return output, nil
}

func run(args ...string) (string, error) {
	command := exec.Command("git", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			return "", fmt.Errorf("%w: %s", err, message)
		}
		return "", err
	}
	return string(output), nil
}
