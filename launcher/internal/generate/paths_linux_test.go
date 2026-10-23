//go:build linux

package generate

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func fakeAppImage(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
if [ "$1" != Generate ] || [ "$2" != -- ]; then
  echo "launcher did not dispatch Generate: $*" >&2
  exit 2
fi
shift 2
out=""
while [ $# -gt 0 ]; do
  case "$1" in --outputpath) out="$2"; shift;; esac
  shift
done
: > "$out/AP_AppImage.zip"
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestLinuxFindsAndRunsOfficialAppImageShape(t *testing.T) {
	home, data := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "")
	t.Setenv("XDG_DATA_HOME", data)
	appImage := filepath.Join(home, "Applications", "Archipelago_0.6.7_linux-x86_64.AppImage")
	fakeAppImage(t, appImage)

	if paths := SearchPath(""); !slices.Contains(paths, appImage) {
		t.Fatalf("search path does not contain %s: %v", appImage, paths)
	}
	found, err := FindApp("")
	if err != nil || found != appImage {
		t.Fatalf("FindApp() = %q, %v; want %q", found, err, appImage)
	}

	s := settings.Defaults()
	s.InstallRoot = t.TempDir()
	result, err := Run(context.Background(), Options{
		Settings: s,
		Apworld:  []byte("PK-fake"),
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(result.Archive) != "AP_AppImage.zip" {
		t.Fatalf("archive = %s", result.Archive)
	}
	world := filepath.Join(data, "Archipelago", "worlds", "tf2_mvm.apworld")
	if _, err := os.Stat(world); err != nil {
		t.Fatalf("apworld was not installed into AppImage user data: %v", err)
	}
}

func TestLinuxFindsAnExtractedTarballInLocalOpt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "")
	appDir := filepath.Join(home, ".local", "opt", "Archipelago")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "ArchipelagoGenerate"), []byte("generator"), 0o755); err != nil {
		t.Fatal(err)
	}
	found, err := FindApp("")
	if err != nil || found != appDir {
		t.Fatalf("FindApp() = %q, %v; want %q", found, err, appDir)
	}
}

func TestLinuxFindsGeneratorOnPath(t *testing.T) {
	home, bin := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)
	if err := os.WriteFile(filepath.Join(bin, "ArchipelagoGenerate"), []byte("generator"), 0o755); err != nil {
		t.Fatal(err)
	}
	paths := SearchPath("")
	if len(paths) == 0 || paths[0] != bin {
		t.Fatalf("SearchPath() = %v; PATH generator directory should be first", paths)
	}
}

func TestLinuxIgnoresAnAppImageWithoutExecutePermission(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "")
	path := filepath.Join(home, "Downloads", "Archipelago_0.6.7_linux-x86_64.AppImage")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not executable"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range SearchPath("") {
		if strings.EqualFold(candidate, path) {
			t.Fatalf("non-executable AppImage appeared in search path: %v", candidate)
		}
	}
	if _, err := FindApp(path); err == nil {
		t.Fatal("an explicitly selected non-executable AppImage was accepted")
	}
}
