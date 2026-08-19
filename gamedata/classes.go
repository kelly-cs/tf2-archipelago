package gamedata

// ClassID identifies one mercenary class. Explicit literals, append-only,
// never reused: these ids reach seeds through the class items.
type ClassID uint8

const (
	ClassScout    ClassID = 1
	ClassSoldier  ClassID = 2
	ClassPyro     ClassID = 3
	ClassDemoman  ClassID = 4
	ClassHeavy    ClassID = 5
	ClassEngineer ClassID = 6
	ClassMedic    ClassID = 7
	ClassSniper   ClassID = 8
	ClassSpy      ClassID = 9
)

// Class is one mercenary class. Key is what crosses the wire to the plugin,
// which maps it to TF2's own class enum. SlotOrder is the order the progressive
// weapon slot opens this class's loadout in: copy n of the item opens the n-th
// slot listed.
//
// Six classes take the order of WeaponSlots itself. Three cannot. The class a
// Medic plays is the Medigun, an Engineer's is the Wrench and the PDAs it
// carries, and a Spy's is the Knife. Give one of them the primary slot first
// and the class does nothing, which is what a run that draws it as its only
// starting class found.
//
// The generator never sees this order. Every class holds the same number of
// slots whatever the order is, so no access rule changes and the export does
// not carry it. The plugin holds the same table, and gamedata's plugin key
// test compares the two.
type Class struct {
	ID        ClassID
	Key       string
	Name      string
	SlotOrder []WeaponSlotID
}

var (
	slotOrderDefault = []WeaponSlotID{WeaponSlotPrimary, WeaponSlotSecondary, WeaponSlotMelee}
	slotOrderMedic   = []WeaponSlotID{WeaponSlotSecondary, WeaponSlotPrimary, WeaponSlotMelee}
	slotOrderMelee   = []WeaponSlotID{WeaponSlotMelee, WeaponSlotSecondary, WeaponSlotPrimary}
)

// Classes is the nine classes, in the order the class selection menu lists them.
var Classes = []Class{
	{ClassScout, "scout", "Scout", slotOrderDefault},
	{ClassSoldier, "soldier", "Soldier", slotOrderDefault},
	{ClassPyro, "pyro", "Pyro", slotOrderDefault},
	{ClassDemoman, "demoman", "Demoman", slotOrderDefault},
	{ClassHeavy, "heavy", "Heavy", slotOrderDefault},
	// The Wrench first, then the Shotgun: an Engineer holds his ground with a
	// sentry, and the Pistol is what he needs least.
	{ClassEngineer, "engineer", "Engineer", []WeaponSlotID{WeaponSlotMelee, WeaponSlotPrimary, WeaponSlotSecondary}},
	{ClassMedic, "medic", "Medic", slotOrderMedic},
	{ClassSniper, "sniper", "Sniper", slotOrderDefault},
	// The Knife, then the Sapper: a sapped group of robots is worth more in
	// Mann vs Machine than the Revolver is.
	{ClassSpy, "spy", "Spy", slotOrderMelee},
}

var classesByID = indexClasses()

func indexClasses() map[ClassID]Class {
	byID := make(map[ClassID]Class, len(Classes))
	for _, c := range Classes {
		byID[c.ID] = c
	}
	return byID
}

// ClassByID returns the class with that id.
func ClassByID(id ClassID) (Class, bool) {
	c, ok := classesByID[id]
	return c, ok
}
