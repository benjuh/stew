// Package tasks resolves named project tasks from manifests and project type.
package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/benjuh/stew/manifest"
)

type Resolved struct {
	Name        string        `json:"name"`
	Task        manifest.Task `json:"task"`
	Source      string        `json:"source"`
	Description string        `json:"description,omitempty"`
}

func FindProjectRoot(start string) (string, error) {
	path, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		path = filepath.Dir(path)
	}
	for {
		if fileExists(filepath.Join(path, manifest.Filename)) || hasProjectMarker(path) {
			return path, nil
		}
		parent := filepath.Dir(path)
		if parent == path {
			return "", fmt.Errorf("could not find a project root from %s", start)
		}
		path = parent
	}
}

func Resolve(root, name string) (Resolved, error) {
	metadata, err := manifest.Load(root)
	if err != nil {
		return Resolved{}, err
	}
	if err := validateManifestTasks(metadata); err != nil {
		return Resolved{}, err
	}
	if canonical, task, ok, err := manifestTask(metadata, name); ok {
		if err != nil {
			return Resolved{}, err
		}
		if matches(root, task.Detect) {
			return resolvedManifestTask(canonical, task), nil
		}
		return Resolved{}, fmt.Errorf("task %q is not applicable to this project", name)
	}
	for _, candidate := range builtIns(root) {
		if candidate.Name == name {
			return candidate, nil
		}
	}
	return Resolved{}, fmt.Errorf("task %q was not found", name)
}

// Plan resolves a task and all of its dependencies in execution order.
func Plan(root, name string) ([]Resolved, error) {
	metadata, err := manifest.Load(root)
	if err != nil {
		return nil, err
	}
	if err := validateManifestTasks(metadata); err != nil {
		return nil, err
	}
	result := make([]Resolved, 0)
	visiting := make(map[string]bool)
	visited := make(map[string]bool)
	var visit func(string) error
	visit = func(requested string) error {
		canonical, task, ok, taskErr := manifestTask(metadata, requested)
		if taskErr != nil {
			return taskErr
		}
		if !ok {
			resolved, err := Resolve(root, requested)
			if err != nil {
				return err
			}
			canonical = resolved.Name
			if visited[canonical] {
				return nil
			}
			result = append(result, resolved)
			visited[canonical] = true
			return nil
		}
		if visiting[canonical] {
			return fmt.Errorf("task dependency cycle detected at %q", canonical)
		}
		if visited[canonical] {
			return nil
		}
		if !matches(root, task.Detect) {
			return fmt.Errorf("task %q is not applicable to this project", requested)
		}
		visiting[canonical] = true
		for _, dependency := range task.DependsOn {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		delete(visiting, canonical)
		result = append(result, resolvedManifestTask(canonical, task))
		visited[canonical] = true
		return nil
	}
	if err := visit(name); err != nil {
		return nil, err
	}
	return result, nil
}

func Available(root string) ([]Resolved, error) {
	metadata, err := manifest.Load(root)
	if err != nil {
		return nil, err
	}
	if err := validateManifestTasks(metadata); err != nil {
		return nil, err
	}
	result := make([]Resolved, 0, len(metadata.Tasks)+4)
	seen := make(map[string]struct{})
	for name, task := range metadata.Tasks {
		if err := validateTask(name, task); err != nil {
			return nil, err
		}
		if matches(root, task.Detect) {
			result = append(result, resolvedManifestTask(name, task))
			seen[name] = struct{}{}
		}
	}
	for _, candidate := range builtIns(root) {
		if _, exists := seen[candidate.Name]; !exists {
			result = append(result, candidate)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// ValidateProject returns diagnostics for project structure, manifests, and tasks.
func ValidateProject(root string) []string {
	diagnostics := make([]string, 0)
	metadata, err := manifest.Load(root)
	if err != nil {
		return append(diagnostics, fmt.Sprintf("manifest: %v", err))
	}
	if err := metadata.Validate(); err != nil {
		diagnostics = append(diagnostics, fmt.Sprintf("manifest: %v", err))
	}
	if err := validateManifestTasks(metadata); err != nil {
		diagnostics = append(diagnostics, fmt.Sprintf("tasks: %v", err))
	}
	for name, task := range metadata.Tasks {
		if task.Dir != "" {
			if err := validateTaskDir(root, task.Dir); err != nil {
				diagnostics = append(diagnostics, fmt.Sprintf("task %q: %v", name, err))
			}
		}
		if task.Timeout != "" {
			if _, err := time.ParseDuration(task.Timeout); err != nil {
				diagnostics = append(diagnostics, fmt.Sprintf("task %q: invalid timeout %q", name, task.Timeout))
			}
		}
		for key := range task.Env {
			if strings.TrimSpace(key) == "" || strings.ContainsRune(key, '=') {
				diagnostics = append(diagnostics, fmt.Sprintf("task %q: invalid environment variable name", name))
			}
		}
		if _, err := executableFor(root, task); err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("task %q: %v", name, err))
		}
		if _, err := Plan(root, name); err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("task %q: %v", name, err))
		}
	}
	if available, err := Available(root); err == nil {
		for _, task := range available {
			if task.Source == "manifest" {
				continue
			}
			if _, err := executableFor(root, task.Task); err != nil {
				diagnostics = append(diagnostics, fmt.Sprintf("task %q: %v", task.Name, err))
			}
		}
	}
	return diagnostics
}

