// Package botcards holds the named defender prototypes available to a test run.
// A card is an identity, not a bot client: its seat is rebuilt after map changes.
package botcards

type Card struct {
	ID            string
	Name          string
	Class         string
	Loadout       string
	Tier          Tier
	Cosmetic      int // TF2 item definition index from the installed item schema.
	UnusualEffect int // Fixed cosmetic particle ID; zero for non-Legendary cards.
	// BuffIDs identify normal eligible Archipelago effects through an item
	// the class can equip. The card grants each distinct effect to every
	// weapon it carries, once per tier stack.
	BuffIDs []uint16
}

type Tier string

const (
	Common    Tier = "common"
	Elite     Tier = "elite"
	Legendary Tier = "legendary"
)

func (tier Tier) Stacks() int {
	switch tier {
	case Elite:
		return 2
	case Legendary:
		return 3
	default:
		return 1
	}
}

// Cards use names from deploy/bots/bot_names.txt and loadouts the defender mod
// already knows. They are deliberately curated while the AP item shape is tested.
var Cards = []Card{
	{ID: "credit-to-team", Name: "CreditToTeam", Class: "scout", Loadout: "milk", Tier: Common, Cosmetic: 111,
		BuffIDs: []uint16{10434}}, // Soda Popper damage
	{ID: "screamin-eagles", Name: "Screamin' Eagles", Class: "soldier", Loadout: "beggar", Tier: Elite, Cosmetic: 378,
		BuffIDs: []uint16{10275, 10577}}, // Beggar damage, Escape Plan firing speed
	{ID: "ivan", Name: "IvanTheSpaceBiker", Class: "heavyweapons", Loadout: "brass", Tier: Elite, Cosmetic: 185,
		BuffIDs: []uint16{10542, 11093}}, // Brass Beast firing speed, Family Business clip size
	{ID: "herr-doktor", Name: "Herr Doktor", Class: "medic", Loadout: "kritz", Tier: Legendary, Cosmetic: 315, UnusualEffect: 13, // Burning Flames
		BuffIDs: []uint16{10305, 17531, 10718}}, // Crossbow damage, Kritz Über rate, Übersaw firing speed
	{ID: "chell", Name: "Chell", Class: "engineer", Loadout: "ranger", Tier: Common, Cosmetic: 484,
		BuffIDs: []uint16{10406}}, // Rescue Ranger damage
	{ID: "mentlegen", Name: "Mentlegen", Class: "spy", Loadout: "diamondback", Tier: Legendary, Cosmetic: 55, UnusualEffect: 14, // Scorching Flames
		BuffIDs: []uint16{10313, 10532, 18212}}, // Diamondback damage, Big Earner firing speed and armor
}

func ByID(id string) (Card, bool) {
	for _, card := range Cards {
		if card.ID == id {
			return card, true
		}
	}
	return Card{}, false
}

func BySeat(class, name string) (Card, bool) {
	for _, card := range Cards {
		if card.Class == class && card.Name == name {
			return card, true
		}
	}
	return Card{}, false
}
