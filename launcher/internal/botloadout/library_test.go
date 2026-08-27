package botloadout

import (
	"strings"
	"testing"
)

// A built loadout reaches the file the mod reads, the same way a preset does.
func TestABuiltLoadoutIsWrittenLikeAPreset(t *testing.T) {
	library := Library{Built: map[string]Built{
		"Gas runner": {Class: "pyro", Primary: 594, Second: 1180, Melee: Stock, PDA2: Stock},
	}}
	got := library.Render(map[string]string{"pyro": CustomKey("Gas runner")}, nil)

	if !strings.Contains(got, "\"pyro\"\n\t{\n\t\t\"primary\"\t\"594\"\n\t\t\"secondary\"\t\"1180\"\n\t}") {
		t.Errorf("the built loadout is not in the file:\n%s", got)
	}
	// Stock in a slot leaves that weapon alone, so the slot is left out.
	if strings.Contains(got, "\"melee\"") {
		t.Errorf("a stock slot was written:\n%s", got)
	}
}

/*
	A loadout the player has deleted falls to stock rather than breaking.

The team that named it is still on disk, and refusing to load it would lose the
seats and the blacklist along with the weapons.
*/
func TestADeletedLoadoutIsStock(t *testing.T) {
	empty := Library{}
	if got := empty.Loadout(Classes[0], CustomKey("gone")); got.Key != StockKey {
		t.Errorf("a missing custom loadout = %q, want stock", got.Key)
	}
	if empty.Anything(map[string]string{"scout": CustomKey("gone")}, nil) {
		t.Error("a missing custom loadout asked for the file to be written")
	}
}

// A Medic cannot hold a Gunslinger, so a loadout built for another class is
// not a loadout this seat can wear.
func TestALoadoutBelongsToOneClass(t *testing.T) {
	library := Library{Built: map[string]Built{
		"Nest": {Class: "engineer", Primary: 997, Second: Stock, Melee: Stock, PDA2: Stock},
	}}
	medic, _ := ClassByKey("medic")
	if got := library.Loadout(medic, CustomKey("Nest")); got.Key != StockKey {
		t.Errorf("the medic wore an engineer's loadout: %q", got.Key)
	}
	engineer, _ := ClassByKey("engineer")
	if got := library.Loadout(engineer, CustomKey("Nest")); got.Primary != 997 {
		t.Errorf("the engineer did not get it: %+v", got)
	}
}

// The name is what the player typed, and the line under it says what the
// loadout actually holds.
func TestABuiltLoadoutNamesItsWeapons(t *testing.T) {
	library := Library{Built: map[string]Built{
		"Gas runner": {Class: "pyro", Primary: 594, Second: 1180, Melee: Stock, PDA2: Stock},
	}}
	pyro, _ := ClassByKey("pyro")
	got := library.Loadout(pyro, CustomKey("Gas runner"))

	if got.Name != "Gas runner" {
		t.Errorf("name = %q", got.Name)
	}
	for _, want := range []string{"Phlogistinator", "Gas Passer"} {
		if !strings.Contains(got.Weapons, want) {
			t.Errorf("%q is not in %q", want, got.Weapons)
		}
	}
	if strings.Contains(got.Weapons, "stock") {
		t.Errorf("a stock slot was named: %q", got.Weapons)
	}
}

// The prefix is what separates the two tables, and a built-in key must never
// be read as a custom one.
func TestTheKeysDoNotCollide(t *testing.T) {
	if CustomName("kunai") != "" {
		t.Error("a built-in key read as custom")
	}
	if CustomName(CustomKey("kunai")) != "kunai" {
		t.Error("a custom loadout may be named after a preset")
	}
	if got := CustomKey("Gas runner"); got != "custom:Gas runner" {
		t.Errorf("key = %q", got)
	}
}

// An index the catalogue does not carry is still a legal pick, because the mod
// validates nothing. It shows as its number rather than blank.
func TestAnUnknownWeaponShowsItsNumber(t *testing.T) {
	if got := WeaponName(999999); got != "item 999999" {
		t.Errorf("WeaponName(999999) = %q", got)
	}
	if got := WeaponName(Stock); got != "stock" {
		t.Errorf("WeaponName(stock) = %q", got)
	}
}
