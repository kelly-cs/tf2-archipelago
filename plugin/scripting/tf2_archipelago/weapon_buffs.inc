/** Every positive weapon/effect permutation received from the multiworld. */

#define ProjectileCountEffect 5

#define BuffMode_Add 1
#define BuffMode_Percentage 2
#define BuffMode_Toggle 3

int g_WeaponEffectLevels[WeaponCount][WeaponEffectCount];
int g_WeaponBuffAppliedRef[2049];
bool g_WeaponBuffApplied[2049][WeaponEffectCount];
int g_WeaponBuffAppliedLevels[2049][WeaponEffectCount];
float g_WeaponBuffAppliedValue[2049][WeaponEffectCount];
ConVar g_WeaponBuffProjectileDebug;
bool g_WeaponBuffFiringExtra;
int g_WeaponBuffLastExtra;
bool g_WeaponBuffNativeExtra[2049];
float g_WeaponBuffNativeExtraBorn[2049];

void WeaponBuffs_Init()
{
    g_WeaponBuffProjectileDebug = CreateConVar("tf2ap_projectile_debug", "0",
        "Log extra-projectile creation and lifetime", FCVAR_NONE, true, 0.0, true, 1.0);
}

public void OnEntityCreated(int entity, const char[] classname)
{
    if (entity <= MaxClients || entity >= sizeof(g_WeaponBuffNativeExtra))
    {
        return;
    }
    g_WeaponBuffNativeExtra[entity] = g_WeaponBuffFiringExtra;
    g_WeaponBuffNativeExtraBorn[entity] = GetGameTime();
    if (g_WeaponBuffFiringExtra && StrEqual(classname, "tf_projectile_rocket", false))
    {
        g_WeaponBuffLastExtra = entity;
    }
    if (StrEqual(classname, "tf_projectile_rocket", false))
    {
        if (!g_WeaponBuffFiringExtra)
        {
            SDKHook(entity, SDKHook_SpawnPost, WeaponBuffs_RocketSpawned);
        }
    }
}

public void OnEntityDestroyed(int entity)
{
    if (entity > MaxClients && entity < sizeof(g_WeaponBuffNativeExtra))
    {
        if (g_WeaponBuffNativeExtra[entity] && g_WeaponBuffProjectileDebug.BoolValue)
        {
            LogMessage("[AP projectile] extra entity=%d removed after %.3f seconds",
                entity, GetGameTime() - g_WeaponBuffNativeExtraBorn[entity]);
        }
        g_WeaponBuffNativeExtra[entity] = false;
    }
}

public void WeaponBuffs_RocketSpawned(int entity)
{
    SDKUnhook(entity, SDKHook_SpawnPost, WeaponBuffs_RocketSpawned);
    RequestFrame(WeaponBuffs_FireExtraRockets, EntIndexToEntRef(entity));
}

static void WeaponBuffs_LogProjectile(const char[] label, int entity)
{
    if (!g_WeaponBuffProjectileDebug.BoolValue || entity <= MaxClients || !IsValidEntity(entity))
    {
        return;
    }
    int owner = HasEntProp(entity, Prop_Send, "m_hOwnerEntity")
        ? GetEntPropEnt(entity, Prop_Send, "m_hOwnerEntity") : -1;
    int launcher = HasEntProp(entity, Prop_Send, "m_hLauncher")
        ? GetEntPropEnt(entity, Prop_Send, "m_hLauncher") : -1;
    int team = GetEntProp(entity, Prop_Send, "m_iTeamNum");
    int collision = GetEntProp(entity, Prop_Send, "m_CollisionGroup");
    float damage = HasEntProp(entity, Prop_Data, "m_flDamage")
        ? GetEntPropFloat(entity, Prop_Data, "m_flDamage") : -1.0;
    float origin[3], velocity[3];
    GetEntPropVector(entity, Prop_Data, "m_vecAbsOrigin", origin);
    GetEntPropVector(entity, Prop_Data, "m_vecAbsVelocity", velocity);
    LogMessage("[AP projectile] %s entity=%d owner=%d launcher=%d team=%d collision=%d damage=%.1f origin=%.1f %.1f %.1f velocity=%.1f %.1f %.1f",
        label, entity, owner, launcher, team, collision, damage, origin[0], origin[1], origin[2],
        velocity[0], velocity[1], velocity[2]);
}

