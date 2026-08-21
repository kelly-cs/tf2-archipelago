package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateLogFileKeepsTheRunBefore(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, LogFileName), []byte("the run that broke\n"), 0o644); err != nil {
		t.Fatalf("cannot seed the log: %v", err)
	}

	file, err := CreateLogFile(root)
	if err != nil {
		t.Fatalf("CreateLogFile: %v", err)
	}
	defer func() { _ = file.Close() }()

	kept, err := os.ReadFile(filepath.Join(root, LogPreviousName))
	if err != nil {
		t.Fatalf("the previous run was not kept: %v", err)
	}
	if string(kept) != "the run that broke\n" {
		t.Errorf("the previous run holds %q", kept)
	}
	// The new run starts empty, rather than appending to the old one.
	current, err := os.ReadFile(filepath.Join(root, LogFileName))
	if err != nil || len(current) != 0 {
		t.Errorf("this run's log is %q, %v", current, err)
	}
}

// The game server appends to its console log across runs, so it is moved out
// of the way before the server opens it.
func TestRotateConsoleLog(t *testing.T) {
	gameDir := t.TempDir()
	tf := filepath.Join(gameDir, "tf")
	if err := os.MkdirAll(tf, 0o755); err != nil {
		t.Fatalf("cannot create %s: %v", tf, err)
	}
	if err := os.WriteFile(filepath.Join(tf, ConsoleLogName), []byte("last evening\n"), 0o644); err != nil {
		t.Fatalf("cannot seed the console log: %v", err)
	}

	RotateConsoleLog(gameDir)

	kept, err := os.ReadFile(filepath.Join(tf, ConsolePreviousName))
	if err != nil {
		t.Fatalf("last evening was not kept: %v", err)
	}
	if string(kept) != "last evening\n" {
		t.Errorf("the kept log holds %q", kept)
	}
	if _, err := os.Stat(filepath.Join(tf, ConsoleLogName)); !os.IsNotExist(err) {
		t.Errorf("the console log was left in place: %v", err)
	}
}

// A game directory with no console log in it yet, which is every first run.
func TestRotateConsoleLogWithNothingToKeep(t *testing.T) {
	RotateConsoleLog(t.TempDir())
}

// A first run has nothing to keep, and an install root that is not there yet
// is the normal case: the log is opened before anything is downloaded.
func TestCreateLogFileOnAFirstRun(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not-yet")
	file, err := CreateLogFile(root)
	if err != nil {
		t.Fatalf("CreateLogFile: %v", err)
	}
	_ = file.Close()
	if _, err := os.Stat(filepath.Join(root, LogPreviousName)); !os.IsNotExist(err) {
		t.Errorf("a first run left a previous log: %v", err)
	}
}
