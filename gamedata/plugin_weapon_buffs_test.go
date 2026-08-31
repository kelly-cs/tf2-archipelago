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
	if !strings.Contains(station, "WeaponBuffs_ApplyNextFrame(client)") {
		t.Fatal("upgrade-station handler does not recalculate the independent provider")
	}

	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	source, err := os.ReadFile(buffs)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "WeaponBuffs_BeginStationTransaction") {
		t.Fatal("obsolete station transaction still couples Archipelago effects to shopping")
	}
	provider := sourceFunction(t, buffs, "static int WeaponBuffs_Provider")
	for _, required := range []string{
		`CreateEntityByName("tf_wearable")`, "RENDER_NONE",
		`SetEntProp(provider, Prop_Send, "m_ProviderType", 0)`,
		`SetEntPropEnt(provider, Prop_Send, "m_hOwnerEntity", client)`,
		`AcceptEntityInput(provider, "RunScriptCode")`,
		`TF2Attrib_HookValueFloat(1.0, "mult_dmg", weaponEntity)`,
		"TF2Attrib_ClearCache(weaponEntity)",
		"g_WeaponBuffProviderRef",
	} {
		if !strings.Contains(provider, required) {
			t.Fatalf("private attribute provider has no %s", required)
		}
	}
	apply := sourceFunction(t, buffs, "void WeaponBuffs_Apply(int client)")
	for _, required := range []string{
		"TF2Attrib_RemoveAll(provider)",
		"TF2Attrib_ClearCache(entity)",
		"TF2Attrib_HookValueFloat",
		"TF2Attrib_SetByName(provider",
	} {
		if !strings.Contains(apply, required) {
			t.Fatalf("independent attribute composition has no %s", required)
		}
	}
	for _, forbidden := range []string{
		"TF2Attrib_SetByName(entity", "GetPlayerWeaponSlot", "g_WeaponBuffWaveActive",
	} {
		if strings.Contains(apply, forbidden) {
			t.Fatalf("Archipelago apply path still uses station-owned state: %s", forbidden)
		}
	}
	if strings.Contains(string(source), "TF2Econ_") || strings.Contains(string(source), "TF2Util_") {
		t.Fatal("private provider still depends on update-sensitive TF Econ Data or TF2 Utils natives")
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

func TestWeaponBuffOwnershipPolicyDefaultsToPlayers(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	source, err := os.ReadFile(buffs)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		`CreateConVar("tf2ap_mirror_buffs_to_robots", "0"`,
		"g_WeaponBuffMirrorRobots.AddChangeHook",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("robot mirroring toggle has no %s", required)
		}
	}

	policy := sourceFunction(t, buffs, "static bool WeaponBuffs_CanUseRunBuffs(int client)")
	for _, required := range []string{
		"MvM_IsPlayer(client)",
		"g_WeaponBuffMirrorRobots.BoolValue",
		"WeaponBuffs_IsEnemyRobot(client)",
	} {
		if !strings.Contains(policy, required) {
			t.Fatalf("weapon-buff ownership policy has no %s", required)
		}
	}

	for _, signature := range []string{
		"void WeaponBuffs_Apply(int client)",
		"public void WeaponBuffs_OnTakeDamagePost",
		"static void WeaponBuffs_FireExtraProjectiles(any reference)",
		"public void WeaponBuffs_SubstanceProjectileSpawned(int entity)",
	} {
		body := sourceFunction(t, buffs, signature)
		if !strings.Contains(body, "WeaponBuffs_CanUseRunBuffs(") {
			t.Fatalf("%s bypasses the run-buff ownership policy", signature)
		}
	}

	changed := sourceFunction(t, buffs, "public void WeaponBuffs_MirrorRobotsChanged")
	for _, required := range []string{"WeaponBuffs_Apply(client)", "WeaponBuffs_Remove(client)"} {
		if !strings.Contains(changed, required) {
			t.Fatalf("live robot mirror toggle has no %s", required)
		}
	}
}
