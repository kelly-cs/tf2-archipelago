package gamedata

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestWeaponBuffCatalogCoversTheFunctionalWeapons(t *testing.T) {
	if len(WeaponBuffs) < 212 {
		t.Fatalf("catalog has %d weapon buffs, want at least 212", len(WeaponBuffs))
	}
	var definitions []int
	for _, buff := range WeaponBuffs {
		if buff.ID == 0 || buff.Key == "" || buff.Weapon == "" || buff.Attribute == "" || buff.Description == "" {
			t.Errorf("incomplete weapon buff: %+v", buff)
		}
		if len(buff.DefIndexes) == 0 {
			t.Errorf("%s has no item definition", buff.Weapon)
		}
		for _, definition := range buff.DefIndexes {
			if slices.Contains(definitions, definition) {
				t.Errorf("item definition %d belongs to more than one buff", definition)
			}
			definitions = append(definitions, definition)
		}
	}
	if len(definitions) < 400 {
		t.Fatalf("catalog covers %d concrete item definitions, want at least 400", len(definitions))
	}
}

func TestReserveShooterHasAConcreteBuff(t *testing.T) {
	for _, buff := range WeaponBuffs {
		if buff.Weapon == "Reserve Shooter" {
			if !slices.Contains(buff.DefIndexes, 415) {
				t.Fatalf("Reserve Shooter definitions = %v, want 415", buff.DefIndexes)
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
