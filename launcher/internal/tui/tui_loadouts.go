package tui

import (
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
)

/*
The Loadouts page: one editor, not a list.

Pick a class, pick a weapon per slot, name it and save. The saved ones then
appear at the bottom of every loadout menu on the Bots page, for that class
only, because a Medic cannot hold a Gunslinger.

The editor holds one loadout at a time in loadoutDraft. Nothing here writes to
the settings until Save, so leaving the page changes nothing.
*/
type loadoutDraft struct {
	name  string
	class string
	built botloadout.Built
}

/*
	loadoutSlots is which slots a class can fill, in the order the page shows

them.

Four slots, because four is what configs/defenderbots/loadout.cfg carries:
writeSlots names primary, secondary, melee and pda2 and nothing else. The Spy's
sapper is in the mod's weapon pools and has no key in that file, so it is not
offered. A menu entry the mod never reads is worse than a missing one, because
it reads as a broken server rather than a missing feature.

The Spy has no primary: his revolver is the secondary and his watch is pda2.
*/
func loadoutSlots(class string) []string {
	if class == "spy" {
		return []string{"secondary", "melee", "pda2"}
	}
	return []string{"primary", "secondary", "melee"}
}

// slotName is what the page calls a slot.
func slotName(slot string) string {
	switch slot {
	case "pda2":
		return "Watch"
	default:
		return strings.ToUpper(slot[:1]) + slot[1:]
	}
}

// slotValue reads one slot out of the draft, and setSlot writes it. The four
// are separate fields rather than an array because the mod's file names them,
// and a slice indexed by a string constant is a lookup nobody can read.
func (d *loadoutDraft) slotValue(slot string) int {
	switch slot {
	case "primary":
		return d.built.Primary
	case "secondary":
		return d.built.Second
	case "melee":
		return d.built.Melee
	default:
		return d.built.PDA2
	}
}

func (d *loadoutDraft) setSlot(slot string, defIndex int) {
	switch slot {
	case "primary":
		d.built.Primary = defIndex
	case "secondary":
		d.built.Second = defIndex
	case "melee":
		d.built.Melee = defIndex
	default:
		d.built.PDA2 = defIndex
	}
}

// stockDraft is an empty loadout for a class: every slot on stock.
func stockDraft(class string) loadoutDraft {
	return loadoutDraft{
		class: class,
		built: botloadout.Built{
			Class:   class,
			Primary: botloadout.Stock,
			Second:  botloadout.Stock,
			Melee:   botloadout.Stock,
			PDA2:    botloadout.Stock,
		},
	}
}

func (f *settingsForm) loadoutFields() []field {
	if f.draft.class == "" {
		f.draft = stockDraft(botloadout.Classes[0].Key)
	}

	fields := []field{
		f.draftClassField(),
	}
	for _, slot := range loadoutSlots(f.draft.class) {
		fields = append(fields, f.draftSlotField(slot))
	}
	return append(fields,
		&textField{
			label:       "Name",
			help:        "What this loadout is called in the menus on the Bots page.",
			value:       &f.draft.name,
			placeholder: "gas runner",
		},
		f.saveLoadoutField(),
		f.loadLoadoutField(),
		f.removeLoadoutField(),
	)
}

// draftClassField picks the class. Changing it clears the slots, because a
// weapon of the class this loadout no longer belongs to is not a choice
// anybody made.
func (f *settingsForm) draftClassField() field {
	options := make([]string, 0, len(botloadout.Classes))
	for _, class := range botloadout.Classes {
		options = append(options, class.Name)
	}
	index := max(slices.IndexFunc(botloadout.Classes, func(c botloadout.Class) bool {
		return c.Key == f.draft.class
	}), 0)

	return &choiceField{
		label:   "Class",
		help:    "Who holds this loadout. A loadout belongs to one class, so only that class can pick it.",
		options: options,
		index:   index,
		apply: func(i int) {
			class := botloadout.Classes[i].Key
			if class == f.draft.class {
				return
			}
			name := f.draft.name
			f.draft = stockDraft(class)
			f.draft.name = name
			f.build()
		},
	}
}

// draftSlotField is one weapon menu. Stock first, then what the catalogue says
// this class can hold in this slot.
func (f *settingsForm) draftSlotField(slot string) field {
	weapons := gamedata.WeaponsFor(f.draft.class, slot)
	options := make([]string, 0, len(weapons)+1)
	options = append(options, "stock")
	for _, weapon := range weapons {
		options = append(options, weapon.Name)
	}

	current := f.draft.slotValue(slot)
	index := 0
	if at := slices.IndexFunc(weapons, func(w gamedata.Weapon) bool { return w.DefIndex == current }); at >= 0 {
		index = at + 1
	}

	return &choiceField{
		label:   "  " + slotName(slot),
		help:    "The weapon in this slot. Stock leaves the game's own alone.",
		options: options,
		index:   index,
		apply: func(i int) {
			if i == 0 {
				f.draft.setSlot(slot, botloadout.Stock)
				return
			}
			f.draft.setSlot(slot, weapons[i-1].DefIndex)
		},
	}
}

func (f *settingsForm) saveLoadoutField() field {
	return &actionField{
		label: "Save this loadout as",
		help:  "Keeps the weapons above under the name in the box. Saving over a name replaces it.",
		hint:  "enter",
		run: func() tea.Cmd {
			name := strings.TrimSpace(f.draft.name)
			if name == "" {
				return func() tea.Msg { return noticeMsg("name the loadout first") }
			}
			if f.edited.SrcdsBotCustomLoadouts == nil {
				f.edited.SrcdsBotCustomLoadouts = map[string]botloadout.Built{}
			}
			f.edited.SrcdsBotCustomLoadouts[name] = f.draft.built
			f.build()
			return func() tea.Msg { return noticeMsg("saved the loadout as " + name) }
		},
	}
}

func (f *settingsForm) loadLoadoutField() field {
	names := f.builtNames()
	options := append([]string{"keep the loadout above"}, names...)

	return &choiceField{
		label:   "Load a loadout",
		help:    "Brings a saved loadout back into the menus above, to change it or to save it under another name.",
		options: options,
		apply: func(i int) {
			if i == 0 || i > len(names) {
				return
			}
			name := names[i-1]
			f.draft = loadoutDraft{
				name: name, class: f.edited.SrcdsBotCustomLoadouts[name].Class,
				built: f.edited.SrcdsBotCustomLoadouts[name],
			}
			f.build()
		},
	}
}

func (f *settingsForm) removeLoadoutField() field {
	names := f.builtNames()
	options := append([]string{"keep them all"}, names...)

	return &choiceField{
		label:   "Remove a loadout",
		help:    "Throws one away. A seat still naming it plays stock rather than refusing to load.",
		options: options,
		apply: func(i int) {
			if i == 0 || i > len(names) {
				return
			}
			delete(f.edited.SrcdsBotCustomLoadouts, names[i-1])
			f.build()
		},
	}
}

// builtNames is every saved loadout, in one order so the menus do not shuffle.
func (f *settingsForm) builtNames() []string {
	names := make([]string, 0, len(f.edited.SrcdsBotCustomLoadouts))
	for name := range f.edited.SrcdsBotCustomLoadouts {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