static void WeaponBuffs_FireExtraRockets(any reference)
{
    int source = EntRefToEntIndex(reference);
    if (source <= MaxClients || !IsValidEntity(source))
    {
        return;
    }
    int owner = HasEntProp(source, Prop_Send, "m_hOwnerEntity")
        ? GetEntPropEnt(source, Prop_Send, "m_hOwnerEntity") : -1;
    int launcher = HasEntProp(source, Prop_Send, "m_hLauncher")
        ? GetEntPropEnt(source, Prop_Send, "m_hLauncher") : -1;
    if (!MvM_IsPlayer(owner) || launcher <= MaxClients || !IsValidEntity(launcher))
    {
        return;
    }
    int weapon = WeaponBuffs_ForEntity(launcher);
    if (weapon < 0)
    {
        return;
    }
    int copies = g_WeaponEffectLevels[weapon][ProjectileCountEffect];
    if (copies <= 0)
    {
        return;
    }

    WeaponBuffs_LogProjectile("original", source);
    float origin[3], sourceVelocity[3];
    GetEntPropVector(source, Prop_Data, "m_vecAbsOrigin", origin);
    GetEntPropVector(source, Prop_Data, "m_vecAbsVelocity", sourceVelocity);
    float speed = SquareRoot(sourceVelocity[0] * sourceVelocity[0]
        + sourceVelocity[1] * sourceVelocity[1]
        + sourceVelocity[2] * sourceVelocity[2]);
    float damage = HasEntProp(source, Prop_Data, "m_flDamage")
        ? GetEntPropFloat(source, Prop_Data, "m_flDamage") : 90.0;
    bool critical = HasEntProp(source, Prop_Send, "m_bCritical")
        && GetEntProp(source, Prop_Send, "m_bCritical") != 0;

    for (int copy = 0; copy < copies; copy++)
    {
        float velocity[3], angles[3];
        velocity = sourceVelocity;
        float degrees = 3.0 * float(copy / 2 + 1);
        if (copy % 2 != 0) degrees = -degrees;
        float radians = DegToRad(degrees);
        float x = velocity[0];
        velocity[0] = x * Cosine(radians) - velocity[1] * Sine(radians);
        velocity[1] = x * Sine(radians) + velocity[1] * Cosine(radians);
        GetVectorAngles(velocity, angles);

        // tf_point_weapon_mimic calls TF2's real CTFProjectile_Rocket::Create.
        // Unlike cloning an entity, that initializes its touch/collision state;
        // unlike a private SDKCall, it needs no version-sensitive vtable offset.
        int mimic = CreateEntityByName("tf_point_weapon_mimic");
        if (mimic == -1)
        {
            LogError("Could not create tf_point_weapon_mimic for projectile-count.");
            return;
        }
        DispatchKeyValue(mimic, "WeaponType", "0");
        DispatchKeyValueFloat(mimic, "SpeedMin", speed);
        DispatchKeyValueFloat(mimic, "SpeedMax", speed);
        DispatchKeyValueFloat(mimic, "Damage", damage);
        DispatchKeyValueFloat(mimic, "SplashRadius", 146.0);
        DispatchKeyValue(mimic, "Crits", critical ? "1" : "0");
        DispatchSpawn(mimic);
        TeleportEntity(mimic, origin, angles, NULL_VECTOR);

        g_WeaponBuffLastExtra = -1;
        g_WeaponBuffFiringExtra = true;
        AcceptEntityInput(mimic, "FireOnce");
        g_WeaponBuffFiringExtra = false;
        int extra = g_WeaponBuffLastExtra;
        RemoveEntity(mimic);
        if (extra <= MaxClients || !IsValidEntity(extra))
        {
            if (g_WeaponBuffProjectileDebug.BoolValue)
            {
                LogMessage("[AP projectile] weapon mimic produced no rocket (entity %d)", extra);
            }
            continue;
        }

        SetEntPropEnt(extra, Prop_Send, "m_hOwnerEntity", owner);
        SetEntPropEnt(extra, Prop_Send, "m_hLauncher", launcher);
        if (HasEntProp(extra, Prop_Send, "m_hOriginalLauncher"))
        {
            SetEntPropEnt(extra, Prop_Send, "m_hOriginalLauncher", launcher);
        }
        SetEntProp(extra, Prop_Send, "m_iTeamNum", GetClientTeam(owner));
        if (HasEntProp(extra, Prop_Send, "m_vInitialVelocity"))
        {
            SetEntPropVector(extra, Prop_Send, "m_vInitialVelocity", velocity);
        }
        TeleportEntity(extra, NULL_VECTOR, angles, velocity);
        WeaponBuffs_LogProjectile("mimic-extra", extra);
    }
}

