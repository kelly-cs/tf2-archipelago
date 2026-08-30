package gamedata

import "slices"

//go:generate go run ./cmd/pluginbuffs ../plugin/scripting/tf2_archipelago/weapon_buffs_data.inc

// BuffMode describes how repeated copies compose with the weapon's existing
// attribute (including values bought at an MvM upgrade station).
type BuffMode uint8

const (
	BuffAdd BuffMode = iota + 1
	BuffPercentage
	BuffToggle
)

// BuffWeapon is one inventory weapon and every concrete definition that resolves
// to it. ID is append-only because it participates in item IDs.
type BuffWeapon struct {
	ID         uint16
	Key        string
	Name       string
	DefIndexes []int
	// ApplyID is the canonical member of a functional reskin family. Rewards
	// for any historical family member accumulate on this weapon index.
	ApplyID uint16
}

// WeaponEffect is one positive TF2 item attribute. Every effect is deliberately
// paired with every weapon, including combinations the stock game never uses.
type WeaponEffect struct {
	ID          uint8
	Key         string
	Attribute   string
	Increment   float32
	Mode        BuffMode
	Description string
}

// WeaponBuff is one weapon/effect permutation exposed as an Archipelago item.
type WeaponBuff struct {
	ID            uint16
	Key           string
	WeaponID      uint16
	ApplyWeaponID uint16
	EffectID      uint8
	Weapon        string
	DefIndexes    []int
	Attribute     string
	Value         float32
	Description   string
	Additive      bool
	Mode          BuffMode
	Eligible      bool
}

func (b WeaponBuff) ItemID() int64 {
	return BaseID + itemSpaceOffset + itemBlockWeaponBuff + int64(b.ID)
}

func (b WeaponBuff) ItemName() string {
	return "Weapon Buff: " + b.Weapon + " — " + b.Description
}

var weaponBuffsByID = indexWeaponBuffs()

// BuffWeapons keeps the stable weapon roster separate from the effects so the
// SourcePawn tables do not repeat 400 definition indexes for every effect.
var BuffWeapons = buildBuffWeapons()

// WeaponBuffs is every possible weapon/effect permutation. The legacy entry
// for each weapon retains its old ID and key; all other IDs live in fixed
// effect blocks, so adding an effect or weapon cannot renumber an old item.
var WeaponBuffs = buildWeaponBuffs()

func buildBuffWeapons() []BuffWeapon {
	weapons := make([]BuffWeapon, 0, len(legacyWeaponBuffs))
	for _, old := range legacyWeaponBuffs {
		weapons = append(weapons, BuffWeapon{
			ID: old.ID, Key: old.Key, Name: old.Weapon, DefIndexes: slices.Clone(old.DefIndexes),
			ApplyID: old.ID,
		})
	}
	replaceDecoratedWeaponDefinitions(weapons)
	mergeWeaponFamilies(weapons)
	return weapons
}

// The original catalog treated a 150xx decorated definition as though its
// suffix were the stock definition. Valve assigns those numbers sequentially
// across collections instead: 15015 is a Scattergun, not Minigun 15, while
// Brain Candy Minigun is 15098. Replace that inferred data with the prefab
// mapping audited from the item schema.
func replaceDecoratedWeaponDefinitions(weapons []BuffWeapon) {
	byName := make(map[string]int, len(weapons))
	for index := range weapons {
		byName[weapons[index].Name] = index
		weapons[index].DefIndexes = slices.DeleteFunc(weapons[index].DefIndexes, func(definition int) bool {
			return definition >= 15000 && definition < 16000
		})
	}
	for name, definitions := range decoratedWeaponDefinitions {
		index, ok := byName[name]
		if !ok {
			panic("decorated weapon has no catalog member: " + name)
		}
		weapons[index].DefIndexes = append(weapons[index].DefIndexes, definitions...)
		slices.Sort(weapons[index].DefIndexes)
	}
}

func buildWeaponBuffs() []WeaponBuff {
	all := make([]WeaponBuff, 0, len(BuffWeapons)*len(WeaponEffects))
	for _, weapon := range BuffWeapons {
		legacy := legacyWeaponBuffs[weapon.ID-1]
		canonical := BuffWeapons[weapon.ApplyID-1]
		for _, effect := range WeaponEffects {
			id := uint16(10000 + int(effect.ID)*256 + int(weapon.ID))
			key := weapon.Key + "-" + effect.Key
			if effect.Attribute == legacy.Attribute {
				id, key = legacy.ID, legacy.Key
			}
			all = append(all, WeaponBuff{
				ID: id, Key: key, WeaponID: weapon.ID, ApplyWeaponID: weapon.ApplyID,
				EffectID: effect.ID,
				Weapon:   weapon.Name, DefIndexes: weapon.DefIndexes,
				Attribute: effect.Attribute, Value: effect.Increment,
				Description: effect.Description, Additive: effect.Mode == BuffAdd,
				Mode:     effect.Mode,
				Eligible: weapon.ID == weapon.ApplyID && weaponEffectEligible(canonical, effect),
			})
		}
	}
	return all
}

func indexWeaponBuffs() map[uint16]WeaponBuff {
	byID := make(map[uint16]WeaponBuff, len(WeaponBuffs))
	for _, buff := range WeaponBuffs {
		byID[buff.ID] = buff
	}
	return byID
}

func WeaponBuffByID(id uint16) (WeaponBuff, bool) {
	buff, ok := weaponBuffsByID[id]
	return buff, ok
}
