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

func TestNonMinigunProjectileDestructionUsesConfirmedShots(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	postThink := sourceFunction(t, buffs, "public void WeaponBuffs_PostThinkPost")
	for _, required := range []string{
		"clip < g_WeaponBuffShotClip[client]",
		"ammo < g_WeaponBuffShotAmmo[client]",
		"energy + 0.001 < g_WeaponBuffShotEnergy[client]",
		"nextAttack > g_WeaponBuffShotNextAttack[client] + 0.001",
		"WeaponBuffs_AttackEnemyProjectiles(client, levels)",
		"!WeaponBuffs_UsesNativeProjectileDestruction(catalog)",
	} {
		if !strings.Contains(postThink, required) {
			t.Errorf("confirmed-shot detector has no %q", required)
		}
	}

	sweep := sourceFunction(t, buffs, "static void WeaponBuffs_AttackEnemyProjectiles")
	for _, required := range []string{
		"ProjectileDestructionRange",
		"ProjectileDestructionRadius",
		"WeaponBuffs_ProjectileDestructionCooldown(levels)",
		"WeaponBuffs_IsDestroyableProjectile(projectile)",
		"WeaponBuffs_ProjectileVisible(client, start, point)",
		"chosen = projectile",
		"RemoveEntity(chosen)",
		"WeaponBuffs_ShowProjectileDud(chosen, chosenPoint)",
		`EmitGameSoundToAll("Halloween.HeadlessBossAxeHitWorld", chosen)`,
		"TE_SetupSparks(chosenPoint, sparkDirection, 2, 1)",
		`PrintCenterText(client, "PROJECTILE DESTROYED")`,
		`WeaponBuffs_DebugLog("[AP destroy]`,
	} {
		if !strings.Contains(sweep, required) {
			t.Errorf("projectile sweep has no %q", required)
		}
	}
	visible := sourceFunction(t, buffs, "static bool WeaponBuffs_ProjectileVisible")
	for _, required := range []string{
		"ScaleVector(direction, 48.0)",
		"fraction >= 0.99",
		"TR_TraceRayFilterEx",
		"WeaponBuffs_ProjectileVisibilityFilter, client",
		`WeaponBuffs_DebugLog("[AP destroy] candidate blocked at trace fraction`,
	} {
		if !strings.Contains(visible, required) {
			t.Errorf("projectile visibility test has no %q", required)
		}
	}
	filter := sourceFunction(t, buffs, "public bool WeaponBuffs_ProjectileVisibilityFilter")
	for _, required := range []string{
		"entity == client",
		`HasEntProp(entity, Prop_Send, "m_hOwnerEntity")`,
		`GetEntPropEnt(entity, Prop_Send, "m_hOwnerEntity") == client`,
	} {
		if !strings.Contains(filter, required) {
			t.Errorf("projectile visibility filter has no %q", required)
		}
	}
	target := sourceFunction(t, buffs, "static bool WeaponBuffs_IsDestroyableProjectile")
	if !strings.Contains(target, `StrContains(classname, "tf_projectile_", false) == 0`) {
		t.Fatal("projectile destruction does not accept every TF2 projectile class")
	}

	text, err := os.ReadFile(buffs)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"#define ProjectileDestructionRange 2000.0",
		"#define ProjectileDestructionRadius 256.0",
		"#define ProjectileDestructionBaseCooldown 2.0",
		"#define ProjectileDestructionCooldownStep 0.25",
		"#define ProjectileDestructionMinimumCooldown 0.5",
	} {
		if !strings.Contains(string(text), required) {
			t.Errorf("projectile destruction configuration has no %q", required)
		}
	}

	cooldown := sourceFunction(t, buffs, "static float WeaponBuffs_ProjectileDestructionCooldown")
	for _, required := range []string{
		"float(levels - 1) * ProjectileDestructionCooldownStep",
		"ProjectileDestructionMinimumCooldown",
	} {
		if !strings.Contains(cooldown, required) {
			t.Errorf("projectile destruction cooldown has no %q", required)
		}
	}
	dud := sourceFunction(t, buffs, "static void WeaponBuffs_ShowProjectileDud")
	for _, required := range []string{
		`StartMessageAll("BreakModelRocketDud", USERMSG_RELIABLE)`,
		"BfWriteShort(message, model)",
		"BfWriteVecCoord(message, point)",
		"BfWriteAngles(message, angles)",
	} {
		if !strings.Contains(dud, required) {
			t.Errorf("projectile dud effect has no %q", required)
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
		"WeaponBuffs_WeaponOfHit(attacker, inflictor, weapon)",
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

func TestExtraHealingBoltsUseNativeHealingAndPreservePenetration(t *testing.T) {
	buffs := "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"
	native := sourceFunction(t, buffs, "static int WeaponBuffs_FireNativeArrow")
	for _, required := range []string{
		"g_WeaponBuffArrowCreate", "m_iProjectileType",
		"SDKCall(g_WeaponBuffArrowCreate", "g_WeaponBuffProjectileSetLauncher",
		"SDKCall(g_WeaponBuffProjectileSetLauncher, extra, launcher)",
		`GetEntPropFloat(source, Prop_Data, "m_flDamage")`,
		"WeaponBuffs_CopyArrowPenetration(source, extra)",
	} {
		if !strings.Contains(native, required) {
			t.Fatalf("native arrow fanout has no %s", required)
		}
	}

	for _, forbidden := range []string{"g_WeaponBuffArrowFire", "SDKCall(g_WeaponBuffArrowFire"} {
		if strings.Contains(native, forbidden) {
			t.Fatalf("extra arrows still patch private penetration state with %s", forbidden)
		}
	}

	for _, path := range []string{
		"../plugin/gamedata/tf2_archipelago.txt",
		"../launcher/internal/assets/embedded/tf2_archipelago.txt",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if !strings.Contains(text,
			"@_ZN19CTFProjectile_Arrow6CreateERK6VectorRK6QAngleff16ProjectileType_tP11CBaseEntityS8_") {
			t.Fatalf("%s has no Linux native arrow factory", path)
		}
		if !strings.Contains(text, `"CBaseProjectile::SetLauncher"`) {
			t.Fatalf("%s has no native SetLauncher offset", path)
		}
	}
}
