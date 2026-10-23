package webapi

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
)

func TestMismatchedImportClearsPreviousImportedMarker(t *testing.T) {
	folder := t.TempDir()
	name := "archive-assets.zip"
	if err := os.WriteFile(filepath.Join(folder, name+".imported"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	writer := zip.NewWriter(&body)
	entry, err := writer.Create("tf/download/maps/example.bsp")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("map")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = importCommunityArchive(folder, name, bytes.NewReader(body.Bytes()))
	if _, ok := errors.AsType[*installer.CommunityArchiveHashMismatchError](err); !ok {
		t.Fatalf("import error = %v, want hash mismatch", err)
	}
	if _, err := os.Stat(filepath.Join(folder, name+".imported")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale import marker remains: %v", err)
	}
	if _, err := os.Stat(filepath.Join(folder, name+".hash-mismatch")); err != nil {
		t.Fatalf("held ZIP missing: %v", err)
	}
}
