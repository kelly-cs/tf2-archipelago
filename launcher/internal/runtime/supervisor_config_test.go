package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readServerCfg returns what the supervisor rendered into the game tree.
func readServerCfg(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "tf-dedicated", "tf", "cfg", "server.cfg")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no server.cfg at %s: %v", path, err)
	}
	return string(raw)
}

/*
A start renders the configs from the settings that start is using.

The game server reads server.cfg once, at its own startup, so a setting the
launcher holds but has not written is a setting the server never sees. The
launcher used to render the file once per run, before either interface was
reached, which left every later edit on disk and nowhere else: a class unticked
in the terminal interface was saved, shown as unticked, and still drawn by the
mod for the rest of the evening.
*/
func TestStartRendersTheConfigsFromTheCurrentSettings(t *testing.T) {
	root := t.TempDir()
	s := fakeServer(t, root)
	s.SrcdsBotClassBlacklist = []string{"spy"}

	sup := NewSupervisor(s, nil, func(Line) {})
	if err := sup.Start(func(error) {}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	sup.Stop()

	first := readServerCfg(t, root)
	if !strings.Contains(first, `sm_redbots_manager_class_blacklist "spy"`) {
		t.Fatalf("the first start did not write the blacklist:\n%s", first)
	}

	// The edit an interface makes: saved to the supervisor, not yet on disk.
	next := s
	next.SrcdsBotClassBlacklist = []string{"sniper", "spy"}
	sup.SetSettings(next)

	if again := readServerCfg(t, root); again != first {
		t.Fatal("the settings change rewrote server.cfg before any start")
	}

	if err := sup.Start(func(error) {}); err != nil {
		t.Fatalf("second Start: %v", err)
	}
	sup.Stop()

	second := readServerCfg(t, root)
	if !strings.Contains(second, `sm_redbots_manager_class_blacklist "sniper,spy"`) {
		t.Fatalf("the second start kept the stale blacklist:\n%s", second)
	}
}
