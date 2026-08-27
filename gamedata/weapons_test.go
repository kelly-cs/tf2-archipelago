package gamedata

import (
	"strings"
	"testing"
)

/*
	Every preset the launcher offers names item definition indexes, and the

catalogue has to be able to name them back, or a menu shows a number.

This is the test that catches the catalogue drifting from the mod's pools: a
weapon the mod stopped offering, or a class whose pool was retyped.
*/
func TestEveryClassCanFillEverySlot(t *testing.T) {
	classes := []string{"scout", "soldier", "pyro", "demoman", "heavyweapons", "engineer", "medic", "sniper", "spy"}
	for _, class := range classes {
		for _, slot := range []string{"primary", "secondary", "melee"} {
			// The Spy has no primary: his pool is the revolver in secondary,
			// the sapper in building and the watch in pda2.
			if class == "spy" && slot == "primary" {
				continue
			}
			if len(WeaponsFor(class, slot)) == 0 {
				t.Errorf("%s has nothing to hold in %s", class, slot)
			}
		}
	}
	for _, slot := range []string{"pda2", "building"} {
		if len(WeaponsFor("spy", slot)) == 0 {
			t.Errorf("the spy has nothing in %s", slot)
		}
	}
}

// A repaint of a gun already in the list is what the catalogue exists to keep
// out: twelve Scatterguns in one menu is not a menu.
func TestTheRepaintsAreNotInTheCatalogue(t *testing.T) {
	for _, weapon := range Weapons {
		for _, mark := range []string{"Botkiller", "Festive ", "Australium "} {
			if strings.Contains(weapon.Name, mark) {
				t.Errorf("%d %q is a repaint", weapon.DefIndex, weapon.Name)
			}
		}
	}
	if got := len(WeaponsFor("scout", "primary")); got > 8 {
		t.Errorf("%d scout primaries, which reads as a list rather than a menu", got)
	}
}

/*
	One index is one weapon, whatever class holds it.

The shotgun is in four pools and the Frying Pan in nine, so an index appearing
more than once is expected. Two names for one index is not: it would make the
same pick read differently depending on the seat it was made from.
*/
func TestOneIndexIsOneWeapon(t *testing.T) {
	names := map[int]string{}
	for _, weapon := range Weapons {
		if was, seen := names[weapon.DefIndex]; seen && was != weapon.Name {
			t.Errorf("%d is %q and also %q", weapon.DefIndex, was, weapon.Name)
		}
		names[weapon.DefIndex] = weapon.Name
	}
	if _, found := WeaponByIndex(-1); found {
		t.Error("stock is not a catalogue entry")
	}
	if weapon, found := WeaponByIndex(448); !found || !strings.Contains(weapon.Name, "Soda Popper") {
		t.Errorf("448 = %+v, want the Soda Popper", weapon)
	}
}

// Nothing in the catalogue is nameless or unattached: a blank cell is a menu
// entry nobody can pick on purpose.
func TestEveryEntryIsComplete(t *testing.T) {
	for _, weapon := range Weapons {
		if weapon.Name == "" || weapon.Class == "" || weapon.Slot == "" || weapon.DefIndex <= 0 {
			t.Errorf("incomplete entry: %+v", weapon)
		}
	}
}
