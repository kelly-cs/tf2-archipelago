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

func TestWeaponBuffSubstancesUseTheSharedHitPath(t *testing.T) {
	path := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	apply := sourceFunction(t, path, "static void WeaponBuffs_ApplyHitEffects")
	for _, effect := range []string{
		"TF2_MakeBleed", "TF2_IgnitePlayer", "TFCond_Milked", "TFCond_Gas",
		"TFCond_MarkedForDeath", "TFCond_Jarated",
	} {
		if !strings.Contains(apply, effect) {
			t.Fatalf("shared hit path does not apply %s", effect)
		}
	}

	jar := sourceFunction(t, path, "public void WeaponBuffs_ApplySubstances")
	if !strings.Contains(jar, "WeaponBuffs_ApplyHitEffects(victim, attacker, weapon)") {
		t.Fatal("jar splashes bypass the shared hit-effect path")
	}

	damage := sourceFunction(t, path, "public void WeaponBuffs_OnTakeDamagePost")
	for _, guard := range []string{
		"GetClientTeam(victim) == GetClientTeam(attacker)",
		"damagecustom == TF_CUSTOM_BURNING",
		"damagecustom == TF_CUSTOM_BLEEDING",
		"WeaponBuffs_IsSubstanceProjectile(inflictorClass)",
	} {
		if !strings.Contains(damage, guard) {
			t.Fatalf("direct-hit path lost guard %q", guard)
		}
	}
	if !strings.Contains(damage, "WeaponBuffs_ForEntity(weapon)") ||
		!strings.Contains(damage, "WeaponBuffs_ApplyHitEffects(victim, attacker, catalog)") {
		t.Fatal("direct hits do not resolve the canonical weapon and apply its hit effects")
	}
}
