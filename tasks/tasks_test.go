package tasks

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/benjuh/stew/manifest"
)

func TestResolveManifestTaskAndFindProjectRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".stew.yaml"), []byte("tasks:\n  check:\n    command: echo\n    args: [ok]\n    detect: [go.mod]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "internal", "pkg")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	found, err := FindProjectRoot(nested)
	if err != nil || found != root {
		t.Fatalf("root = %q, err = %v", found, err)
	}
	resolved, err := Resolve(root, "check")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Source != "manifest" || resolved.Task.Command != "echo" || len(resolved.Task.Args) != 1 {
		t.Fatalf("resolved = %+v", resolved)
	}
}

func TestResolveManifestTaskHonorsDetection(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".stew.yaml"), []byte("tasks:\n  check:\n    command: echo\n    detect: [go.mod]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(root, "check"); err == nil {
		t.Fatal("expected detection mismatch")
	}
}

func TestAvailableIncludesGoBuiltIns(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	available, err := Available(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(available) != 4 {
		t.Fatalf("available = %+v", available)
	}
	for _, task := range available {
		if task.Source != "go" {
			t.Fatalf("task = %+v", task)
		}
	}
}

func TestPlanResolvesAliasesDependenciesAndDeduplicates(t *testing.T) {
	root := t.TempDir()
	content := `tasks:
  fmt:
    command: echo
    args: [fmt]
    aliases: [format]
  test:
    command: echo
    args: [test]
    depends_on: [format]
  check:
    command: echo
    args: [check]
    depends_on: [fmt, test]
`
	if err := os.WriteFile(filepath.Join(root, ".stew.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	plan, err := Plan(root, "check")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != 3 || plan[0].Name != "fmt" || plan[1].Name != "test" || plan[2].Name != "check" {
		t.Fatalf("plan = %+v", plan)
	}
}

func TestPlanRejectsDependencyCycles(t *testing.T) {
	root := t.TempDir()
	content := "tasks:\n  one:\n    command: echo\n    depends_on: [two]\n  two:\n    command: echo\n    depends_on: [one]\n"
	if err := os.WriteFile(filepath.Join(root, ".stew.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Plan(root, "one"); err == nil {
		t.Fatal("expected dependency cycle error")
	}
}

func TestResolveUsesPlatformOverride(t *testing.T) {
	root := t.TempDir()
	content := "tasks:\n  check:\n    command: base\n    args: [base]\n    platforms:\n      " + runtime.GOOS + ":\n        command: native\n        args: [platform]\n"
	if err := os.WriteFile(filepath.Join(root, ".stew.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolve(root, "check")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Task.Command != "native" || len(resolved.Task.Args) != 1 || resolved.Task.Args[0] != "platform" {
		t.Fatalf("resolved = %+v", resolved)
	}
}

func TestTaskRuntimeSettings(t *testing.T) {
	root := t.TempDir()
	working := filepath.Join(root, "backend")
	if err := os.Mkdir(working, 0755); err != nil {
		t.Fatal(err)
	}
	task := manifest.Task{Dir: "backend", Timeout: "5m", Env: map[string]string{"MODE": "test"}}
	gotDir, err := TaskDir(root, task)
	if err != nil || gotDir != working {
		t.Fatalf("dir = %q, err = %v", gotDir, err)
	}
	timeout, err := TaskTimeout(task)
	if err != nil || timeout != 5*time.Minute {
		t.Fatalf("timeout = %v, err = %v", timeout, err)
	}
	if _, err := TaskDir(root, manifest.Task{Dir: "../outside"}); err == nil {
		t.Fatal("expected working directory escape error")
	}
}
