package gamedata

//go:generate go run ./cmd/pluginbuffs ../plugin/scripting/tf2_archipelago/weapon_buffs_data.inc

// WeaponBuff is one useful Archipelago item and the attribute it adds while
// that weapon is equipped. DefIndexes groups stock definitions shared by
// classes and named definitions that resolve to the same inventory weapon.
type WeaponBuff struct {
	ID          uint16
	Key         string
	Weapon      string
	DefIndexes  []int
	Attribute   string
	Value       float32
	Description string
	Additive    bool
}

func (b WeaponBuff) ItemID() int64 {
	return BaseID + itemSpaceOffset + itemBlockWeaponBuff + int64(b.ID)
}

func (b WeaponBuff) ItemName() string {
	return "Weapon Buff: " + b.Weapon + " — " + b.Description
}

var weaponBuffsByID = indexWeaponBuffs()

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
