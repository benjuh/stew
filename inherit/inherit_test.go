package inherit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benjuh/stew/manifest"
	"github.com/benjuh/stew/types"
)

func TestResolveMergesParentAndChild(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(root, "child")
	if err := os.MkdirAll(parent, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(parent, manifest.Filename), "description: Base\ntags: [base]\nvariables:\n  - name: base\n    default: yes\ntasks:\n  base:\n    command: echo\n")
	write(t, filepath.Join(parent, "README.md"), "base\n")
	write(t, filepath.Join(parent, "base.txt"), "parent\n")
	write(t, filepath.Join(child, manifest.Filename), "extends: base\ndescription: Child\ntags: [child, base]\nvariables:\n  - name: child\n    required: true\ntasks:\n  child:\n    command: echo\n")
	write(t, filepath.Join(child, "README.md"), "child\n")
	registry := filepath.Join(root, "stews.json")
	stews := types.Stews{{Name: "base", Path: parent}, {Name: "child", Path: child}}
	if err := stews.Save(registry); err != nil {
		t.Fatal(err)
	}

	resolved, err := Resolve(registry, "child")
	if err != nil {
		t.Fatal(err)
	}
	defer resolved.Cleanup()
	if resolved.Manifest.Extends != "" || resolved.Manifest.Description != "Child" || len(resolved.Manifest.Tags) != 2 {
		t.Fatalf("manifest = %+v", resolved.Manifest)
	}
	if len(resolved.Manifest.Variables) != 2 || len(resolved.Manifest.Tasks) != 2 {
		t.Fatalf("merged manifest = %+v", resolved.Manifest)
	}
	readme, err := os.ReadFile(filepath.Join(resolved.Root, "README.md"))
	if err != nil || string(readme) != "child\n" {
		t.Fatalf("README = %q, err = %v", readme, err)
	}
	if _, err := os.Stat(filepath.Join(resolved.Root, "base.txt")); err != nil {
		t.Fatalf("inherited file: %v", err)
	}
}

func TestResolveRejectsCycles(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	if err := os.MkdirAll(a, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(b, 0755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(a, manifest.Filename), "extends: b\n")
	write(t, filepath.Join(b, manifest.Filename), "extends: a\n")
	registry := filepath.Join(root, "stews.json")
	stews := types.Stews{{Name: "a", Path: a}, {Name: "b", Path: b}}
	if err := stews.Save(registry); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(registry, "a"); err == nil {
		t.Fatal("expected inheritance cycle error")
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
