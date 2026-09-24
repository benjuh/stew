package catalog

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestLoadCachedDoesNotPersistEmptyIndexes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("version: 1\ntemplates: []\n"))
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	if _, err := LoadCached(server.URL, cacheDir); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("cache entries = %d, want 0", len(entries))
	}
}

func TestLoadCachedRefreshesPastFreshCache(t *testing.T) {
	content := "version: 1\ntemplates:\n  - id: first\n    name: First\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	if _, err := LoadCached(server.URL, cacheDir); err != nil {
		t.Fatal(err)
	}
	content = "version: 1\ntemplates:\n  - id: second\n    name: Second\n"

	index, err := LoadCached(server.URL, cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(index.Templates) != 1 || index.Templates[0].ID != "first" {
		t.Fatalf("cached index = %+v", index)
	}
	index, err = LoadCachedWithOptions(server.URL, cacheDir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(index.Templates) != 1 || index.Templates[0].ID != "second" {
		t.Fatalf("refreshed index = %+v", index)
	}
	if strings.Contains(index.Templates[0].ID, "first") {
		t.Fatal("refresh returned stale template")
	}
}
