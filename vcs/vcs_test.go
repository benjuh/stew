package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCloneFetchAndRevision(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	remote := filepath.Join(root, "remote.git")
	clonePath := filepath.Join(root, "clone")
	runGit(t, root, "init", "-b", "main", source)
	runGit(t, source, "config", "user.email", "test@example.com")
	runGit(t, source, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("one\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "add", ".")
	runGit(t, source, "commit", "-m", "initial")
	runGit(t, root, "clone", "--bare", source, remote)
	runGit(t, source, "remote", "add", "origin", remote)
	runGit(t, source, "push", "-u", "origin", "main")

	if err := Clone("file://"+remote, clonePath, "main"); err != nil {
		t.Fatal(err)
	}
	initial, err := Revision(clonePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("two\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, source, "commit", "-am", "update")
	runGit(t, source, "push", "origin", "main")
	if err := Fetch(clonePath); err != nil {
		t.Fatal(err)
	}
	remoteRevision, err := RemoteRevision(clonePath, "main")
	if err != nil || remoteRevision == initial {
		t.Fatalf("remote revision = %q, err = %v", remoteRevision, err)
	}
	diff, err := Diff(clonePath, initial, remoteRevision)
	if err != nil || diff == "" {
		t.Fatalf("diff = %q, err = %v", diff, err)
	}
	dirty, err := Dirty(clonePath)
	if err != nil || dirty {
		t.Fatalf("clean clone dirty = %v, err = %v", dirty, err)
	}
	if err := os.WriteFile(filepath.Join(clonePath, "local.txt"), []byte("local\n"), 0644); err != nil {
		t.Fatal(err)
	}
	dirty, err = Dirty(clonePath)
	if err != nil || !dirty {
		t.Fatalf("modified clone dirty = %v, err = %v", dirty, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
