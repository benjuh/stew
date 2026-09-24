package upgrade

import (
	"strings"
	"testing"
)

func TestModulePath(t *testing.T) {
	if modulePath != "github.com/benjuh/stew" {
		t.Fatalf("modulePath = %q", modulePath)
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		version, goos, goarch, want string
	}{
		{"v1.6.0", "darwin", "arm64", "stew_1.6.0_darwin_arm64.tar.gz"},
		{"1.6.0", "linux", "amd64", "stew_1.6.0_linux_amd64.tar.gz"},
		{"v1.6.0", "windows", "amd64", "stew_1.6.0_windows_amd64.zip"},
	}
	for _, test := range tests {
		if got := assetName(test.version, test.goos, test.goarch); got != test.want {
			t.Errorf("assetName(%q, %q, %q) = %q, want %q", test.version, test.goos, test.goarch, got, test.want)
		}
	}
}

func TestChecksumFor(t *testing.T) {
	checksum := strings.Repeat("a", 64)
	got, err := checksumFor([]byte(checksum+"  stew_1.6.0_linux_amd64.tar.gz\n"), "stew_1.6.0_linux_amd64.tar.gz")
	if err != nil || got != checksum {
		t.Fatalf("checksumFor() = %q, %v", got, err)
	}
}

func TestIsPackageManagerPath(t *testing.T) {
	if !isPackageManagerPath("/opt/homebrew/Cellar/stew/1.6.0/bin/stew") {
		t.Fatal("Homebrew path was not detected")
	}
	if isPackageManagerPath("/Users/benjamin/go/bin/stew") {
		t.Fatal("Go path was incorrectly detected as package-managed")
	}
}
