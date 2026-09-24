package bundle

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateExtractAndChecksum(t *testing.T) {
	src := t.TempDir()
	if err := os.Mkdir(filepath.Join(src, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".git", "config"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(src, "run.sh")
	if err := os.WriteFile(file, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "template.tar.gz")
	if err := Create(src, archive); err != nil {
		t.Fatal(err)
	}
	checksum, err := Checksum(archive)
	if err != nil || len(checksum) != len("sha256:")+64 {
		t.Fatalf("checksum = %q, err = %v", checksum, err)
	}
	destination := filepath.Join(t.TempDir(), "template")
	if err := Extract(archive, destination); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(destination, ".git")); !os.IsNotExist(err) {
		t.Fatalf(".git was included: %v", err)
	}
	info, err := os.Stat(filepath.Join(destination, "run.sh"))
	if err != nil || info.Mode().Perm() != 0755 {
		t.Fatalf("extracted file mode = %v, err = %v", info.Mode().Perm(), err)
	}
}

func TestExtractRejectsTraversal(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "bad.tar.gz")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "../../outside", Mode: 0644, Size: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Extract(archive, filepath.Join(t.TempDir(), "destination")); err == nil {
		t.Fatal("expected traversal archive to be rejected")
	}
}
