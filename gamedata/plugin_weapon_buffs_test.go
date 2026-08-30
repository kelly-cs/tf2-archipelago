package gamedata

import (
	"os"
	"strings"
	"testing"
)

func sourceFunction(t *testing.T, path, signature string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	start := strings.Index(text, signature)
	if start < 0 {
		t.Fatalf("%s has no %s", path, signature)
	}
	open := strings.IndexByte(text[start:], '{')
	if open < 0 {
		t.Fatalf("%s has no body for %s", path, signature)
	}
	open += start
	depth := 0
	for index := open; index < len(text); index++ {
		switch text[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : index+1]
			}
		}
	}
	t.Fatalf("%s has an unterminated body for %s", path, signature)
	return ""
}

func TestWeaponBuffsStayOutOfMvMShopping(t *testing.T) {
	bots := "../plugin/scripting/tf2_archipelago/bots.inc"
	before := sourceFunction(t, bots, "public Action OnClientCommandKeyValues(int client")
	if !strings.Contains(before, "WeaponBuffs_BeginStationTransaction(client)") {
		t.Fatal("station command does not remove Archipelago attributes before TF2 handles it")
	}

	station := sourceFunction(t, "../plugin/scripting/tf2_archipelago/mvm.inc",
		"stock void MvM_OnCommandKeyValues")
	for _, command := range []string{"MVM_Upgrade", "MVM_Respec"} {
		if !strings.Contains(station, command) {
			t.Fatalf("upgrade-station handler no longer handles %s", command)
		}
	}
	if !strings.Contains(station, "WeaponBuffs_ApplyNextFrame(client)") {
		t.Fatal("upgrade-station handler does not restore Archipelago attributes after TF2 handles it")
	}

	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	transaction := sourceFunction(t, buffs, "void WeaponBuffs_BeginStationTransaction")
	for _, required := range []string{"WeaponBuffs_Remove(client)", "WeaponBuffs_ApplyNextFrame(client)"} {
		if !strings.Contains(transaction, required) {
			t.Fatalf("station transaction has no %s", required)
		}
	}
	apply := sourceFunction(t, buffs, "void WeaponBuffs_Apply(int client)")
	if strings.Contains(apply, "g_WeaponBuffWaveActive") {
		t.Fatal("weapon buffs are incorrectly gated on an active wave")
	}
	applyEntity := sourceFunction(t, buffs, "static void WeaponBuffs_ApplyEntity")
	if !strings.Contains(applyEntity, "effect == FireRateEffect") {
		t.Fatal("fire rate is still written into the MvM station's runtime attribute list")
	}
	fireRate := sourceFunction(t, buffs, "public void WeaponBuffs_PostThinkPost")
	for _, required := range []string{
		"m_flNextPrimaryAttack", "g_WeaponEffectLevels[weapon][FireRateEffect]",
		"TF2Attrib_GetByName", "combined / station",
	} {
		if !strings.Contains(fireRate, required) {
			t.Fatalf("plugin-owned fire rate has no %s", required)
		}
	}

	plugin := "../plugin/scripting/tf2_archipelago.sp"
	begin := sourceFunction(t, plugin, "public void Event_BeginWave")
	if !strings.Contains(begin, "WeaponBuffs_BeginWave()") {
		t.Fatal("wave start does not apply Archipelago attributes")
	}
	for _, signature := range []string{
		"public Action Command_TestWeaponBuff", "public Action Command_GiveWeaponBuff",
	} {
		body := sourceFunction(t, buffs, signature)
		if strings.Contains(body, "WeaponBuffs_ApplyEntity") {
			t.Fatalf("%s bypasses the shopping-period guard", signature)
		}
		if !strings.Contains(body, "WeaponBuffs_Apply(") {
			t.Fatalf("%s does not apply through the guarded client path", signature)
		}
	}
}
