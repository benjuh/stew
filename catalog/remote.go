package catalog

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadRemote(source string) (Index, error) {
	var data []byte
	var err error
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		response, requestErr := http.Get(source)
		if requestErr != nil {
			return Index{}, requestErr
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return Index{}, fmt.Errorf("catalog request returned %s", response.Status)
		}
		data, err = io.ReadAll(response.Body)
	} else {
		data, err = os.ReadFile(source)
	}
	if err != nil {
		return Index{}, err
	}
	var index Index
	if err := yaml.Unmarshal(data, &index); err != nil {
		return Index{}, fmt.Errorf("parse catalog: %w", err)
	}
	if index.Version == 0 {
		index.Version = 1
	}
	for i := range index.Templates {
		if index.Templates[i].ID == "" {
			return Index{}, fmt.Errorf("catalog template at index %d has no id", i)
		}
	}
	return index, nil
}

// LoadCached loads a catalog and keeps a short-lived local fallback for offline use.
func LoadCached(source, cacheDir string) (Index, error) {
	return LoadCachedWithOptions(source, cacheDir, false)
}

// LoadCachedWithOptions loads a catalog, optionally bypassing the fresh cache.
// Empty indexes are returned but never written to the cache.
func LoadCachedWithOptions(source, cacheDir string, refresh bool) (Index, error) {
	if !strings.HasPrefix(source, "http://") && !strings.HasPrefix(source, "https://") {
		return LoadRemote(source)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(source)))[:16]
	cachePath := filepath.Join(cacheDir, ".catalog-"+hash+".yaml")
	if !refresh {
		if info, err := os.Stat(cachePath); err == nil && time.Since(info.ModTime()) < 24*time.Hour {
			return LoadRemote(cachePath)
		}
	}
	index, err := LoadRemote(source)
	if err == nil {
		if len(index.Templates) == 0 {
			_ = os.Remove(cachePath)
			return index, nil
		}
		if mkdirErr := os.MkdirAll(cacheDir, 0755); mkdirErr == nil {
			if data, marshalErr := yaml.Marshal(index); marshalErr == nil {
				_ = os.WriteFile(cachePath, data, 0644)
			}
		}
		return index, nil
	}
	if _, cacheErr := os.Stat(cachePath); cacheErr == nil {
		return LoadRemote(cachePath)
	}
	return Index{}, err
}

func Search(index Index, query string) []Template {
	query = strings.ToLower(strings.TrimSpace(query))
	result := make([]Template, 0)
	for _, template := range index.Templates {
		if query == "" || containsTemplate(template, query) {
			result = append(result, template)
		}
	}
	return result
}

func Find(index Index, id string) (Template, error) {
	for _, template := range index.Templates {
		if template.ID == id {
			return template, nil
		}
	}
	return Template{}, fmt.Errorf("catalog template %q was not found", id)
}

func (template Template) FindVariant(name string) (Variant, error) {
	for _, variant := range template.Variants {
		if variant.ID == name || variant.Name == name {
			return variant, nil
		}
	}
	return Variant{}, fmt.Errorf("template %q has no variant %q", template.ID, name)
}

func containsTemplate(template Template, query string) bool {
	values := []string{template.ID, template.Name, template.Description, template.Family, template.Language}
	values = append(values, template.Tags...)
	for _, variant := range template.Variants {
		values = append(values, variant.ID, variant.Name, variant.Description)
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}
