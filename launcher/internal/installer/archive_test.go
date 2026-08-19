package installer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

// tarball builds a .tar.gz in memory from the entries a real mod drop holds: a
// directory, a file that has to stay executable, and a symlink.
func tarball(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	writer := tar.NewWriter(gz)

	entries := []struct {
		header tar.Header
		body   string
	}{
		{tar.Header{Name: "addons/", Typeflag: tar.TypeDir, Mode: 0o755}, ""},
		{tar.Header{Name: "steamcmd.sh", Typeflag: tar.TypeReg, Mode: 0o755, Size: 5}, "hello"},
		{tar.Header{Name: "addons/notes.txt", Typeflag: tar.TypeReg, Mode: 0o644, Size: 2}, "hi"},
		{tar.Header{Name: "linux32/steamclient.so", Typeflag: tar.TypeSymlink, Linkname: "../steamcmd.sh"}, ""},
	}
	for _, entry := range entries {
		header := entry.header
		if err := writer.WriteHeader(&header); err != nil {
			t.Fatal(err)
		}
		if entry.body != "" {
			if _, err := writer.Write([]byte(entry.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

// Valve serves SteamCMD as a zip on Windows and a tarball on Linux, and
// AlliedModders does the same with its two drops. unpackTo reads the bytes
// rather than the file name, so one caller handles both.
func TestUnpackToReadsATarball(t *testing.T) {
	dir := t.TempDir()
	if err := unpackTo(tarball(t), dir); err != nil {
		t.Fatalf("unpackTo: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(dir, "addons", "notes.txt"))
	if err != nil || string(body) != "hi" {
		t.Errorf("notes.txt = %q, %v", body, err)
	}

	// steamcmd.sh and srcds_run are useless without the execute bit, and a
	// tar reader that drops the mode is the way to lose it.
	info, err := os.Stat(filepath.Join(dir, "steamcmd.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("steamcmd.sh came out %v, which cannot run", info.Mode().Perm())
	}

	// The SteamCMD tarball ships steamclient.so as a symlink; following it
	// into a copy would be a second copy of a large library.
	link, err := os.Readlink(filepath.Join(dir, "linux32", "steamclient.so"))
	if err != nil || link != "../steamcmd.sh" {
		t.Errorf("symlink = %q, %v", link, err)
	}
}

// Valve's SteamCMD tarball carries no entry for linux32/ and every one of its
// binaries under it. An unpacker that waits for a directory entry before
// making the directory loses all of them, which is what this caught.
func TestUnpackToMakesMissingDirectories(t *testing.T) {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	writer := tar.NewWriter(gz)
	header := tar.Header{Name: "linux32/steamcmd", Typeflag: tar.TypeReg, Mode: 0o755, Size: 2}
	if err := writer.WriteHeader(&header); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	_ = writer.Close()
	_ = gz.Close()

	dir := t.TempDir()
	if err := unpackTo(buffer.Bytes(), dir); err != nil {
		t.Fatalf("unpackTo: %v", err)
	}
	if !exists(filepath.Join(dir, "linux32", "steamcmd")) {
		t.Error("a file whose directory had no entry of its own was not written")
	}
}

// A second install over the first is the normal case: Ensure runs on every
// start, and Repair exists to run it again on purpose.
func TestUnpackToIsRepeatable(t *testing.T) {
	dir := t.TempDir()
	for range 2 {
		if err := unpackTo(tarball(t), dir); err != nil {
			t.Fatalf("unpackTo: %v", err)
		}
	}
	if _, err := os.Readlink(filepath.Join(dir, "linux32", "steamclient.so")); err != nil {
		t.Errorf("the symlink did not survive a second unpack: %v", err)
	}
}

// An archive that writes outside the directory it was told to is the one worth
// stopping. Nothing upstream does it; that is the point of checking.
func TestUnpackToRefusesAnEscape(t *testing.T) {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	writer := tar.NewWriter(gz)
	header := tar.Header{Name: "../escaped.txt", Typeflag: tar.TypeReg, Mode: 0o644, Size: 1}
	if err := writer.WriteHeader(&header); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("x")); err != nil {
		t.Fatal(err)
	}
	_ = writer.Close()
	_ = gz.Close()

	dir := t.TempDir()
	// filepath.Clean("/../escaped.txt") is "/escaped.txt", so the entry lands
	// inside rather than above: either way nothing is written outside dir.
	if err := unpackTo(buffer.Bytes(), dir); err != nil {
		t.Fatalf("unpackTo: %v", err)
	}
	if exists(filepath.Join(filepath.Dir(dir), "escaped.txt")) {
		t.Error("an archive entry wrote outside the install directory")
	}
}