public Action Command_ProjectileDebug(int client, int args)
{
    if (args >= 1)
    {
        char wanted[16];
        GetCmdArg(1, wanted, sizeof(wanted));
        g_WeaponBuffProjectileDebug.BoolValue = StrEqual(wanted, "on", false)
            || StrEqual(wanted, "1");
    }
    ReplyToCommand(client, "[AP] Projectile debug %s; backend=tf_point_weapon_mimic.",
        g_WeaponBuffProjectileDebug.BoolValue ? "ON" : "OFF");
    if (MvM_IsPlayer(client) && IsPlayerAlive(client))
    {
        int entity = GetEntPropEnt(client, Prop_Send, "m_hActiveWeapon");
        int weapon = WeaponBuffs_ForEntity(entity);
        char classname[64];
        GetEntityClassname(entity, classname, sizeof(classname));
        int levels = weapon >= 0 ? g_WeaponEffectLevels[weapon][ProjectileCountEffect] : 0;
        ReplyToCommand(client, "[AP] Active entity=%d class=%s catalog=%d projectile-count=%d.",
            entity, classname, weapon, levels);
    }
    return Plugin_Handled;
}

void WeaponBuffs_Clear()
{
    for (int weapon = 0; weapon < WeaponCount; weapon++)
    {
        for (int effect = 0; effect < WeaponEffectCount; effect++)
        {
            g_WeaponEffectLevels[weapon][effect] = 0;
        }
    }
}

bool WeaponBuffs_Grant(const char[] key)
{
    for (int index = 0; index < sizeof(g_WeaponBuffKeys); index++)
    {
        if (StrEqual(g_WeaponBuffKeys[index], key))
        {
            int weapon = g_WeaponBuffWeapon[index];
            int effect = g_WeaponBuffEffect[index];
            if (g_WeaponEffectModes[effect] == BuffMode_Toggle)
            {
                g_WeaponEffectLevels[weapon][effect] = 1;
            }
            else
            {
                g_WeaponEffectLevels[weapon][effect]++;
            }
            return true;
        }
    }
    LogError("The bridge granted weapon buff %s. This plugin does not know it.", key);
    return false;
}

static int WeaponBuffs_ForDefinition(int definition)
{
    for (int index = 0; index < sizeof(g_WeaponByDefinition); index++)
    {
        if (g_WeaponByDefinition[index][0] == definition)
        {
            return g_WeaponByDefinition[index][1];
        }
    }
    return -1;
}

static int WeaponBuffs_ForEntity(int entity)
{
    if (entity <= MaxClients || !IsValidEntity(entity)
        || !HasEntProp(entity, Prop_Send, "m_iItemDefinitionIndex"))
    {
        return -1;
    }
    return WeaponBuffs_ForDefinition(GetEntProp(entity, Prop_Send, "m_iItemDefinitionIndex"));
}

