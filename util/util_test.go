package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPath(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	got, err := GetPath(".")
	expected, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err = filepath.EvalSymlinks(got)
	if err != nil || got != expected {
		t.Fatalf("GetPath(.) = %q, %v; want %q, nil", got, err, expected)
	}
}

func TestCopyDirProtectsExistingFilesAndForceOverwrites(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "README.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "README.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := CopyDir(src, dst); err == nil {
		t.Fatal("CopyDir should reject an existing destination file")
	}
	if err := CopyDirForce(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "README.md"))
	if err != nil || string(got) != "new" {
		t.Fatalf("forced copy produced %q, %v", got, err)
	}
}

func TestUpdateProjectName(t *testing.T) {
	dir := t.TempDir()
	textPath := filepath.Join(dir, "README.md")
	binPath := filepath.Join(dir, "image.bin")
	if err := os.WriteFile(textPath, []byte("Project PROJECT project"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte{0, 'P', 'r', 'o', 'j'}, 0600); err != nil {
		t.Fatal(err)
	}

	count, err := UpdateProjectName(dir, "project", "app", true)
	if err != nil || count != 3 {
		t.Fatalf("UpdateProjectName count = %d, %v; want 3, nil", count, err)
	}
	got, err := os.ReadFile(textPath)
	if err != nil || string(got) != "app app app" {
		t.Fatalf("text file = %q, %v", got, err)
	}
	got, err = os.ReadFile(binPath)
	if err != nil || string(got) != string([]byte{0, 'P', 'r', 'o', 'j'}) {
		t.Fatalf("binary file was modified: %v", got)
	}
}
