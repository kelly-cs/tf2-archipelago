//go:build !windows

package runtime

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// srcds_run is a shell script that runs the game server as a child. Killing
// the script alone leaves that child holding the game port and the rcon port,
// and the next Start binds neither.
func TestStopTakesDownTheServerTheScriptStarted(t *testing.T) {
	root := t.TempDir()
	s := fakeServer(t, root)
	pidFile := filepath.Join(root, "child.pid")
	script := "#!/bin/sh\nsleep 600 &\necho $! > " + pidFile + "\necho fake server up\nwait\n"
	if err := os.WriteFile(filepath.Join(root, "tf-dedicated", "srcds_run"), []byte(script), 0o755); err != nil {
		t.Fatalf("cannot write the fake server: %v", err)
	}

	spoke := make(chan struct{})
	var once sync.Once
	sup := NewSupervisor(s, nil, func(line Line) {
		if line.Source == "srcds" && line.Text == "fake server up" {
			once.Do(func() { close(spoke) })
		}
	})
	if err := sup.Start(func(error) {}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer sup.Stop()
	select {
	case <-spoke:
	case <-time.After(10 * time.Second):
		t.Fatal("the fake server never spoke")
	}

	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("the fake server wrote no pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("the pid file holds %q: %v", raw, err)
	}

	sup.Stop()
	if alive(pid) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
		t.Fatalf("process %d outlived the Stop", pid)
	}
}

// alive reports whether the process is still there, with a moment's grace for
// the signal to land.
func alive(pid int) bool {
	for range 20 {
		if err := syscall.Kill(pid, 0); err != nil {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	return true
}
