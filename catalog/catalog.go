package catalog

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/benjuh/stew/types"
)

type Index struct {
	Version   int        `yaml:"version" json:"version"`
	Templates []Template `yaml:"templates" json:"templates"`
}

type Template struct {
	ID          string    `yaml:"id" json:"id"`
	Name        string    `yaml:"name" json:"name"`
	Description string    `yaml:"description,omitempty" json:"description,omitempty"`
	Family      string    `yaml:"family,omitempty" json:"family,omitempty"`
	Language    string    `yaml:"language,omitempty" json:"language,omitempty"`
	Tags        []string  `yaml:"tags,omitempty" json:"tags,omitempty"`
	Variants    []Variant `yaml:"variants,omitempty" json:"variants,omitempty"`
}

type Variant struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Source      string `yaml:"source" json:"source"`
	Path        string `yaml:"path,omitempty" json:"path,omitempty"`
	Ref         string `yaml:"ref,omitempty" json:"ref,omitempty"`
}

func CachePath(root, name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", err
	}
	return filepath.Join(root, name), nil
}

func ValidateName(name string) error {
	if strings.TrimSpace(name) == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return fmt.Errorf("invalid template name: %q", name)
	}
	return nil
}

func Register(registryPath string, entry types.Stew) error {
	if err := ValidateName(entry.Name); err != nil {
		return err
	}
	stews := types.Stews{}
	if err := stews.Load(registryPath); err != nil {
		return err
	}
	if _, err := stews.GetByName(entry.Name); err == nil {
		return fmt.Errorf("stew name already exists: %s", entry.Name)
	}
	for _, existing := range stews {
		if existing.Path == entry.Path {
			return fmt.Errorf("stew path already exists: %s", entry.Path)
		}
	}
	stews = append(stews, entry)
	return stews.Save(registryPath)
}

func NameFromArchive(path string) string {
	name := filepath.Base(path)
	for _, suffix := range []string{".tar.gz", ".tgz", ".tar"} {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix)
		}
	}
	return name
}

func NameFromGitURL(url string) string {
	name := strings.TrimSuffix(filepath.Base(strings.TrimSuffix(url, "/")), ".git")
	return name
}
