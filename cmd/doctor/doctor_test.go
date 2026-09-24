package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiagnoseValidProject(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".stew.yaml"), []byte("description: demo\ntasks:\n  check:\n    command: go\n    args: [version]\n    timeout: 30s\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Demo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, diagnostic := range diagnose(root) {
		if diagnostic.Level == "ERROR" {
			t.Fatalf("unexpected diagnostic: %+v", diagnostic)
		}
	}
}

func TestDiagnoseReportsInvalidTask(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".stew.yaml"), []byte("tasks:\n  broken:\n    command: definitely-not-installed\n    timeout: nope\n"), 0644); err != nil {
		t.Fatal(err)
	}
	diagnostics := diagnose(root)
	for _, diagnostic := range diagnostics {
		if diagnostic.Level == "ERROR" {
			return
		}
	}
	if len(diagnostics) == 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
}
