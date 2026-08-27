package srcdsconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

// A loadout the player built reaches configs/defenderbots/loadout.cfg, which
// is the only thing the mod reads.
func TestABuiltLoadoutReachesTheGameTree(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "tf-dedicated", "tf", "addons", "sourcemod", "configs", "defenderbots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	s := settings.Settings{
		InstallRoot:          root,
		SrcdsBotTeamComp:     []string{"pyro"},
		SrcdsBotSeatLoadouts: []string{botloadout.CustomKey("Gas runner")},
		SrcdsBotCustomLoadouts: map[string]botloadout.Built{
			"Gas runner": {Class: "pyro", Primary: 594, Second: 1180, Melee: botloadout.Stock, PDA2: botloadout.Stock},
		},
	}
	if err := Install(s); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "loadout.cfg"))
	if err != nil {
		t.Fatalf("no loadout file: %v", err)
	}
	for _, want := range []string{"\"class\"\t\"pyro\"", "\"primary\"\t\"594\"", "\"secondary\"\t\"1180\""} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("missing %q in:\n%s", want, raw)
		}
	}
}
