package gamedata

// Weapon is one item the bot mod will hand a defender bot: the definition
// index the mod writes into its loadout file, and what to call it in a menu.
type Weapon struct {
	DefIndex int
	Name     string
	Class    string
	Slot     string
}

// WeaponsFor is what one class can hold in one slot, in the order a menu
// should show them. Empty for a pair the mod has no pool for, which is every
// slot the Spy does not have and the Spy's own pda2 for everybody else.
func WeaponsFor(class, slot string) []Weapon {
	var out []Weapon
	for _, weapon := range Weapons {
		if weapon.Class == class && weapon.Slot == slot {
			out = append(out, weapon)
		}
	}
	return out
}

// WeaponByIndex is the item with that definition index, and whether the
// catalogue carries it at all. A loadout naming an index this does not know is
// still legal, because the mod validates nothing: it just cannot be named.
func WeaponByIndex(defIndex int) (Weapon, bool) {
	for _, weapon := range Weapons {
		if weapon.DefIndex == defIndex {
			return weapon, true
		}
	}
	return Weapon{}, false
}
