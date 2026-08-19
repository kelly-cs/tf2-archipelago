// Package botloadout is what the launcher knows about the defender bots'
// classes and weapons: which classes the mod names how, and a handful of
// loadouts per class worth handing a bot. The mod reads the result from
// configs/defenderbots/loadout.cfg, a KeyValues file this package renders.
package botloadout

import (
	"fmt"
	"slices"
	"strings"
)

// Stock is the item definition the mod reads as "leave the stock weapon".
const Stock = -1

// StockKey is the preset that changes nothing, and the one a class without a
// saved choice gets.
const StockKey = "stock"

// Loadout is one named set of weapons for one class. The indexes are TF2 item
// definition indexes, and every one of them is in the pool the mod itself
// draws random loadouts from, so a bot knows how to hold it.
type Loadout struct {
	Key     string
	Name    string
	Weapons string
	Primary int
	Second  int
	Melee   int
	// PDA2 is the Spy's watch. Stock for every other class.
	PDA2 int
}

// Label is what a menu shows: the name, then the weapons behind it.
func (l Loadout) Label() string {
	if l.Key == StockKey {
		return l.Name
	}
	return fmt.Sprintf("%s: %s", l.Name, l.Weapons)
}

// Class is one class as the mod names it, with its presets. Stock is always
// first.
type Class struct {
	Key      string
	Name     string
	Loadouts []Loadout
}

var stock = Loadout{Key: StockKey, Name: "Stock weapons", Primary: Stock, Second: Stock, Melee: Stock, PDA2: Stock}

// Classes lists the nine classes in the class menu's order, keyed the way the
// mod's convars and files spell them.
var Classes = []Class{
	{"scout", "Scout", []Loadout{
		stock,
		{Key: "milk", Name: "Milk runner", Weapons: "Soda Popper, Mad Milk, Fan O'War", Primary: 448, Second: 222, Melee: 355, PDA2: Stock},
		{Key: "fan", Name: "Force-A-Nature", Weapons: "Force-A-Nature, Pistol, Sandman", Primary: 45, Second: Stock, Melee: 44, PDA2: Stock},
	}},
	{"soldier", "Soldier", []Loadout{
		stock,
		{Key: "banner", Name: "Banner", Weapons: "Black Box, Buff Banner, Escape Plan", Primary: 228, Second: 129, Melee: 775, PDA2: Stock},
		{Key: "beggar", Name: "Beggar", Weapons: "Beggar's Bazooka, Buff Banner, Escape Plan", Primary: 730, Second: 129, Melee: 775, PDA2: Stock},
	}},
	{"pyro", "Pyro", []Loadout{
		stock,
		{Key: "phlog", Name: "Phlog", Weapons: "Phlogistinator, Scorch Shot, Powerjack", Primary: 594, Second: 740, Melee: 214, PDA2: Stock},
		{Key: "flare", Name: "Flare", Weapons: "Backburner, Flare Gun, Axtinguisher", Primary: 40, Second: 39, Melee: 38, PDA2: Stock},
	}},
	{"demoman", "Demoman", []Loadout{
		stock,
		{Key: "resistance", Name: "Scottish Resistance", Weapons: "Iron Bomber, Scottish Resistance, Bottle", Primary: 1151, Second: 130, Melee: Stock, PDA2: Stock},
		{Key: "cannon", Name: "Loose Cannon", Weapons: "Loose Cannon, Stickybomb Launcher, Pain Train", Primary: 996, Second: Stock, Melee: 154, PDA2: Stock},
	}},
	{"heavyweapons", "Heavy", []Loadout{
		stock,
		{Key: "brass", Name: "Brass Beast", Weapons: "Brass Beast, Family Business, Fists of Steel", Primary: 312, Second: 425, Melee: 331, PDA2: Stock},
		{Key: "tomislav", Name: "Tomislav", Weapons: "Tomislav, Panic Attack, Gloves of Running Urgently", Primary: 424, Second: 1153, Melee: 239, PDA2: Stock},
	}},
	{"engineer", "Engineer", []Loadout{
		stock,
		{Key: "ranger", Name: "Rescue Ranger", Weapons: "Rescue Ranger, Wrangler, Jag", Primary: 997, Second: 140, Melee: 329, PDA2: Stock},
		{Key: "widowmaker", Name: "Widowmaker", Weapons: "Widowmaker, Short Circuit, Jag", Primary: 527, Second: 528, Melee: 329, PDA2: Stock},
	}},
	{"medic", "Medic", []Loadout{
		stock,
		{Key: "kritz", Name: "Kritzkrieg", Weapons: "Crusader's Crossbow, Kritzkrieg, Ubersaw", Primary: 305, Second: 35, Melee: 37, PDA2: Stock},
		{Key: "quickfix", Name: "Quick-Fix", Weapons: "Overdose, Quick-Fix, Amputator", Primary: 412, Second: 411, Melee: 304, PDA2: Stock},
	}},
	{"sniper", "Sniper", []Loadout{
		stock,
		{Key: "heatmaker", Name: "Heatmaker", Weapons: "Hitman's Heatmaker, Jarate, Bushwacka", Primary: 752, Second: 58, Melee: 232, PDA2: Stock},
		{Key: "machina", Name: "Machina", Weapons: "Machina, Razorback, Kukri", Primary: 526, Second: 57, Melee: Stock, PDA2: Stock},
	}},
	{"spy", "Spy", []Loadout{
		stock,
		{Key: "diamondback", Name: "Diamondback", Weapons: "Diamondback, Red-Tape Recorder, Big Earner", Primary: 525, Second: 810, Melee: 461, PDA2: Stock},
		{Key: "kunai", Name: "Kunai", Weapons: "Ambassador, Sapper, Conniver's Kunai, Dead Ringer", Primary: 61, Second: Stock, Melee: 356, PDA2: 59},
	}},
}

