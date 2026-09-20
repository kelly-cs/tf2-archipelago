package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPatchSigmodLoadoutMenu(t *testing.T) {
	root := t.TempDir()
	for _, relative := range sigmodExtensionPaths("linux") {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		arch := "x86"
		if filepath.Dir(relative) == "addons/sourcemod/extensions/x64" {
			arch = "x64"
		}
		body := append([]byte("before\x00"), sigmodLoadoutMenuBroken...)
		body = append(body, sigmodTitleBroken...)
		for _, call := range sigmodTitleCalls[arch] {
			body = append(body, call.old...)
		}
		body = append(body, []byte("\x00after")...)
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := patchSigmodLoadoutMenu(root, "linux"); err != nil {
		t.Fatal(err)
	}
	if err := patchSigmodLoadoutMenu(root, "linux"); err != nil {
		t.Fatalf("second patch should be harmless: %v", err)
	}
	if !sigmodLoadoutMenuPatched(root, "linux") {
		t.Fatal("patched extensions were not recognized")
	}
	for _, relative := range sigmodExtensionPaths("linux") {
		arch := "x86"
		if filepath.Dir(relative) == "addons/sourcemod/extensions/x64" {
			arch = "x64"
		}
		body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		want := append([]byte("before\x00"), sigmodLoadoutMenuFixed...)
		want = append(want, sigmodTitleFixed...)
		for _, call := range sigmodTitleCalls[arch] {
			want = append(want, call.fixed...)
		}
		want = append(want, []byte("\x00after")...)
		if !bytes.Equal(body, want) {
			t.Fatalf("unexpected bytes in %s", relative)
		}
	}
}

func TestPatchSigmodLoadoutMenuRefusesUnknownBinary(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, filepath.FromSlash(sigmodExtensionPaths("windows")[0]))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("unknown binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := patchSigmodLoadoutMenu(root, "windows"); err == nil {
		t.Fatal("unknown SigMod binary should require a new patch assessment")
	}
}

func TestPatchSigmodLoadoutMenuWindows(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, filepath.FromSlash(sigmodExtensionPaths("windows")[0]))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, sigmodLoadoutMenuBroken, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := patchSigmodLoadoutMenu(root, "windows"); err != nil {
		t.Fatal(err)
	}
	if !sigmodLoadoutMenuPatched(root, "windows") {
		t.Fatal("patched Windows extension was not recognized")
	}
}
