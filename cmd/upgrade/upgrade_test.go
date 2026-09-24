package upgrade

import "testing"

func TestModulePath(t *testing.T) {
	if modulePath != "github.com/benjuh/stew" {
		t.Fatalf("modulePath = %q", modulePath)
	}
}
