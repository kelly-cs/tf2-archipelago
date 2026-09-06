package gamedata

// ItemKind is what an item does when it lands. It is also the plugin's grant
// vocabulary: the bridge sends the kind and its payload, never an item id.
type ItemKind uint8

const (
	ItemMissionTicket ItemKind = iota + 1
	ItemClass
	ItemWeaponSlot
	ItemCredits
	ItemWeaponBuff
	ItemTrap
	ItemServerSetting
	ItemTrophy
)

var itemKindKeys = [...]string{
	ItemMissionTicket: "mission_ticket",
	ItemClass:         "class",
	ItemWeaponSlot:    "weapon_slot",
	ItemCredits:       "credits",
	ItemWeaponBuff:    "weapon_buff",
	ItemTrap:          "trap",
	ItemServerSetting: "server_setting",
	ItemTrophy:        "trophy",
}

// ItemKinds is every kind that exists, in id order. The bridge walks it to
// build the unlock set, so a kind added here needs no second list anywhere.
var ItemKinds = []ItemKind{
	ItemMissionTicket, ItemClass, ItemWeaponSlot, ItemCredits, ItemWeaponBuff, ItemTrap,
	ItemServerSetting, ItemTrophy,
}

// Key is the string on the wire between the bridge and the plugin.
func (k ItemKind) Key() string { return itemKindKeys[k] }

// OneShot reports whether applying the item a second time differs from applying
// it once. A class is state: granting it again changes nothing, so it can be
// replayed after any reload. Credits are an effect: granting them again pays a
// second time.
//
// The distinction is what decides how the bridge delivers an item. State goes
// in the unlock set and is resent freely; an effect is sent once and is not
// sent again until the plugin says it applied it. A trap is an effect for the
// same reason: firing it twice soaks a team that only earned it once.
func (k ItemKind) OneShot() bool { return k == ItemCredits || k == ItemTrap }

/*
Granted reports whether the plugin ever sees this kind.

A trophy is not a grant. It is locked onto a mission clear so that generation
can ask whether the mission is done, and nothing happens in the game when it
lands: the bridge drops it and the plugin is never told. Every other kind is
something a player receives, and one the plugin does not handle is an item the
seed loses in silence, which is what the plugin-keys test is for.
*/
func (k ItemKind) Granted() bool { return k != ItemTrophy }

// Item is one entry in the multiworld's item pool. Mission, Class and Credits
// are the payload of the kind that uses them and zero elsewhere; Count is zero
// for filler, whose copy count is decided at generation time.
type Item struct {
	ID             int64
	Name           string
	Kind           ItemKind
	Classification Classification
	Count          uint8
	Mission        MissionID
	Class          ClassID
	Credits        uint16
	WeaponBuff     uint16
	Trap           TrapID
	ServerSetting  ServerSettingID
}

// ProgressiveWeaponSlotName is the one item that unlocks loadout slots: copy n
// grants WeaponSlots[n-1].
const ProgressiveWeaponSlotName = "Progressive Weapon Slot"

// progressiveWeaponSlotID takes offset zero, leaving 1 to 3 free for per-slot items later.
var progressiveWeaponSlotID = BaseID + itemSpaceOffset + itemBlockWeaponSlot

// cashBundleCredits is kept low so a filler-heavy sphere cannot buy a wave outright.
const cashBundleCredits uint16 = 200

var cashBundleID = BaseID + itemSpaceOffset + itemBlockCredits + 1

// Items is the whole item pool template: a ticket per mission, a class item
// per class, the progressive weapon slot, and the filler that pads the rest.
//
// Weapon ownership, canteens and robot templates remain out of scope.
var Items = buildItems()

var itemsByID = indexItems()

func indexItems() map[int64]Item {
	byID := make(map[int64]Item, len(Items))
	for _, it := range Items {
		byID[it.ID] = it
	}
	return byID
}

// ItemByID resolves an id from a ReceivedItems payload into the kind the
// plugin is told about.
func ItemByID(id int64) (Item, bool) {
	it, ok := itemsByID[id]
	return it, ok
}

func buildItems() []Item {
	all := make([]Item, 0, len(Missions)+len(Classes)+len(WeaponBuffs)+len(Traps)+2)
	for _, m := range Missions {
		all = append(all, Item{
			ID:             m.TicketItemID(),
			Name:           m.TicketItemName(),
			Kind:           ItemMissionTicket,
			Classification: Progression,
			Count:          1,
			Mission:        m.ID,
		})
	}
	for _, c := range Classes {
		all = append(all, Item{
			ID:             c.ItemID(),
			Name:           c.ItemName(),
			Kind:           ItemClass,
			Classification: Progression,
			Count:          1,
			Class:          c.ID,
		})
	}
	all = append(all, Item{
		ID:             progressiveWeaponSlotID,
		Name:           ProgressiveWeaponSlotName,
		Kind:           ItemWeaponSlot,
		Classification: Progression,
		Count:          uint8(len(WeaponSlots)),
	})
	all = append(all, Item{
		ID:             cashBundleID,
		Name:           "Cash Bundle",
		Kind:           ItemCredits,
		Classification: Filler,
		Credits:        cashBundleCredits,
	})
	for _, buff := range WeaponBuffs {
		all = append(all, Item{
			ID:             buff.ItemID(),
			Name:           buff.ItemName(),
			Kind:           ItemWeaponBuff,
			Classification: Useful,
			WeaponBuff:     buff.ID,
		})
	}
	for _, trap := range Traps {
		all = append(all, Item{
			ID:             trap.ItemID(),
			Name:           trap.ItemName(),
			Kind:           ItemTrap,
			Classification: TrapClassification,
			Trap:           trap.ID,
		})
	}
	/* Useful and never progression

	A wave has to stay winnable without one, so no access rule may ever need
	a setting: a seed that puts the hook behind a check nobody can reach
	would still be beatable. */
	for _, setting := range ServerSettings {
		all = append(all, Item{
			ID:             setting.ItemID(),
			Name:           setting.ItemName(),
			Kind:           ItemServerSetting,
			Classification: Useful,
			Count:          1,
			ServerSetting:  setting.ID,
		})
	}
	return append(all, trophyItems()...)
}

/*
trophyItems is a medal per mission, locked onto that mission's clear.

Never in the pool: the world places one on each clear when the option is on and
adds none otherwise, so a seed without the option never sees these ids.
Progression because both goals read them.
*/
func trophyItems() []Item {
	all := make([]Item, 0, len(Missions))
	for _, m := range Missions {
		all = append(all, Item{
			ID:             m.TrophyItemID(),
			Name:           m.TrophyItemName(),
			Kind:           ItemTrophy,
			Classification: Progression,
			Count:          1,
			Mission:        m.ID,
		})
	}
	return all
}
