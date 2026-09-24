package remove

import (
	"path/filepath"
	"testing"
)

func TestManagedTemplatePath(t *testing.T) {
	root := t.TempDir()
	managed := filepath.Join(root, "templates", "starter")

	if path, ok := managedTemplatePath(filepath.Join(root, "templates"), managed); !ok || path != managed {
		t.Fatalf("managed path = %q, %v; want %q, true", path, ok, managed)
	}
	if path, ok := managedTemplatePath(filepath.Join(root, "templates"), filepath.Join(root, "templates")); ok || path != "" {
		t.Fatalf("cache root = %q, %v; want empty, false", path, ok)
	}
	if path, ok := managedTemplatePath(filepath.Join(root, "templates"), filepath.Join(root, "other")); ok || path != "" {
		t.Fatalf("outside path = %q, %v; want empty, false", path, ok)
	}
}