// Upgrade-station values are composed first. A matching previous result means
// only newly received levels are added; a different value means TF2 rebuilt
// the attribute and the complete Archipelago contribution is added again.
static void WeaponBuffs_ApplyEntity(int entity, int weapon)
{
    int reference = EntIndexToEntRef(entity);
    if (g_WeaponBuffAppliedRef[entity] != reference)
    {
        for (int effect = 0; effect < WeaponEffectCount; effect++)
        {
            g_WeaponBuffApplied[entity][effect] = false;
            g_WeaponBuffAppliedLevels[entity][effect] = 0;
        }
        g_WeaponBuffAppliedRef[entity] = reference;
    }

    for (int effect = 0; effect < WeaponEffectCount; effect++)
    {
        int levels = g_WeaponEffectLevels[weapon][effect];
        if (levels == 0)
        {
            continue;
        }
        Address attribute = TF2Attrib_GetByName(entity, g_WeaponEffectAttributes[effect]);
        float current = g_WeaponEffectModes[effect] == BuffMode_Add ? 0.0 : 1.0;
        if (attribute != Address_Null)
        {
            current = TF2Attrib_GetValue(attribute);
        }

        bool unchanged = g_WeaponBuffApplied[entity][effect]
            && FloatAbs(current - g_WeaponBuffAppliedValue[entity][effect]) < 0.0001;
        if (unchanged && g_WeaponBuffAppliedLevels[entity][effect] == levels)
        {
            continue;
        }

        float value;
        if (g_WeaponEffectModes[effect] == BuffMode_Toggle)
        {
            value = 1.0;
        }
        else
        {
            int addedLevels = unchanged
                ? levels - g_WeaponBuffAppliedLevels[entity][effect]
                : levels;
            value = current + g_WeaponEffectIncrements[effect] * float(addedLevels);
        }
        if (!TF2Attrib_SetByName(entity, g_WeaponEffectAttributes[effect], value))
        {
            LogError("Could not apply %s to %s.",
                g_WeaponEffectAttributes[effect], g_WeaponNames[weapon]);
            continue;
        }
        g_WeaponBuffApplied[entity][effect] = true;
        g_WeaponBuffAppliedLevels[entity][effect] = levels;
        g_WeaponBuffAppliedValue[entity][effect] = value;
    }
}

void WeaponBuffs_Apply(int client)
{
    if (!g_HaveUnlocks || !MvM_IsPlayer(client) || !IsPlayerAlive(client))
    {
        return;
    }
    for (int slot = 0; slot <= 5; slot++)
    {
        int entity = GetPlayerWeaponSlot(client, slot);
        int weapon = WeaponBuffs_ForEntity(entity);
        if (weapon >= 0)
        {
            WeaponBuffs_ApplyEntity(entity, weapon);
        }
    }
}

static void Frame_ApplyWeaponBuffs(any userid)
{
    int client = GetClientOfUserId(userid);
    if (client > 0)
    {
        WeaponBuffs_Apply(client);
    }
}

void WeaponBuffs_ApplyNextFrame(int client)
{
    if (client > 0)
    {
        RequestFrame(Frame_ApplyWeaponBuffs, GetClientUserId(client));
    }
}

public int Menu_WeaponBuffs(Menu menu, MenuAction action, int client, int selection)
{
    if (action == MenuAction_End)
    {
        delete menu;
    }
    return 0;
}

void WeaponBuffs_ShowMenu(int client)
{
    if (!MvM_IsPlayer(client))
    {
        return;
    }
    Menu menu = new Menu(Menu_WeaponBuffs);
    menu.SetTitle("Archipelago weapon buffs");
    int listed;
    for (int slot = 0; slot <= 5; slot++)
    {
        int weapon = WeaponBuffs_ForEntity(GetPlayerWeaponSlot(client, slot));
        if (weapon < 0)
        {
            continue;
        }
        for (int effect = 0; effect < WeaponEffectCount; effect++)
        {
            int levels = g_WeaponEffectLevels[weapon][effect];
            if (levels == 0)
            {
                continue;
            }
            char line[160];
            if (levels > 1)
            {
                FormatEx(line, sizeof(line), "%s: %s x%d",
                    g_WeaponNames[weapon], g_WeaponEffectDescriptions[effect], levels);
            }
            else
            {
                FormatEx(line, sizeof(line), "%s: %s",
                    g_WeaponNames[weapon], g_WeaponEffectDescriptions[effect]);
            }
            menu.AddItem(g_WeaponEffectKeys[effect], line);
            listed++;
        }
    }
    if (listed == 0)
    {
        menu.AddItem("none", "No buffs for this loadout yet", ITEMDRAW_DISABLED);
    }
    menu.ExitButton = true;
    menu.Display(client, 20);
}

public Action Command_WeaponBuffs(int client, int args)
{
    WeaponBuffs_ShowMenu(client);
    return Plugin_Handled;
}

static int WeaponBuffs_AddTestEffects(int weapon, const char[] wanted, int levels)
{
    int granted;
    for (int effect = 0; effect < WeaponEffectCount; effect++)
    {
        bool selected = StrEqual(wanted, "all", false)
            || StrEqual(wanted, g_WeaponEffectKeys[effect], false)
            || StringToInt(wanted) == effect + 1;
        if (!selected)
        {
            continue;
        }
        g_WeaponEffectLevels[weapon][effect] = g_WeaponEffectModes[effect] == BuffMode_Toggle
            ? 1 : g_WeaponEffectLevels[weapon][effect] + levels;
        granted++;
    }
    return granted;
}

