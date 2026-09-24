// Package bundle creates and extracts portable stew template archives.
package bundle

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Create(src, archivePath string) error {
	src = filepath.Clean(src)
	archivePath = filepath.Clean(archivePath)
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	if archivePath == src || isInside(src, archivePath) {
		return errors.New("archive destination cannot be inside the template")
	}
	if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(archivePath), ".stew-bundle-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	gzipWriter := gzip.NewWriter(tmp)
	tarWriter := tar.NewWriter(gzipWriter)
	walkErr := filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(os.PathSeparator)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		name := filepath.ToSlash(rel)
		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return err
		}
		header.Name = name
		if entry.IsDir() {
			header.Name += "/"
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tarWriter, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if walkErr != nil {
		tarWriter.Close()
		gzipWriter.Close()
		tmp.Close()
		return walkErr
	}
	if err := tarWriter.Close(); err != nil {
		gzipWriter.Close()
		tmp.Close()
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, archivePath)
}

func Extract(archivePath, destination string) error {
	destination = filepath.Clean(destination)
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return err
	}
	archive, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()
	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tmp, err := os.MkdirTemp(filepath.Dir(destination), ".stew-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	tarReader := tar.NewReader(gzipReader)
	seen := make(map[string]struct{})
	for {
		header, readErr := tarReader.Next()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return readErr
		}
		rel, err := safeArchivePath(header.Name)
		if err != nil {
			return err
		}
		if _, exists := seen[rel]; exists {
			return fmt.Errorf("duplicate archive entry: %s", rel)
		}
		seen[rel] = struct{}{}
		path := filepath.Join(tmp, rel)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode().Perm())
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, io.LimitReader(tarReader, header.Size))
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			return fmt.Errorf("unsupported archive entry type %q", header.Name)
		}
	}
	return os.Rename(tmp, destination)
}

func Checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

func safeArchivePath(name string) (string, error) {
	if name == "" || filepath.IsAbs(name) {
		return "", fmt.Errorf("unsafe archive path: %q", name)
	}
	rel := filepath.Clean(filepath.FromSlash(name))
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe archive path: %q", name)
	}
	return rel, nil
}

func isInside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