func builtIns(root string) []Resolved {
	result := make([]Resolved, 0)
	if fileExists(filepath.Join(root, "go.mod")) {
		result = append(result,
			Resolved{Name: "fmt", Source: "go", Task: manifest.Task{Command: "go", Args: []string{"fmt", "./..."}}},
			Resolved{Name: "test", Source: "go", Task: manifest.Task{Command: "go", Args: []string{"test", "./..."}}},
			Resolved{Name: "lint", Source: "go", Task: manifest.Task{Command: "go", Args: []string{"vet", "./..."}}},
			Resolved{Name: "build", Source: "go", Task: manifest.Task{Command: "go", Args: []string{"build", "./..."}}},
		)
	}
	if fileExists(filepath.Join(root, "Cargo.toml")) {
		result = append(result,
			Resolved{Name: "fmt", Source: "rust", Task: manifest.Task{Command: "cargo", Args: []string{"fmt"}}},
			Resolved{Name: "test", Source: "rust", Task: manifest.Task{Command: "cargo", Args: []string{"test"}}},
		)
	}
	if packageScripts, ok := nodeScripts(filepath.Join(root, "package.json")); ok {
		for _, name := range []string{"fmt", "format", "test", "lint", "build"} {
			if _, exists := packageScripts[name]; exists {
				result = append(result, Resolved{Name: name, Source: "node", Task: manifest.Task{Command: "npm", Args: []string{"run", name}}})
			}
		}
	}
	return result
}

func manifestTask(metadata manifest.Manifest, requested string) (string, manifest.Task, bool, error) {
	if task, ok := metadata.Tasks[requested]; ok {
		if err := validateTask(requested, task); err != nil {
			return "", manifest.Task{}, true, err
		}
		return requested, task, true, nil
	}
	for name, task := range metadata.Tasks {
		for _, alias := range task.Aliases {
			if alias == requested {
				if err := validateTask(name, task); err != nil {
					return "", manifest.Task{}, true, err
				}
				return name, task, true, nil
			}
		}
	}
	return "", manifest.Task{}, false, nil
}

func validateManifestTasks(metadata manifest.Manifest) error {
	seen := make(map[string]string, len(metadata.Tasks)*2)
	for name, task := range metadata.Tasks {
		if err := validateTask(name, task); err != nil {
			return err
		}
		if previous, exists := seen[name]; exists {
			return fmt.Errorf("task name %q is declared more than once (by %q and %q)", name, previous, name)
		}
		seen[name] = name
		for _, alias := range task.Aliases {
			if previous, exists := seen[alias]; exists {
				return fmt.Errorf("task alias %q for %q conflicts with %q", alias, name, previous)
			}
			seen[alias] = name
		}
	}
	return nil
}

func resolvedManifestTask(name string, task manifest.Task) Resolved {
	if platform, ok := task.Platforms[runtime.GOOS]; ok && platform.Command != "" {
		task.Command = platform.Command
		task.Args = platform.Args
	}
	return Resolved{Name: name, Task: task, Source: "manifest", Description: task.Description}
}

func validateTaskDir(root, dir string) error {
	if filepath.IsAbs(dir) {
		return fmt.Errorf("working directory must be relative to the project root")
	}
	path := filepath.Join(root, dir)
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("working directory escapes the project root")
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("working directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("working directory %q is not a directory", dir)
	}
	return nil
}

// TaskDir returns the validated working directory for a task.
func TaskDir(root string, task manifest.Task) (string, error) {
	if task.Dir == "" {
		return root, nil
	}
	if err := validateTaskDir(root, task.Dir); err != nil {
		return "", err
	}
	return filepath.Join(root, task.Dir), nil
}

func executableFor(root string, task manifest.Task) (string, error) {
	resolved := resolvedManifestTask("", task)
	if resolved.Task.Command == "" {
		return "", fmt.Errorf("no command for current platform")
	}
	if filepath.IsAbs(resolved.Task.Command) || strings.ContainsRune(resolved.Task.Command, filepath.Separator) {
		path := resolved.Task.Command
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("executable %q is unavailable", resolved.Task.Command)
		}
		return path, nil
	}
	if _, err := exec.LookPath(resolved.Task.Command); err != nil {
		return "", fmt.Errorf("executable %q is unavailable", resolved.Task.Command)
	}
	return resolved.Task.Command, nil
}

// TaskTimeout parses a task timeout such as "30s" or "5m".
func TaskTimeout(task manifest.Task) (time.Duration, error) {
	if task.Timeout == "" {
		return 0, nil
	}
	return time.ParseDuration(task.Timeout)
}

func nodeScripts(path string) (map[string]string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var packageJSON struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &packageJSON) != nil {
		return nil, false
	}
	return packageJSON.Scripts, true
}

func validateTask(name string, task manifest.Task) error {
	if task.Command == "" && task.Platforms[runtime.GOOS].Command == "" {
		return fmt.Errorf("task %q has no command", name)
	}
	for _, alias := range task.Aliases {
		if strings.TrimSpace(alias) == "" {
			return fmt.Errorf("task %q has an empty alias", name)
		}
	}
	for key := range task.Env {
		if strings.TrimSpace(key) == "" || strings.ContainsRune(key, '=') {
			return fmt.Errorf("task %q has an invalid environment variable name", name)
		}
	}
	return nil
}

func matches(root string, markers []string) bool {
	if len(markers) == 0 {
		return true
	}
	for _, marker := range markers {
		if fileExists(filepath.Join(root, marker)) {
			return true
		}
	}
	return false
}

func hasProjectMarker(root string) bool {
	for _, marker := range []string{"go.mod", "Cargo.toml", "package.json", "pyproject.toml"} {
		if fileExists(filepath.Join(root, marker)) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
