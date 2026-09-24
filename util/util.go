package util

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"fmt"
	"io"

	"github.com/spf13/viper"
)

const (
	red       = "\033[31m"
	pathColor = "\033[33m"
	green     = "\033[32m"
	quoted    = "\033[35m"
	reset     = "\033[0m"
	dirColor  = "\033[38;2;190;130;255m"
	fileColor = "\033[38;2;210;210;210m"
)

// GetHomeDir returns the home directory
func GetHomeDir() string {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDir

}

// PrintTree prints the directory tree.
func PrintTree(path string) {
	node, err := BuildTree(path, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	PrintTreeNode(node)
}

// GetCurrentDir returns the current directory
func GetCurrentDir() string {
	cwd, _ := os.Getwd()
	return cwd
}

// FormatTime formats a time
func FormatTime(t time.Time) string {
	// format yyyy-mm-dd hh:mm:ss
	return t.Format(viper.GetString("timeFormat"))

}

// GetPath returns the path to a file
func GetPath(path string) (string, error) {
	path = ResolvePath(path)

	if !CheckIfDirExists(path) {
		err := fmt.Sprintf("\n%v%v%v does not exist", red, path, reset)
		return path, errors.New(err)
	}
	return path, nil
}

// ResolvePath returns a cleaned absolute path without requiring it to exist.
func ResolvePath(path string) string {
	if path == "" {
		path = GetCurrentDir()
	}
	if path == "~" || strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		path = filepath.Join(GetHomeDir(), strings.TrimPrefix(path, "~"+string(os.PathSeparator)))
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(GetCurrentDir(), path)
	}
	return filepath.Clean(path)
}

// CheckIfDirExists checks if a directory exists
func CheckIfDirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// getCWD returns the current working directory
func getCWD() string {
	cwd, _ := os.Getwd()
	return cwd
}

// CopyFile copies a file from one location to another
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree without overwriting files.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	return copyDir(src, dst, false)
}

// CopyDirForce recursively copies a directory tree and overwrites destination files.
func CopyDirForce(src string, dst string) error {
	return copyDir(src, dst, true)
}

func copyDir(src string, dst string, overwrite bool) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	if src == dst {
		return fmt.Errorf("source and destination are the same directory")
	}
	if rel, relErr := filepath.Rel(src, dst); relErr == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("destination cannot be inside source directory")
	}

	si, err := os.Stat(src)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	if info, statErr := os.Stat(dst); statErr == nil {
		if !info.IsDir() {
			return fmt.Errorf("destination is not a directory: %s", dst)
		}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		fmt.Println(err)
		return
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		if entry.IsDir() {
			err = copyDir(srcPath, dstPath, overwrite)
			if err != nil {
				return
			}
		} else {
			if _, statErr := os.Stat(dstPath); statErr == nil && !overwrite {
				return fmt.Errorf("destination file already exists: %s", dstPath)
			}
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

// UpdateProjectName updates a project name in a file
func UpdateProjectName(path string, oldString string, newString string, ignoreCase bool) (int, error) {
	if oldString == "" {
		return 0, errors.New("old string cannot be empty")
	}

	var filesChanged []string
	count := 0
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil {
			return nil
		}
		if info.IsDir() {
			if shouldSkipDirectory(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		read, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.IndexByte(read, 0) >= 0 {
			return nil
		}

		contents := string(read)
		var newContents string
		if ignoreCase {
			re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(oldString))
			matches := re.FindAllStringIndex(contents, -1)
			if len(matches) == 0 {
				return nil
			}
			count += len(matches)
			newContents = re.ReplaceAllStringFunc(contents, func(string) string { return newString })
		} else {
			matches := strings.Count(contents, oldString)
			if matches == 0 {
				return nil
			}
			count += matches
			newContents = strings.ReplaceAll(contents, oldString, newString)
		}

		if err := atomicWrite(path, []byte(newContents), info.Mode().Perm()); err != nil {
			return err
		}
		filesChanged = append(filesChanged, path)
		return nil
	})
	if err != nil {
		return count, err
	}
	if len(filesChanged) != 0 {
		fmt.Printf("\nReplaced in:\n%v", filesReplacedString(filesChanged))
	}
	return count, nil
}

func shouldSkipDirectory(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", "node_modules", "vendor":
		return true
	default:
		return false
	}
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".stew-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
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
	return os.Rename(tmpPath, path)
}

func filesReplacedString(filesChanged []string) string {
	var str string
	for _, file := range filesChanged {
		str += fmt.Sprintf(" - %v%v%v\n", quoted, file, reset)
	}
	return str
}
