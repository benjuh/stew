// Package inherit resolves template inheritance into a standalone template tree.
package inherit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/types"
	"gopkg.in/yaml.v3"
)

type Resolved struct {
	Root     string
	Manifest manifest.Manifest
	Cleanup  func()
}

// Resolve returns a temporary standalone tree for a catalog template.
// Call Cleanup when the tree is no longer needed.
func Resolve(registryPath, name string) (Resolved, error) {
	stews := types.Stews{}
	if err := stews.Load(registryPath); err != nil {
		return Resolved{}, err
	}
	return resolve(stews, name, make(map[string]bool))
}

func resolve(stews types.Stews, name string, visiting map[string]bool) (Resolved, error) {
	if visiting[name] {
		return Resolved{}, fmt.Errorf("template inheritance cycle detected at %q", name)
	}
	stew, err := stews.GetByName(name)
	if err != nil {
		return Resolved{}, fmt.Errorf("resolve parent template %q: %w", name, err)
	}
	metadata, err := manifest.Load(stew.Path)
	if err != nil {
		return Resolved{}, fmt.Errorf("load template %q manifest: %w", name, err)
	}
	visiting[name] = true
	defer delete(visiting, name)

	parent := Resolved{}
	if metadata.Extends != "" {
		parent, err = resolve(stews, metadata.Extends, visiting)
		if err != nil {
			return Resolved{}, err
		}
		defer parent.Cleanup()
	}

	root, err := os.MkdirTemp("", "stew-inherit-")
	if err != nil {
		return Resolved{}, err
	}
	cleanup := func() { _ = os.RemoveAll(root) }
	if parent.Root != "" {
		if err := overlay(parent.Root, root); err != nil {
			cleanup()
			return Resolved{}, err
		}
	}
	if err := overlay(stew.Path, root); err != nil {
		cleanup()
		return Resolved{}, err
	}
	merged := merge(parent.Manifest, metadata)
	merged.Extends = ""
	data, err := yaml.Marshal(merged)
	if err != nil {
		cleanup()
		return Resolved{}, err
	}
	if err := os.WriteFile(filepath.Join(root, manifest.Filename), data, 0644); err != nil {
		cleanup()
		return Resolved{}, err
	}
	return Resolved{Root: root, Manifest: merged, Cleanup: cleanup}, nil
}

func merge(parent, child manifest.Manifest) manifest.Manifest {
	result := parent
	if child.Description != "" {
		result.Description = child.Description
	}
	result.Tags = unique(append(append([]string{}, parent.Tags...), child.Tags...))
	variables := append([]manifest.Variable{}, parent.Variables...)
	for _, variable := range child.Variables {
		replaced := false
		for i := range variables {
			if variables[i].Name == variable.Name {
				variables[i] = variable
				replaced = true
				break
			}
		}
		if !replaced {
			variables = append(variables, variable)
		}
	}
	result.Variables = variables
	result.Tasks = make(map[string]manifest.Task, len(parent.Tasks)+len(child.Tasks))
	for name, task := range parent.Tasks {
		result.Tasks[name] = task
	}
	for name, task := range child.Tasks {
		result.Tasks[name] = task
	}
	return result
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func overlay(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == src || entry.Name() == manifest.Filename {
			if entry.IsDir() && path != src {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0755)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return err
		}
		return copyFile(path, destination)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
