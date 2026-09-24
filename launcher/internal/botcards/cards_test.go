package botcards

import (
	"slices"
	"testing"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/botloadout"
)

func TestInnatesAreDistinctStackableAPEffectsForTheClass(t *testing.T) {
	for _, card := range Cards {
		t.Run(card.Name, func(t *testing.T) {
			class, ok := botloadout.ClassByKey(card.Class)
			if !ok {
				t.Fatalf("unknown class %q", card.Class)
			}
			loadout := class.LoadoutByKey(card.Loadout)
			definitions := []int{loadout.Primary, loadout.Second, loadout.Melee, loadout.PDA2}
			if len(card.BuffIDs) != card.Tier.Stacks() {
				t.Fatalf("%s has %d innates, want %d", card.Tier, len(card.BuffIDs), card.Tier.Stacks())
			}
			seen := map[int]bool{}
			for _, id := range card.BuffIDs {
				buff, found := gamedata.WeaponBuffByID(id)
				if !found || !buff.Eligible || buff.Mode == gamedata.BuffToggle {
					t.Errorf("innate %d is not an eligible stackable AP weapon buff", id)
					continue
				}
				if seen[int(buff.EffectID)] {
					t.Errorf("duplicate innate effect %d", buff.EffectID)
				}
				seen[int(buff.EffectID)] = true
				if !slices.ContainsFunc(definitions, func(definition int) bool {
					return slices.Contains(buff.DefIndexes, definition)
				}) {
					t.Errorf("%s cannot equip %s", card.Class, buff.ItemName())
				}
			}
		})
	}
}

func TestLegendaryUnusualEffectsBelongToCards(t *testing.T) {
	seen := map[int]bool{}
	for _, card := range Cards {
		if card.Tier != Legendary {
			if card.UnusualEffect != 0 {
				t.Errorf("%s is not Legendary but has effect %d", card.ID, card.UnusualEffect)
			}
			continue
		}
		if card.UnusualEffect <= 0 || seen[card.UnusualEffect] {
			t.Errorf("%s needs its own fixed Unusual effect, got %d", card.ID, card.UnusualEffect)
		}
		seen[card.UnusualEffect] = true
	}
}
