package form

import (
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

/*
State is everything a settings screen is editing.

Most rows write a setting, and for a long time that was assumed to be all of
them. It is not. The Bots page has a name to save a team under and a loadout
being built slot by slot, and neither is a setting: nothing in the config file
holds them, they exist for as long as the page is open, and pressing Save turns
them into a preset rather than storing them.

The window and the terminal each kept that scratch in their own screen struct,
which is why the loadout builder was the one part of the settings the two did
not agree about even in shape. So it is here, beside the settings, and a Spec
reads and writes the whole State. What is saved to disk is State.Settings and
nothing else.
*/
type State struct {
	// Settings is what the config file holds and what Save writes.
	Settings settings.Settings

	// Draft is the scratch: real while the page is open, gone when it shuts.
	Draft Draft
}

// Draft is the scratch the Bots page edits before anything is named. A team and
// a loadout are both built up a row at a time and only become a preset when the
// row that saves them is pressed.
type Draft struct {
	// TeamName is what the next "Save this team" keeps the seats under.
	TeamName string

	// LoadoutName is what the loadout being built will be called, and Loadout
	// is the weapons in it. Changing its class clears the slots, because a
	// weapon of the class the loadout no longer belongs to is not a choice
	// anybody made.
	LoadoutName string
	Loadout     botloadout.Built

	// Room is the Archipelago room address as it has been typed. It is held as
	// text rather than as a host and a port because it is parsed, and a row
	// that refused every incomplete address could never be typed into: reaching
	// "archipelago.gg:12345" means passing through "a", which is not an
	// address. The parse happens on Save, and settings.ParseRoom is what says
	// whether what is there now is one.
	Room string
}

/*
	Slot and WithSlot name the four slots the mod's file names

configs/defenderbots/loadout.cfg carries primary, secondary, melee and pda2 and
nothing else, so those are the four. They are separate fields on Built rather
than a map because that file names them, and a map keyed by a string constant is
a lookup nobody can check.
*/
func (d Draft) Slot(key string) int {
	switch key {
	case "primary":
		return d.Loadout.Primary
	case "secondary":
		return d.Loadout.Second
	case "melee":
		return d.Loadout.Melee
	default:
		return d.Loadout.PDA2
	}
}

// WithSlot returns the draft with one slot changed. A copy by value, because a
// Spec's Set must not write through to a State the caller still holds: a
// refused change leaves everything alone.
func (d Draft) WithSlot(key string, defIndex int) Draft {
	switch key {
	case "primary":
		d.Loadout.Primary = defIndex
	case "secondary":
		d.Loadout.Second = defIndex
	case "melee":
		d.Loadout.Melee = defIndex
	default:
		d.Loadout.PDA2 = defIndex
	}
	return d
}

// StockLoadout is every slot on the game's own weapon, for a class.
func StockLoadout(class string) botloadout.Built {
	return botloadout.Built{
		Class:   class,
		Primary: botloadout.Stock,
		Second:  botloadout.Stock,
		Melee:   botloadout.Stock,
		PDA2:    botloadout.Stock,
	}
}

// LoadoutSlots is which slots a class can fill, in the order the pages show
// them. The Spy has no primary: his revolver is the secondary and his watch is
// pda2. His sapper is in the mod's weapon pools and has no key in loadout.cfg,
// so it is not offered; a menu entry the mod never reads reads as a broken
// server rather than a missing feature.
func LoadoutSlots(class string) []string {
	if class == "spy" {
		return []string{"secondary", "melee", "pda2"}
	}
	return []string{"primary", "secondary", "melee"}
}

// NewState is the state a freshly opened settings screen starts in: the saved
// settings, an empty team name, and a stock loadout of the first class.
func NewState(s settings.Settings) State {
	return State{
		Settings: s,
		Draft: Draft{
			Loadout: StockLoadout(botloadout.Classes[0].Key),
			Room: settings.Room{
				Host: s.APHost, Port: s.APPort, TLS: s.APTls,
			}.String(),
		},
	}
}
