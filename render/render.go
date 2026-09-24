// Package render generates a project from a stew directory.
package render

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Options controls project generation.
type Options struct {
	DryRun    bool
	Overwrite bool
	Variables map[string]string
}

// Change describes a file operation planned or performed by Generate.
type Change struct {
	Action      string `json:"action"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// Generate renders src into dst and returns the file operations performed.
// Text files and paths support Go template syntax. Binary files are copied
// without rendering. Existing files are rejected unless Overwrite is true.
func Generate(src, dst string, options Options) ([]Change, error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	if src == dst {
		return nil, errors.New("source and destination are the same directory")
	}
	if rel, err := filepath.Rel(src, dst); err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, errors.New("destination cannot be inside source directory")
	}
	info, err := os.Stat(src)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source is not a directory: %s", src)
	}

	patterns, err := readIgnorePatterns(filepath.Join(src, ".stewignore"))
	if err != nil {
		return nil, err
	}
	data := templateData(options.Variables)
	changes := make([]Change, 0)
	var pending []pendingFile
	seenDestinations := make(map[string]struct{})
	var generationErr error

	err = filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == src {
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
		if entry.IsDir() {
			if defaultIgnoredDirectory(entry.Name()) || ignored(rel, true, patterns) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == ".stewignore" || ignored(rel, false, patterns) {
			return nil
		}

		renderedRel, err := renderPath(rel, data)
		if err != nil {
			generationErr = fmt.Errorf("render path %q: %w", rel, err)
			return generationErr
		}
		destination, err := safeDestination(dst, renderedRel)
		if err != nil {
			generationErr = err
			return err
		}
		if _, seen := seenDestinations[destination]; seen {
			return fmt.Errorf("multiple template files render to the same destination: %s", destination)
		}
		seenDestinations[destination] = struct{}{}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		mode := entry.Type().Perm()
		if mode == 0 {
			if fileInfo, statErr := entry.Info(); statErr == nil {
				mode = fileInfo.Mode().Perm()
			}
		}
		if bytes.IndexByte(content, 0) < 0 {
			content, err = renderContent(path, content, data)
			if err != nil {
				generationErr = err
				return err
			}
		}
		if _, statErr := os.Stat(destination); statErr == nil {
			if !options.Overwrite {
				generationErr = fmt.Errorf("destination file already exists: %s", destination)
				return generationErr
			}
			changes = append(changes, Change{Action: "overwrite", Source: path, Destination: destination})
		} else if !os.IsNotExist(statErr) {
			return statErr
		} else {
			changes = append(changes, Change{Action: "create", Source: path, Destination: destination})
		}
		pending = append(pending, pendingFile{path: destination, content: content, mode: mode})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if generationErr != nil {
		return nil, generationErr
	}
	if options.DryRun {
		return changes, nil
	}
	for _, file := range pending {
		if err := writeFile(file.path, file.content, file.mode); err != nil {
			return nil, err
		}
	}
	return changes, nil
}

// Validate parses and executes all renderable template files using placeholder
// values. It does not modify the source directory.
func Validate(src string, variableNames []string) error {
	src = filepath.Clean(src)
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	patterns, err := readIgnorePatterns(filepath.Join(src, ".stewignore"))
	if err != nil {
		return err
	}
	values := make(map[string]string, len(variableNames))
	for _, name := range variableNames {
		values[name] = "placeholder"
	}
	data := templateData(values)
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == src {
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
		if entry.IsDir() {
			if defaultIgnoredDirectory(entry.Name()) || ignored(rel, true, patterns) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Name() == ".stewignore" || entry.Name() == ".stew.yaml" || ignored(rel, false, patterns) {
			return nil
		}
		if _, err := renderPath(rel, data); err != nil {
			return fmt.Errorf("validate path %q: %w", rel, err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.IndexByte(content, 0) < 0 {
			if _, err := renderContent(path, content, data); err != nil {
				return err
			}
		}
		return nil
	})
}

type pendingFile struct {
	path    string
	content []byte
	mode    os.FileMode
}

var foreignTemplateExpression = regexp.MustCompile(`(?s)(\$\{\{.*?\}\}|\{\{[<%].*?[>%]\}\})`)

func renderContent(path string, content []byte, data map[string]string) ([]byte, error) {
	if !bytes.Contains(content, []byte("{{")) {
		return content, nil
	}
	protected, expressions := protectForeignExpressions(string(content))
	t, err := template.New(filepath.Base(path)).Option("missingkey=error").Parse(protected)
	if err != nil {
		return nil, fmt.Errorf("parse template %q: %w", path, err)
	}
	var output bytes.Buffer
	if err := t.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("render template %q: %w", path, err)
	}
	return []byte(restoreForeignExpressions(output.String(), expressions)), nil
}

func protectForeignExpressions(content string) (string, map[string]string) {
	expressions := make(map[string]string)
	index := 0
	protected := foreignTemplateExpression.ReplaceAllStringFunc(content, func(expression string) string {
		marker := fmt.Sprintf("__STEW_FOREIGN_TEMPLATE_%d__", index)
		index++
		expressions[marker] = expression
		return marker
	})
	return protected, expressions
}

func restoreForeignExpressions(content string, expressions map[string]string) string {
	for marker, expression := range expressions {
		content = strings.ReplaceAll(content, marker, expression)
	}
	return content
}

func renderPath(path string, data map[string]string) (string, error) {
	if !strings.Contains(path, "{{") {
		return filepath.Clean(path), nil
	}
	t, err := template.New("path").Option("missingkey=error").Parse(filepath.ToSlash(path))
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := t.Execute(&output, data); err != nil {
		return "", err
	}
	return filepath.FromSlash(filepath.Clean(output.String())), nil
}

func safeDestination(root, relative string) (string, error) {
	if filepath.IsAbs(relative) {
		return "", fmt.Errorf("rendered path must be relative: %s", relative)
	}
	destination := filepath.Join(root, relative)
	rel, err := filepath.Rel(root, destination)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("rendered path escapes destination: %s", relative)
	}
	return destination, nil
}

func templateData(values map[string]string) map[string]string {
	data := make(map[string]string, len(values)*2)
	for key, value := range values {
		data[key] = value
		parts := strings.FieldsFunc(key, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
		if len(parts) > 0 {
			var builder strings.Builder
			for _, part := range parts {
				if part == "" {
					continue
				}
				runes := []rune(part)
				builder.WriteString(strings.ToUpper(string(runes[0])))
				builder.WriteString(string(runes[1:]))
			}
			data[builder.String()] = value
		}
	}
	return data
}

func readIgnorePatterns(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var patterns []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, filepath.ToSlash(line))
		}
	}
	return patterns, nil
}

func ignored(path string, directory bool, patterns []string) bool {
	path = filepath.ToSlash(path)
	for _, pattern := range patterns {
		directoryPattern := strings.HasSuffix(pattern, "/")
		pattern = strings.TrimSuffix(pattern, "/")
		pattern = strings.TrimPrefix(pattern, "/")
		if directoryPattern && !directory {
			continue
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

func defaultIgnoredDirectory(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", "node_modules", "vendor":
		return true
	default:
		return false
	}
}

func writeFile(path string, content []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".stew-render-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
