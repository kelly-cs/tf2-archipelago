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
	station := sourceFunction(t, "../plugin/scripting/tf2_archipelago/mvm.inc",
		"stock void MvM_OnCommandKeyValues")
	for _, command := range []string{"MVM_Upgrade", "MVM_Respec"} {
		if !strings.Contains(station, command) {
			t.Fatalf("upgrade-station handler no longer handles %s", command)
		}
	}
	if strings.Contains(station, "WeaponBuffs_Apply") {
		t.Fatal("upgrade-station handler reapplies Archipelago attributes while shopping")
	}

	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	apply := sourceFunction(t, buffs, "void WeaponBuffs_Apply(int client)")
	if !strings.Contains(apply, "!g_WeaponBuffWaveActive") {
		t.Fatal("weapon buffs can be applied outside an active wave")
	}
	end := sourceFunction(t, buffs, "void WeaponBuffs_EndWave()")
	if !strings.Contains(end, "WeaponBuffs_Remove(client)") {
		t.Fatal("ending a wave does not remove Archipelago attributes")
	}

	plugin := "../plugin/scripting/tf2_archipelago.sp"
	begin := sourceFunction(t, plugin, "public void Event_BeginWave")
	if !strings.Contains(begin, "WeaponBuffs_BeginWave()") {
		t.Fatal("wave start does not apply Archipelago attributes")
	}
	for _, signature := range []string{
		"static void ReportWaveFailed", "static void ReportWaveCleared",
		"static void ReportMissionCleared",
	} {
		if body := sourceFunction(t, plugin, signature); !strings.Contains(body, "WeaponBuffs_EndWave()") {
			t.Fatalf("%s does not remove Archipelago attributes", signature)
		}
	}
}

func TestMeleeKillOverhealComposesAfterNativeHealing(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	death := sourceFunction(t, buffs, "void WeaponBuffs_PlayerDeath")
	for _, wanted := range []string{
		"TF2Attrib_GetByName(entity, \"heal on kill\")",
		"MeleeKillOverhealEffect",
		"g_WeaponBuffLastDamageAttackerHealth[victim]",
		"RequestFrame(WeaponBuffs_ApplyKillOverheal",
	} {
		if !strings.Contains(death, wanted) {
			t.Fatalf("melee kill overheal setup does not contain %q", wanted)
		}
	}
	apply := sourceFunction(t, buffs, "public void WeaponBuffs_ApplyKillOverheal")
	for _, wanted := range []string{"SetEntityHealth", "health < target"} {
		if !strings.Contains(apply, wanted) {
			t.Fatalf("melee kill overheal application does not contain %q", wanted)
		}
	}
	plugin := sourceFunction(t, "../plugin/scripting/tf2_archipelago.sp", "public void Event_PlayerDeath")
	if !strings.Contains(plugin, "WeaponBuffs_PlayerDeath(event)") {
		t.Fatal("player death does not invoke melee kill overheal")
	}
}