// sm_ap_buff_test <effect number|effect key|all> [levels]
public Action Command_TestWeaponBuff(int client, int args)
{
    if (!MvM_IsPlayer(client) || !IsPlayerAlive(client))
    {
        ReplyToCommand(client, "[AP] Join RED and spawn before testing weapon buffs.");
        return Plugin_Handled;
    }
    if (args < 1)
    {
        ReplyToCommand(client, "[AP] Usage: sm_ap_buff_test <1-%d|effect-key|all> [levels]", WeaponEffectCount);
        return Plugin_Handled;
    }
    int entity = GetEntPropEnt(client, Prop_Send, "m_hActiveWeapon");
    int weapon = WeaponBuffs_ForEntity(entity);
    if (weapon < 0)
    {
        ReplyToCommand(client, "[AP] The active weapon is not in the catalog.");
        return Plugin_Handled;
    }
    char wanted[64];
    GetCmdArg(1, wanted, sizeof(wanted));
    int levels = 1;
    if (args >= 2)
    {
        char rawLevels[12];
        GetCmdArg(2, rawLevels, sizeof(rawLevels));
        levels = StringToInt(rawLevels);
        if (levels < 1) levels = 1;
    }
    int granted = WeaponBuffs_AddTestEffects(weapon, wanted, levels);
    if (granted == 0)
    {
        ReplyToCommand(client, "[AP] Unknown effect %s. Use 1-%d or an effect key.", wanted, WeaponEffectCount);
        return Plugin_Handled;
    }
    WeaponBuffs_ApplyEntity(entity, weapon);
    ReplyToCommand(client, "[AP] Applied %d test effect(s) to %s.", granted, g_WeaponNames[weapon]);
    WeaponBuffs_ShowMenu(client);
    return Plugin_Handled;
}

// sm_ap_buff_give <target> <effect number|effect key|all> [levels]
public Action Command_GiveWeaponBuff(int client, int args)
{
    if (args < 2)
    {
        ReplyToCommand(client,
            "[AP] Usage: sm_ap_buff_give <target> <1-%d|effect-key|all> [levels]",
            WeaponEffectCount);
        return Plugin_Handled;
    }

    char pattern[MAX_TARGET_LENGTH];
    GetCmdArg(1, pattern, sizeof(pattern));
    int target = FindTarget(client, pattern, true);
    if (target < 1)
    {
        return Plugin_Handled;
    }
    if (!MvM_IsPlayer(target) || !IsPlayerAlive(target))
    {
        ReplyToCommand(client, "[AP] %N must be alive on RED before receiving a test buff.", target);
        return Plugin_Handled;
    }

    int entity = GetEntPropEnt(target, Prop_Send, "m_hActiveWeapon");
    int weapon = WeaponBuffs_ForEntity(entity);
    if (weapon < 0)
    {
        ReplyToCommand(client, "[AP] %N's active weapon is not in the catalog.", target);
        return Plugin_Handled;
    }

    char wanted[64];
    GetCmdArg(2, wanted, sizeof(wanted));
    int levels = 1;
    if (args >= 3)
    {
        char rawLevels[12];
        GetCmdArg(3, rawLevels, sizeof(rawLevels));
        levels = StringToInt(rawLevels);
        if (levels < 1) levels = 1;
    }

    int granted = WeaponBuffs_AddTestEffects(weapon, wanted, levels);
    if (granted == 0)
    {
        ReplyToCommand(client, "[AP] Unknown effect %s. Use 1-%d or an effect key.",
            wanted, WeaponEffectCount);
        return Plugin_Handled;
    }

    WeaponBuffs_ApplyEntity(entity, weapon);
    ReplyToCommand(client, "[AP] Applied %d test effect(s) to %N's %s.",
        granted, target, g_WeaponNames[weapon]);
    if (client != target)
    {
        PrintToChat(target, "[AP] An admin applied %d test effect(s) to your %s.",
            granted, g_WeaponNames[weapon]);
    }
    WeaponBuffs_ShowMenu(target);
    return Plugin_Handled;
}
