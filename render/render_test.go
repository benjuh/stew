package render

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateRendersContentAndPaths(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "project")
	if err := os.Mkdir(filepath.Join(src, "{{ .ProjectName }}"), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(src, "{{ .ProjectName }}", "README.md")
	if err := os.WriteFile(path, []byte("# {{ .ProjectName }}\nmodule {{ .module }}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	changes, err := Generate(src, dst, Options{Variables: map[string]string{
		"project_name": "demo",
		"module":       "example.com/demo",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 1 || changes[0].Action != "create" {
		t.Fatalf("changes = %+v", changes)
	}
	got, err := os.ReadFile(filepath.Join(dst, "demo", "README.md"))
	if err != nil || string(got) != "# demo\nmodule example.com/demo\n" {
		t.Fatalf("rendered content = %q, %v", got, err)
	}
}

func TestGenerateCopiesManifest(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "project")
	manifest := []byte("description: Demo\ntasks:\n  test:\n    command: go\n    args: [test, ./...]\n")
	if err := os.WriteFile(filepath.Join(src, ".stew.yaml"), manifest, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(src, dst, Options{}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, ".stew.yaml"))
	if err != nil || string(got) != string(manifest) {
		t.Fatalf("manifest = %q, %v", got, err)
	}
}

func TestGenerateDryRunDoesNotWriteAndHonorsIgnore(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "project")
	if err := os.WriteFile(filepath.Join(src, ".stewignore"), []byte("ignored.txt\nignored-dir/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "ignored.txt"), []byte("ignore"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(src, "ignored-dir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "ignored-dir", "file.txt"), []byte("ignore"), 0644); err != nil {
		t.Fatal(err)
	}

	changes, err := Generate(src, dst, Options{DryRun: true})
	if err != nil || len(changes) != 1 {
		t.Fatalf("changes = %+v, err = %v", changes, err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("dry-run created destination: %v", err)
	}
}

func TestGenerateRejectsMissingVariableAndExistingFile(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("{{ .Name }}"), 0644); err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	if _, err := Generate(src, dst, Options{}); err == nil {
		t.Fatal("expected missing variable error")
	}
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "README.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(src, dst, Options{}); err == nil {
		t.Fatal("expected existing file conflict")
	}
	if _, err := Generate(src, dst, Options{Overwrite: true}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateUsesManifestVariableNames(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("# {{ .project_name }}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Validate(src, []string{"project_name"}); err != nil {
		t.Fatal(err)
	}
	if err := Validate(src, nil); err == nil {
		t.Fatal("expected undeclared variable validation error")
	}
}

func TestGeneratePreservesForeignTemplateExpressions(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "project")
	content := "name: app\nrun: ${{ matrix.node-version }}\nshortcode: {{< render-image >}}\ntrimmed: {{-< render-image >}}\nowner: {{ .Owner }}\n"
	if err := os.WriteFile(filepath.Join(src, "workflow.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(src, dst, Options{Variables: map[string]string{"Owner": "benjuh"}}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "workflow.yml"))
	if err != nil {
		t.Fatal(err)
	}
	want := "name: app\nrun: ${{ matrix.node-version }}\nshortcode: {{< render-image >}}\ntrimmed: {{-< render-image >}}\nowner: benjuh\n"
	if string(got) != want {
		t.Fatalf("workflow = %q, want %q", got, want)
	}
}

func TestHasForeignTemplateSyntax(t *testing.T) {
	for _, content := range []string{"{{< shortcode >}}", "{{-< shortcode >}}", "{{% block %}}", "${{ matrix.os }}"} {
		if !hasForeignTemplateSyntax(content) {
			t.Errorf("hasForeignTemplateSyntax(%q) = false", content)
		}
	}
	if hasForeignTemplateSyntax("{{ .ProjectName }}") {
		t.Fatal("Stew syntax was detected as foreign syntax")
	}
}
