package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildTreeSortsAndIgnores(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".stewignore"), []byte("ignored.txt\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	tree, err := BuildTree(dir, 0)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Name == "" || len(tree.Children) != 1 || tree.Children[0].Name != "src" {
		t.Fatalf("tree = %+v", tree)
	}
	limited, err := BuildTree(dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(limited.Children) != 1 || len(limited.Children[0].Children) != 0 {
		t.Fatalf("limited tree = %+v", limited)
	}
}
