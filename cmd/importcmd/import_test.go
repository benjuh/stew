package importcmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExcluded(t *testing.T) {
	excludes := []string{".git", "node_modules", ".env"}
	for _, path := range []string{".git/config", "node_modules/react/index.js", ".env.local"} {
		if !excluded(path, false, excludes) {
			t.Errorf("excluded(%q) = false", path)
		}
	}
	if excluded("src/App.tsx", false, excludes) {
		t.Fatal("source file was incorrectly excluded")
	}
}

func TestSelectProfileAuto(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{"react":"18.0.0"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	profile, err := selectProfile(root, "auto")
	if err != nil || profile.Name != "react" {
		t.Fatalf("profile = %+v, err = %v", profile, err)
	}
}

func TestCuratedFilesSkipsExcludedDirectories(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "node_modules", "pkg"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "node_modules", "pkg", "index.js"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("kept"), 0644); err != nil {
		t.Fatal(err)
	}
	files, err := curatedFiles(root, []string{"node_modules"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "README.md" {
		t.Fatalf("files = %v", files)
	}
}
