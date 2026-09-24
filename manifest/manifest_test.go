package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	content := "description: React starter\ntags: [react, frontend]\ntasks:\n  fmt:\n    command: go\n    args: [fmt, ./...]\n    aliases: [format]\n    depends_on: [prepare]\n    description: Format the project\nvariables:\n  - name: project_name\n    required: true\n"
	if err := os.WriteFile(filepath.Join(dir, Filename), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Description != "React starter" || len(result.Tags) != 2 || len(result.Variables) != 1 || !result.Variables[0].Required || result.Tasks["fmt"].Command != "go" || result.Tasks["fmt"].Aliases[0] != "format" || result.Tasks["fmt"].DependsOn[0] != "prepare" {
		t.Fatalf("manifest = %+v", result)
	}
}

func TestLoadMissingIsEmpty(t *testing.T) {
	result, err := Load(t.TempDir())
	if err != nil || result.Description != "" || len(result.Tags) != 0 {
		t.Fatalf("manifest = %+v, err = %v", result, err)
	}
}

func TestResolveUsesPrecedenceAndPromptsRequiredValues(t *testing.T) {
	valuesPath := filepath.Join(t.TempDir(), "values.yaml")
	if err := os.WriteFile(valuesPath, []byte("author: Values Author\n"), 0644); err != nil {
		t.Fatal(err)
	}
	m := Manifest{Variables: []Variable{
		{Name: "project_name", Required: true},
		{Name: "author", Default: "Default Author"},
	}}
	output := &strings.Builder{}
	values, err := m.Resolve(map[string]string{"author": "CLI Author"}, valuesPath, strings.NewReader("demo\n"), output, false)
	if err != nil {
		t.Fatal(err)
	}
	if values["project_name"] != "demo" || values["author"] != "CLI Author" {
		t.Fatalf("values = %+v", values)
	}
	if !strings.Contains(output.String(), "project_name") {
		t.Fatalf("prompt = %q", output.String())
	}
}

func TestResolveNonInteractiveRequiresValue(t *testing.T) {
	m := Manifest{Variables: []Variable{{Name: "project_name", Required: true}}}
	if _, err := m.Resolve(nil, "", strings.NewReader(""), &strings.Builder{}, true); err == nil {
		t.Fatal("expected missing required variable error")
	}
}
