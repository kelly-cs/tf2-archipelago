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

/* Render is configs/defenderbots/loadout.cfg.
 *
 * One block per class with a preset, one key per slot the preset sets. A slot
 * left out keeps the stock weapon, and a class left out plays stock. That is
 * what the mod reads today: GetServerLoadoutWeapon looks a class up by name.
 *
 * The seats go in as well, under "seats", one numbered block each carrying the
 * class that seat plays and the weapons it holds. Two engineers cannot hold
 * different weapons through a table keyed by class, and one holding the
 * Wrangler while the other holds something else is the reason to name a team
 * seat by seat at all.
 *
 * The mod does not read that half yet. Written anyway, and written first,
 * because a KeyValues file the reader has not caught up with costs nothing and
 * the alternative is a launcher that cannot say what it means.
 */
func Render(picks map[string]string, seats []Seat) string {
	var b strings.Builder
	b.WriteString("// Managed by tf2ap. Edits here are replaced the next time the launcher starts.\n")
	b.WriteString("\"loadout\"\n{\n")
	for _, class := range Classes {
		loadout := class.LoadoutByKey(picks[class.Key])
		if loadout.Key == StockKey {
			continue
		}
		fmt.Fprintf(&b, "\t\"%s\"\n\t{\n", class.Key)
		writeSlots(&b, "\t\t", loadout)
		b.WriteString("\t}\n")
	}
	writeSeats(&b, seats)
	b.WriteString("}\n")
	return b.String()
}

// Seat is one place on RED: the class it plays and the loadout it carries.
// An empty class is a seat the mod draws for itself.
type Seat struct {
	Class   string
	Loadout string
}

// Seats pairs the team's classes with the loadouts chosen for each place. The
// two lists are stored apart because the classes are also what the mod's
// team_composition convar carries, and a seat with no loadout of its own is
// not the same as a seat playing stock: it falls back to the class's pick.
func Seats(comp, loadouts []string) []Seat {
	seats := make([]Seat, 0, len(comp))
	for index, class := range comp {
		seat := Seat{Class: class}
		if index < len(loadouts) {
			seat.Loadout = loadouts[index]
		}
		seats = append(seats, seat)
	}
	return seats
}

// CustomSeats reports whether any seat asks for weapons of its own.
func CustomSeats(seats []Seat) bool {
	for _, seat := range seats {
		if seat.Class == "" || seat.Loadout == "" || seat.Loadout == StockKey {
			continue
		}
		return true
	}
	return false
}

func writeSeats(b *strings.Builder, seats []Seat) {
	named := 0
	for _, seat := range seats {
		if seat.Class != "" {
			named++
		}
	}
	if named == 0 {
		return
	}
	b.WriteString("\t\"seats\"\n\t{\n")
	for index, seat := range seats {
		if seat.Class == "" {
			continue
		}
		class, found := ClassByKey(seat.Class)
		if !found {
			continue
		}
		fmt.Fprintf(b, "\t\t\"%d\"\n\t\t{\n", index+1)
		fmt.Fprintf(b, "\t\t\t\"class\"\t\"%s\"\n", class.Key)
		writeSlots(b, "\t\t\t", class.LoadoutByKey(seat.Loadout))
		b.WriteString("\t\t}\n")
	}
	b.WriteString("\t}\n")
}

func writeSlots(b *strings.Builder, indent string, loadout Loadout) {
	for _, slot := range []struct {
		key   string
		index int
	}{{"primary", loadout.Primary}, {"secondary", loadout.Second}, {"melee", loadout.Melee}, {"pda2", loadout.PDA2}} {
		if slot.index != Stock {
			fmt.Fprintf(b, "%s\"%s\"\t\"%d\"\n", indent, slot.key, slot.index)
		}
	}
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

/* Composition is the value of sm_redbots_manager_team_composition: the class
 * keys the bots fill RED with, one entry per seat, in seat order. The mod
 * fills the first entries when the team is short, and keeps the repeats: two
 * Heavies is a team somebody asked for.
 *
 * A seat left on the draw is an empty entry. The mod counts a seat by its
 * place in this list. Drop the empty and seat 4 becomes seat 1, and the
 * loadout file names the wrong bot.
 *
 * An empty string makes the mod play the map's own default lineup, and a named
 * lineup beats the blacklist. So a team of nothing but draws still writes its
 * seats when a class is unticked.
 */
func Composition(comp, blacklist []string) string {
	seats := make([]string, 0, len(comp))
	for _, wanted := range comp {
		key := ""
		for _, class := range Classes {
			if class.Key == wanted {
				key = class.Key
				break
			}
		}
		seats = append(seats, key)
	}
	for len(seats) > 0 && seats[len(seats)-1] == "" {
		seats = seats[:len(seats)-1]
	}
	if len(seats) == 0 {
		if Blacklist(blacklist) == "" {
			return ""
		}
		if len(comp) > 0 {
			seats = make([]string, len(comp))
		} else {
			seats = make([]string, botSeats)
		}
	}
	return strings.Join(seats, ",")
}

// botSeats is how many seats a team of nothing but draws writes out. At least
// as many as the settings page offers.
const botSeats = 6
