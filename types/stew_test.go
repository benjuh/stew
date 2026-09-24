package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStewsAddRejectsDuplicateNamesAndPaths(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	var stews Stews
	if err := stews.Add("app", "first", first); err != nil {
		t.Fatal(err)
	}
	if err := stews.Add("app", "second", second); err == nil {
		t.Fatal("expected duplicate name to be rejected")
	}
	if err := stews.Add("other", "same path", first); err == nil {
		t.Fatal("expected duplicate path to be rejected")
	}
	if len(stews) != 1 {
		t.Fatalf("got %d stews, want 1", len(stews))
	}
}

func TestStewsSaveLoadAndEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "stews.json")
	stews := Stews{{Name: "app", Path: t.TempDir()}}
	if err := stews.Save(path); err != nil {
		t.Fatal(err)
	}
	var loaded Stews
	if err := loaded.Load(path); err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].Name != "app" {
		t.Fatalf("loaded %+v", loaded)
	}

	empty := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(empty, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := (&Stews{}).Load(empty); err != nil {
		t.Fatalf("empty registry should load: %v", err)
	}
}
