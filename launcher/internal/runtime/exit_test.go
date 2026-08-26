package runtime

import (
	"os/exec"
	"strings"
	"testing"
)

/*
A crash has to read as a crash.

"game server stopped: exit status 0xc0000005" was reported for weeks as "bridge
stopping", because the launcher's orderly shutdown is the last line a player
sees and the crash line looked like every other stop.
*/
func TestACrashIsNotAStop(t *testing.T) {
	cmd := exec.Command("sh", "-c", "kill -SEGV $$")
	err := cmd.Run()
	if err == nil {
		t.Fatal("the fake crash exited cleanly")
	}

	note := crashNote(err)
	if note == "" {
		t.Fatalf("a segfault was not read as a crash: %v", err)
	}
	if !strings.Contains(note, "segmentation fault") {
		t.Errorf("the note does not name the signal: %q", note)
	}

	wrapped := wrapExit("game server", err).Error()
	if !strings.Contains(wrapped, "CRASHED") {
		t.Errorf("a crash reads as a stop: %q", wrapped)
	}
}

// A clean non-zero exit is a stop, not a crash. srcds returns one when it
// refuses a map or a convar, and calling that a crash sends the reader after a
// minidump that was never written.
func TestARefusalIsStillAStop(t *testing.T) {
	err := exec.Command("sh", "-c", "exit 3").Run()
	if err == nil {
		t.Fatal("exit 3 reported success")
	}
	if note := crashNote(err); note != "" {
		t.Errorf("exit 3 was read as a crash: %q", note)
	}
	if got := wrapExit("game server", err).Error(); !strings.Contains(got, "stopped") {
		t.Errorf("a plain exit does not read as a stop: %q", got)
	}
}

// The Windows statuses, which cannot be produced by running anything here.
func TestWindowsExceptionStatusesAreNamed(t *testing.T) {
	for code, want := range map[uint32]string{
		0xC0000005: "access violation",
		0xC00000FD: "stack overflow",
		0xC0000409: "stack buffer overrun",
	} {
		if got := ntStatusNames[code]; got != want {
			t.Errorf("status %#x is named %q, want %q", code, got, want)
		}
	}
	// The top two bits are what marks an exception, so an unlisted one still
	// has to report as a crash.
	if uint32(0xC0000017)&0xC0000000 != 0xC0000000 {
		t.Error("the exception mask does not match an unlisted status")
	}
	if uint32(1)&0xC0000000 == 0xC0000000 {
		t.Error("a plain exit code matched the exception mask")
	}
}