// ClassByKey finds a class by the mod's key.
func ClassByKey(key string) (Class, bool) {
	for _, class := range Classes {
		if class.Key == key {
			return class, true
		}
	}
	return Class{}, false
}

// LoadoutByKey finds a preset of a class. An unknown key is stock, so a saved
// choice that no longer exists changes nothing.
func (c Class) LoadoutByKey(key string) Loadout {
	for _, loadout := range c.Loadouts {
		if loadout.Key == key {
			return loadout
		}
	}
	return stock
}

// Custom reports whether any class has a preset other than stock, which is
// what decides whether the mod's custom loadouts are turned on at all.
func Custom(picks map[string]string) bool {
	for _, class := range Classes {
		if picks[class.Key] != "" && picks[class.Key] != StockKey {
			return true
		}
	}
	return false
}

// Render is configs/defenderbots/loadout.cfg: one block per class with a
// preset, one key per slot the preset sets. A slot left out keeps the stock
// weapon, and a class left out plays stock.
func Render(picks map[string]string) string {
	var b strings.Builder
	b.WriteString("// Managed by tf2ap. Edits here are replaced the next time the launcher starts.\n")
	b.WriteString("\"loadout\"\n{\n")
	for _, class := range Classes {
		loadout := class.LoadoutByKey(picks[class.Key])
		if loadout.Key == StockKey {
			continue
		}
		fmt.Fprintf(&b, "\t\"%s\"\n\t{\n", class.Key)
		for _, slot := range []struct {
			key   string
			index int
		}{{"primary", loadout.Primary}, {"secondary", loadout.Second}, {"melee", loadout.Melee}, {"pda2", loadout.PDA2}} {
			if slot.index != Stock {
				fmt.Fprintf(&b, "\t\t\"%s\"\t\"%d\"\n", slot.key, slot.index)
			}
		}
		b.WriteString("\t}\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// Blacklist is the value of sm_redbots_manager_class_blacklist: the class keys
// the bots may not play, comma-separated, unknown keys dropped.
func Blacklist(classes []string) string {
	kept := make([]string, 0, len(classes))
	for _, class := range Classes {
		if slices.Contains(classes, class.Key) {
			kept = append(kept, class.Key)
		}
	}
	return strings.Join(kept, ",")
}
