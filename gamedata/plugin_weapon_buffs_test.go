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
	weaponCheck := strings.Index(apply, "if (weapon < 0)")
	providerCreate := strings.Index(apply, "WeaponBuffs_Provider(client, entity)")
	if weaponCheck < 0 || providerCreate < 0 || weaponCheck > providerCreate {
		t.Fatal("buff application creates a provider before validating the active weapon")
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
		"public Action Command_GiveSlotWeaponBuff",
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

func TestAttributeProviderFailureIsRateLimited(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	source, err := os.ReadFile(buffs)
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"#define WeaponBuffProviderRetrySeconds 5.0",
		"g_WeaponBuffProviderRetryAt[MAXPLAYERS + 1]",
		"GetGameTime() < g_WeaponBuffProviderRetryAt[client]",
		"GetGameTime() + WeaponBuffProviderRetrySeconds",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("provider retry guard has no %q", required)
		}
	}
}

func TestSlotBuffCommandFindsWeaponWearables(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	resolver := sourceFunction(t, buffs, "static int WeaponBuffs_EntityInLoadoutSlot")
	for _, required := range []string{
		"GetPlayerWeaponSlot(client, slot)",
		`FindEntityByClassname(entity, "tf_wearable*")`,
		"Unlocks_WearableSlot(class, definition) == slot",
	} {
		if !strings.Contains(resolver, required) {
			t.Errorf("loadout slot resolver has no %q", required)
		}
	}

	command := sourceFunction(t, buffs, "public Action Command_GiveSlotWeaponBuff")
	for _, required := range []string{
		"WeaponBuffs_ParseLoadoutSlot(rawSlot)",
		"WeaponBuffs_EntityInLoadoutSlot(target, slot)",
		"WeaponBuffs_AddTestEffects(weapon, wanted, levels)",
	} {
		if !strings.Contains(command, required) {
			t.Errorf("slot buff command has no %q", required)
		}
	}
}

func TestJoiningClientSnapshotIsNotChurnedByBotRemoval(t *testing.T) {
	plugin := "../plugin/scripting/tf2_archipelago.sp"
	putInServer := sourceFunction(t, plugin, "public void OnClientPutInServer")
	if strings.Contains(putInServer, "Bots_MakeRoom()") {
		t.Fatal("initial client sign-on still removes a bot during the entity snapshot")
	}
	joinRed := sourceFunction(t, plugin, "public Action Command_JoinRed")
	if !strings.Contains(joinRed, "Bots_MakeRoom()") {
		t.Fatal("join-team path no longer frees a bot seat for the loaded player")
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

func TestSelfBlastBuffsPreserveTheNativeExplosionAndPush(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	hook := sourceFunction(t, buffs, "void WeaponBuffs_HookClient(int client)")
	if !strings.Contains(hook, "SDKHook_OnTakeDamage, WeaponBuffs_OnTakeDamage") {
		t.Fatal("players are not hooked for modifiable damage")
	}

	damage := sourceFunction(t, buffs, "public Action WeaponBuffs_OnTakeDamage")
	for _, required := range []string{
		"victim != attacker",
		"!(damagetype & DMG_BLAST)",
		"WeaponBuffs_WeaponOfHit(attacker, inflictor, weapon, damagecustom)",
		"WeaponBuffs_ForEntity(weapon)",
		"g_WeaponEffectLevels[catalog][RocketJumpProtectionEffect]",
		"damage *= kept",
	} {
		if !strings.Contains(damage, required) {
			t.Fatalf("self-blast damage path has no %q", required)
		}
	}
	if strings.Contains(damage, "NoSelfBlastEffect") || strings.Contains(damage, "damage = 0.0") {
		t.Fatal("no-self-blast still zeroes the SDKHook event and suppresses native blast movement")
	}

	native := sourceFunction(t, buffs, "static void WeaponBuffs_SyncNativeSelfBlast")
	for _, required := range []string{
		"GetPlayerWeaponSlot(client, slot)",
		"TF2Attrib_RemoveByName(entity, SelfBlastDamageAttribute)",
		"g_WeaponEffectLevels[weapon][NoSelfBlastEffect]",
		"TF2Attrib_SetByName(entity, SelfBlastDamageAttribute, 0.0)",
		"TF2Attrib_ClearCache(entity)",
	} {
		if !strings.Contains(native, required) {
			t.Fatalf("native no-self-blast path has no %q", required)
		}
	}
	if strings.Contains(native, "g_WeaponEffectAttributes[NoSelfBlastEffect], 2.0") {
		t.Fatal("native no-self-blast path still selects TF2's replacement Jumper explosion")
	}

	apply := sourceFunction(t, buffs, "void WeaponBuffs_Apply(int client)")
	if !strings.Contains(apply, "WeaponBuffs_SyncNativeSelfBlast(client, false)") {
		t.Fatal("buff application does not synchronize native no-self-blast attributes")
	}
	if !strings.Contains(apply, "effect == NoSelfBlastEffect") {
		t.Fatal("generic provider still duplicates the native no-self-blast attribute")
	}
	remove := sourceFunction(t, buffs, "static void WeaponBuffs_Remove(int client)")
	if !strings.Contains(remove, "WeaponBuffs_SyncNativeSelfBlast(client, true)") {
		t.Fatal("buff removal leaves native no-self-blast attributes behind")
	}
}
