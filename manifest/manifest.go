// Package manifest loads optional metadata stored alongside a template.
package manifest

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const Filename = ".stew.yaml"

type Variable struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

type Task struct {
	Command     string                  `yaml:"command" json:"command"`
	Args        []string                `yaml:"args,omitempty" json:"args,omitempty"`
	Detect      []string                `yaml:"detect,omitempty" json:"detect,omitempty"`
	Description string                  `yaml:"description,omitempty" json:"description,omitempty"`
	Aliases     []string                `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	DependsOn   []string                `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Platforms   map[string]PlatformTask `yaml:"platforms,omitempty" json:"platforms,omitempty"`
	Env         map[string]string       `yaml:"env,omitempty" json:"env,omitempty"`
	Dir         string                  `yaml:"dir,omitempty" json:"dir,omitempty"`
	Timeout     string                  `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

type PlatformTask struct {
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
}

type Manifest struct {
	Extends     string          `yaml:"extends,omitempty" json:"extends,omitempty"`
	Description string          `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string        `yaml:"tags,omitempty" json:"tags,omitempty"`
	Variables   []Variable      `yaml:"variables,omitempty" json:"variables,omitempty"`
	Tasks       map[string]Task `yaml:"tasks,omitempty" json:"tasks,omitempty"`
}

func Load(templatePath string) (Manifest, error) {
	path := filepath.Join(templatePath, Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Manifest{}, nil
		}
		return Manifest{}, err
	}
	var result Manifest
	if err := yaml.Unmarshal(data, &result); err != nil {
		return Manifest{}, err
	}
	return result, nil
}

func (m Manifest) Validate() error {
	seen := make(map[string]struct{}, len(m.Variables))
	for _, variable := range m.Variables {
		name := strings.TrimSpace(variable.Name)
		if name == "" {
			return errors.New("manifest variable name cannot be empty")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate manifest variable: %s", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

// Resolve combines values in increasing order of precedence: environment,
// manifest defaults, values file, CLI values, and finally interactive input.
func (m Manifest) Resolve(cli map[string]string, valuesPath string, input io.Reader, output io.Writer, nonInteractive bool) (map[string]string, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	values := make(map[string]string)
	for _, variable := range m.Variables {
		if value, ok := os.LookupEnv(variable.Name); ok {
			values[variable.Name] = value
		} else if value, ok := os.LookupEnv(strings.ToUpper(variable.Name)); ok {
			values[variable.Name] = value
		} else if variable.Default != "" {
			values[variable.Name] = variable.Default
		}
	}
	if valuesPath != "" {
		fileValues, err := loadValuesFile(valuesPath)
		if err != nil {
			return nil, err
		}
		for key, value := range fileValues {
			values[key] = value
		}
	}
	for key, value := range cli {
		values[key] = value
	}

	reader := bufio.NewReader(input)
	for _, variable := range m.Variables {
		value, exists := values[variable.Name]
		if exists && value != "" {
			continue
		}
		if !variable.Required {
			if !exists {
				values[variable.Name] = ""
			}
			continue
		}
		if nonInteractive {
			return nil, fmt.Errorf("required variable %q is missing", variable.Name)
		}
		if output == nil {
			return nil, fmt.Errorf("cannot prompt for required variable %q", variable.Name)
		}
		if variable.Description != "" {
			fmt.Fprintf(output, "%s (%s): ", variable.Name, variable.Description)
		} else {
			fmt.Fprintf(output, "%s: ", variable.Name)
		}
		answer, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			return nil, fmt.Errorf("required variable %q is missing", variable.Name)
		}
		values[variable.Name] = answer
	}
	for key, value := range cli {
		values[key] = value
	}
	return values, nil
}

func loadValuesFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse values file: %w", err)
	}
	values := make(map[string]string, len(raw))
	for key, value := range raw {
		if value == nil {
			values[key] = ""
		} else {
			values[key] = fmt.Sprint(value)
		}
	}
	return values, nil
}
