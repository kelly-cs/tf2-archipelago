package gamedata

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestWeaponBuffCatalogIsEveryWeaponEffectPermutation(t *testing.T) {
	if len(Weapons) < 212 {
		t.Fatalf("catalog has %d weapons, want at least 212", len(Weapons))
	}
	if got, want := len(WeaponBuffs), len(Weapons)*len(WeaponEffects); got != want {
		t.Fatalf("catalog has %d permutations, want %d", got, want)
	}
	var definitions []int
	for _, weapon := range Weapons {
		if weapon.ID == 0 || weapon.Key == "" || weapon.Name == "" {
			t.Errorf("incomplete weapon: %+v", weapon)
		}
		if len(weapon.DefIndexes) == 0 {
			t.Errorf("%s has no item definition", weapon.Name)
		}
		for _, definition := range weapon.DefIndexes {
			if slices.Contains(definitions, definition) {
				t.Errorf("item definition %d belongs to more than one weapon", definition)
			}
			definitions = append(definitions, definition)
		}
	}
	seen := make(map[[2]uint16]bool, len(WeaponBuffs))
	for _, buff := range WeaponBuffs {
		if buff.ID == 0 || buff.Key == "" || buff.Weapon == "" || buff.Attribute == "" || buff.Description == "" {
			t.Errorf("incomplete weapon buff: %+v", buff)
		}
		pair := [2]uint16{buff.WeaponID, uint16(buff.EffectID)}
		if seen[pair] {
			t.Errorf("duplicate weapon/effect permutation %v", pair)
		}
		seen[pair] = true
	}
	if len(definitions) < 400 {
		t.Fatalf("catalog covers %d concrete item definitions, want at least 400", len(definitions))
	}
}

func TestReserveShooterHasAConcreteBuff(t *testing.T) {
	for _, weapon := range Weapons {
		if weapon.Name == "Reserve Shooter" {
			if !slices.Contains(weapon.DefIndexes, 415) {
				t.Fatalf("Reserve Shooter definitions = %v, want 415", weapon.DefIndexes)
			}
			return
		}
	}
	t.Fatal("Reserve Shooter is absent from the weapon buff catalog")
}

func TestGeneratedPluginCatalogContainsEveryBuffKey(t *testing.T) {
	body, err := os.ReadFile("../plugin/scripting/tf2_archipelago/weapon_buffs_data.inc")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, buff := range WeaponBuffs {
		if !strings.Contains(text, `"`+buff.Key+`"`) {
			t.Errorf("generated plugin catalog has no key %q", buff.Key)
		}
	}
}

func TestEveryRequestedSillyEffectIsAvailableForEveryWeapon(t *testing.T) {
	wanted := map[string]BuffMode{
		"projectile-count": BuffPercentage,
		"projectile-speed": BuffPercentage,
		"bleed":            BuffAdd,
		"afterburn-damage": BuffPercentage,
		"airborne-crits":   BuffToggle,
		"ignite":           BuffToggle,
		"gasoline":         BuffToggle,
		"mad-milk":         BuffToggle,
		"no-self-blast":    BuffToggle,
		"heal-on-kill":     BuffAdd,
		"slow-on-hit":      BuffToggle,
		"gesture-speed":    BuffPercentage,
	}
	for _, effect := range WeaponEffects {
		if mode, ok := wanted[effect.Key]; ok {
			if effect.Mode != mode {
				t.Errorf("%s mode = %d, want %d", effect.Key, effect.Mode, mode)
			}
			delete(wanted, effect.Key)
		}
	}
	if len(wanted) != 0 {
		t.Fatalf("missing requested effects: %v", wanted)
	}
}

func TestRequestedSecondPassEffectValues(t *testing.T) {
	wanted := map[string]float32{
		"heal-on-kill":  15,
		"no-self-blast": 1,
		"slow-on-hit":   1,
		"gesture-speed": 0.50,
	}
	for _, effect := range WeaponEffects {
		if value, ok := wanted[effect.Key]; ok {
			if effect.Increment != value {
				t.Errorf("%s increment = %.2f, want %.2f", effect.Key, effect.Increment, value)
			}
			delete(wanted, effect.Key)
		}
	}
	if len(wanted) != 0 {
		t.Fatalf("missing requested second-pass effects: %v", wanted)
	}
}

func TestWeaponEffectsUseDistinctSchemaAttributes(t *testing.T) {
	seen := make(map[string]string, len(WeaponEffects))
	for _, effect := range WeaponEffects {
		if previous, duplicate := seen[effect.Attribute]; duplicate {
			t.Errorf("effects %q and %q both use schema attribute %q",
				previous, effect.Key, effect.Attribute)
		}
		seen[effect.Attribute] = effect.Key
	}
	if got := seen["bullets per shot bonus"]; got != "projectile-count" {
		t.Errorf("projectile-count schema mapping = %q, want bullets per shot bonus", got)
	}
}

func TestLegacyPermutationKeepsItsIDAndEveryEffectExists(t *testing.T) {
	for _, old := range legacyWeaponBuffs {
		seenEffects := make(map[uint8]bool, len(WeaponEffects))
		keptLegacy := false
		for _, buff := range WeaponBuffs {
			if buff.WeaponID != old.ID {
				continue
			}
			seenEffects[buff.EffectID] = true
			if buff.Attribute == old.Attribute {
				keptLegacy = buff.ID == old.ID && buff.Key == old.Key
			}
		}
		if !keptLegacy {
			t.Errorf("%s did not retain legacy id %d and key %q", old.Weapon, old.ID, old.Key)
		}
		if len(seenEffects) != len(WeaponEffects) {
			t.Errorf("%s has %d effects, want %d", old.Weapon, len(seenEffects), len(WeaponEffects))
		}
	}
}

func TestItemExportMarksOnlyNumericBuffsStackable(t *testing.T) {
	for _, item := range buildItemsFile().Items {
		if item.WeaponBuffID == 0 {
			if item.Stackable {
				t.Errorf("non-buff item %q is stackable", item.Name)
			}
			continue
		}
		buff, ok := WeaponBuffByID(item.WeaponBuffID)
		if !ok {
			t.Errorf("item %q has unknown weapon buff %d", item.Name, item.WeaponBuffID)
			continue
		}
		want := buff.Mode != BuffToggle
		if item.Stackable != want {
			t.Errorf("%s stackable = %t, want %t for mode %d",
				buff.Key, item.Stackable, want, buff.Mode)
		}
	}
}
