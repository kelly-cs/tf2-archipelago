package botloadout

import (
	"strings"
	"testing"
)

func TestRenderWritesOnlyWhatChanges(t *testing.T) {
	got := Render(map[string]string{"scout": "milk", "spy": "kunai", "medic": StockKey, "pyro": "gone"})

	for _, want := range []string{
		"\"scout\"\n\t{\n\t\t\"primary\"\t\"448\"\n\t\t\"secondary\"\t\"222\"\n\t\t\"melee\"\t\"355\"\n\t}",
		"\"spy\"\n\t{\n\t\t\"primary\"\t\"61\"\n\t\t\"melee\"\t\"356\"\n\t\t\"pda2\"\t\"59\"\n\t}",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	for _, absent := range []string{"\"medic\"", "\"pyro\"", "\"soldier\""} {
		if strings.Contains(got, absent) {
			t.Errorf("%s has no preset and must not be in:\n%s", absent, got)
		}
	}
}

func TestCustomIsAnyPresetButStock(t *testing.T) {
	if Custom(map[string]string{"scout": StockKey, "spy": ""}) {
		t.Error("stock everywhere counts as custom")
	}
	if !Custom(map[string]string{"heavyweapons": "brass"}) {
		t.Error("a preset does not count as custom")
	}
}

func TestBlacklistKeepsTheModsOrderAndSpelling(t *testing.T) {
	if got := Blacklist([]string{"spy", "heavy", "sniper", "nobody"}); got != "sniper,spy" {
		t.Errorf("blacklist = %q", got)
	}
	// Order and repeats are the point of a composition, and a blacklist has
	// neither: it is a set, so it renders in the table's order.
	if got := Composition([]string{"medic", "heavyweapons", "nobody", "heavyweapons"}); got != "medic,heavyweapons,heavyweapons" {
		t.Errorf("composition = %q", got)
	}
	if got := Composition(nil); got != "" {
		t.Errorf("empty composition = %q", got)
	}
	if got := Blacklist(nil); got != "" {
		t.Errorf("empty blacklist = %q", got)
	}
}

// Every class starts with stock, and every preset key is unique within its
// class: a duplicate key would make the menu pick the wrong one.
func TestPresetsAreWellFormed(t *testing.T) {
	for _, class := range Classes {
		if class.Loadouts[0].Key != StockKey {
			t.Errorf("%s does not start with stock", class.Key)
		}
		seen := map[string]bool{}
		for _, loadout := range class.Loadouts {
			if seen[loadout.Key] {
				t.Errorf("%s has two presets named %s", class.Key, loadout.Key)
			}
			seen[loadout.Key] = true
		}
	}
}
