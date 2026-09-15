package gamedata

import "testing"

func buffFor(t *testing.T, weapon, key string) WeaponBuff {
	t.Helper()
	for _, buff := range WeaponBuffs {
		if buff.Weapon == weapon && buff.EffectID == effectByKey(t, key).ID && buff.WeaponID == buff.ApplyWeaponID {
			return buff
		}
	}
	t.Fatalf("no buff pairs %s with %s", weapon, key)
	return WeaponBuff{}
}

func effectByKey(t *testing.T, key string) WeaponEffect {
	t.Helper()
	for _, effect := range WeaponEffects {
		if effect.Key == key {
			return effect
		}
	}
	t.Fatalf("no effect %q", key)
	return WeaponEffect{}
}

// The three the report named, in the report's words.
func TestAFlagTheWeaponAlreadyCarriesIsNotABuff(t *testing.T) {
	for _, tc := range []struct{ weapon, key string }{
		{"Righteous Bison", "projectile-penetration"},
		{"Rocket Jumper", "no-self-blast"},
		{"Eviction Notice", "speed-on-hit"},
	} {
		if buffFor(t, tc.weapon, tc.key).Eligible {
			t.Errorf("%s already has %s and still draws it", tc.weapon, tc.key)
		}
	}
}

// The rule the table encodes: every row is a real weapon, every cell a flag
// rather than a number, and none of them is in the pool. A number that the
// weapon already has stacks, and a row for one would be a mistake.
func TestNativeCutsAreFlagsOnRealWeaponsAndOutOfThePool(t *testing.T) {
	for weapon, keys := range nativeCuts {
		for key := range keys {
			buff := buffFor(t, weapon, key)
			if buff.Mode != BuffToggle {
				t.Errorf("%s / %s is a number, not a flag: it stacks and belongs in the pool", weapon, key)
			}
			if buff.Eligible {
				t.Errorf("%s still draws %s", weapon, key)
			}
		}
	}
}

// A step up is not a repeat: mini-crits on a burning target leave full crits
// worth drawing.
func TestMiniCritOnBurningKeepsTheFullCritBuff(t *testing.T) {
	for _, weapon := range []string{"Axtinguisher", "Detonator", "Scorch Shot"} {
		if !buffFor(t, weapon, "crits-vs-burning").Eligible {
			t.Errorf("%s only mini-crits a burning target and lost the full-crit buff", weapon)
		}
	}
}
