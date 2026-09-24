package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSearchAndFindVariant(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.yaml")
	content := "version: 1\ntemplates:\n  - id: react-neon-render\n    name: React + Neon + Render\n    language: typescript\n    tags: [react, neon, render]\n    variants:\n      - id: minimal\n        description: Barebones starter\n        source: https://example.com/react-minimal.git\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	index, err := LoadRemote(path)
	if err != nil {
		t.Fatal(err)
	}
	results := Search(index, "neon")
	if len(results) != 1 || results[0].ID != "react-neon-render" {
		t.Fatalf("results = %+v", results)
	}
	template, err := Find(index, "react-neon-render")
	if err != nil {
		t.Fatal(err)
	}
	variant, err := template.FindVariant("minimal")
	if err != nil || variant.Source == "" {
		t.Fatalf("variant = %+v, err = %v", variant, err)
	}
}
