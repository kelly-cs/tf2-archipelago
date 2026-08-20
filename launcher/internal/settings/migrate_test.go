package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A config file written before SrcdsReach existed says only whether sv_lan was
// on. Reading one has to land on the reach it meant: a file that said sv_lan
// was on is a file that asked to stay on the local network.
func TestOldConfigPicksItsReach(t *testing.T) {
	cases := []struct {
		name string
		body string
		want Reach
	}{
		{"sv_lan on", `{"srcds_lan": true}`, ReachLan},
		{"sv_lan off", `{"srcds_lan": false}`, ReachPort},
		{"no reach at all", `{}`, ReachLan},
		{"a reach nobody knows", `{"srcds_reach": "carrier-pigeon"}`, ReachLan},
		{"a reach and the old flag", `{"srcds_reach": "steam", "srcds_lan": true}`, ReachSteam},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("APPDATA", filepath.Join(t.TempDir(), "AppData", "Roaming"))
			path, err := Path()
			if err != nil {
				t.Fatalf("Path: %v", err)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			if err := os.WriteFile(path, []byte(c.body), 0o600); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			loaded, err := Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if loaded.SrcdsReach != c.want {
				t.Errorf("SrcdsReach = %q, want %q", loaded.SrcdsReach, c.want)
			}
			// The old flag is read once and dropped, so the file does not end
			// up carrying two answers to the same question.
			if loaded.SrcdsLanLegacy != nil {
				t.Error("the legacy flag survived the load")
			}
			rendered, err := Render(loaded)
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if strings.Contains(rendered, "srcds_lan") {
				t.Errorf("srcds_lan written back:\n%s", rendered)
			}
		})
	}
}

// SRCDS_LAN is what every existing .env file says. It has to keep working, and
// SRCDS_REACH has to win where both are set.
func TestReachFromEnv(t *testing.T) {
	cases := []struct {
		name  string
		lan   string
		reach string
		want  Reach
	}{
		{"neither", "", "", ReachLan},
		{"old flag on", "1", "", ReachLan},
		{"old flag off", "0", "", ReachPort},
		{"reach alone", "", "steam", ReachSteam},
		{"reach wins over the old flag", "1", "steam", ReachSteam},
		{"reach wins the other way", "0", "lan", ReachLan},
		{"a reach nobody knows changes nothing", "0", "smoke-signal", ReachPort},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.lan != "" {
				t.Setenv("SRCDS_LAN", c.lan)
			}
			if c.reach != "" {
				t.Setenv("SRCDS_REACH", c.reach)
			}
			if got := ApplyEnv(Defaults()).SrcdsReach; got != c.want {
				t.Errorf("SrcdsReach = %q, want %q", got, c.want)
			}
		})
	}
}
