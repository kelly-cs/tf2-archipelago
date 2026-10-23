package webapi

import (
	"os"
	"path/filepath"
	"testing"

	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
)

func TestAttachedLogFollowsCompleteLinesWithoutRepeatingThem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "console.log")
	if err := os.WriteFile(path, []byte("already here\nbeing written"), 0o600); err != nil {
		t.Fatal(err)
	}

	var lines []apruntime.Line
	offset, initialized := readAttachedLog(path, "srcds", 0, false, func(line apruntime.Line) {
		lines = append(lines, line)
	})
	if len(lines) != 1 || lines[0].Source != "srcds" || lines[0].Text != "already here" {
		t.Fatalf("first read = %+v", lines)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(" now\na new line\n"); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	offset, initialized = readAttachedLog(path, "srcds", offset, initialized, func(line apruntime.Line) {
		lines = append(lines, line)
	})
	if !initialized || offset == 0 {
		t.Fatalf("reader state = offset %d, initialized %v", offset, initialized)
	}
	if len(lines) != 3 || lines[1].Text != "being written now" || lines[2].Text != "a new line" {
		t.Fatalf("follow-up read = %+v", lines)
	}

	_, _ = readAttachedLog(path, "srcds", offset, initialized, func(line apruntime.Line) {
		lines = append(lines, line)
	})
	if len(lines) != 3 {
		t.Fatalf("an unchanged file repeated lines: %+v", lines)
	}
}
