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

func TestCrossbowDiagnosticsPreserveNativeHealing(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	init := sourceFunction(t, buffs, "void WeaponBuffs_Init()")
	if !strings.Contains(init, `HookEventEx("crossbow_heal", WeaponBuffs_CrossbowHeal)`) {
		t.Fatal("crossbow diagnostics do not observe TF2's authoritative heal event")
	}

	created := sourceFunction(t, buffs, "public void OnEntityCreated")
	for _, wanted := range []string{"tf_projectile_healing_bolt", "WeaponBuffs_ArrowStartTouch"} {
		if !strings.Contains(created, wanted) {
			t.Fatalf("crossbow projectiles are not hooked for %s diagnostics", wanted)
		}
	}

	touch := sourceFunction(t, buffs, "public Action WeaponBuffs_ArrowStartTouch")
	for _, wanted := range []string{"m_iProjectileType", "m_CollisionGroup", "projectile penetration", "g_WeaponBuffArrowExtra"} {
		if !strings.Contains(touch, wanted) {
			t.Fatalf("crossbow touch diagnostics omit %s", wanted)
		}
	}
	for _, replacement := range []string{"SetEntityHealth", "TF2Util_TakeHealth"} {
		if strings.Contains(touch, replacement) {
			t.Fatalf("diagnostic hook replaces TF2's native healing through %s", replacement)
		}
	}

	fanout := sourceFunction(t, buffs, "static int WeaponBuffs_FireNativeArrow")
	for _, wanted := range []string{"g_WeaponBuffArrowCreate", "m_hLauncher", "WeaponBuffs_ArrowStartTouch"} {
		if !strings.Contains(fanout, wanted) {
			t.Fatalf("fan-out bolt omits native state needed for healing: %s", wanted)
		}
	}
}
