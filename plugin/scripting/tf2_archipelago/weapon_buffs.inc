/** Every positive weapon/effect permutation received from the multiworld. */

#define ProjectileCountEffect 5
#define IgniteEffect 9

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
int g_WeaponBuffStickyMimic[2049];
int g_WeaponBuffLastButtons[MAXPLAYERS + 1];
#define ProjectileDebugLineCount 12
char g_WeaponBuffProjectileLines[ProjectileDebugLineCount][256];
int g_WeaponBuffProjectileLineNext;
int g_WeaponBuffProjectileLineHeld;

static void WeaponBuffs_DebugLog(const char[] format, any ...)
{
    char line[256];
    VFormat(line, sizeof(line), format, 2);
    LogMessage("%s", line);
    strcopy(g_WeaponBuffProjectileLines[g_WeaponBuffProjectileLineNext],
        sizeof(g_WeaponBuffProjectileLines[]), line);
    g_WeaponBuffProjectileLineNext = (g_WeaponBuffProjectileLineNext + 1) % ProjectileDebugLineCount;
    if (g_WeaponBuffProjectileLineHeld < ProjectileDebugLineCount)
    {
        g_WeaponBuffProjectileLineHeld++;
    }
}

void WeaponBuffs_Init()
{
    g_WeaponBuffProjectileDebug = CreateConVar("tf2ap_projectile_debug", "0",
        "Log extra-projectile creation and lifetime", FCVAR_NONE, true, 0.0, true, 1.0);
    for (int client = 1; client <= MaxClients; client++)
    {
        if (IsClientInGame(client))
        {
            WeaponBuffs_HookClient(client);
        }
    }
}

void WeaponBuffs_HookClient(int client)
{
    g_WeaponBuffLastButtons[client] = 0;
    SDKUnhook(client, SDKHook_OnTakeDamagePost, WeaponBuffs_OnTakeDamagePost);
    SDKHook(client, SDKHook_OnTakeDamagePost, WeaponBuffs_OnTakeDamagePost);
}

static bool WeaponBuffs_IsProjectile(const char[] classname)
{
    return StrContains(classname, "tf_projectile_", false) == 0
        && !StrEqual(classname, "tf_projectile_sentryrocket", false);
}

public void OnEntityCreated(int entity, const char[] classname)
{
    if (entity <= MaxClients || entity >= sizeof(g_WeaponBuffNativeExtra))
    {
        return;
    }
    g_WeaponBuffNativeExtra[entity] = g_WeaponBuffFiringExtra;
    g_WeaponBuffNativeExtraBorn[entity] = GetGameTime();
    if (WeaponBuffs_IsProjectile(classname) && g_WeaponBuffProjectileDebug != null
        && g_WeaponBuffProjectileDebug.BoolValue)
    {
        WeaponBuffs_DebugLog("[AP projectile] created entity=%d class=%s extra=%d",
            entity, classname, g_WeaponBuffFiringExtra);
    }
    if (g_WeaponBuffFiringExtra && WeaponBuffs_IsProjectile(classname))
    {
        g_WeaponBuffLastExtra = entity;
    }
    if (WeaponBuffs_IsProjectile(classname))
    {
        if (!g_WeaponBuffFiringExtra)
        {
            SDKHook(entity, SDKHook_SpawnPost, WeaponBuffs_ProjectileSpawned);
        }
    }
}

public void OnEntityDestroyed(int entity)
{
    if (entity > MaxClients && entity < sizeof(g_WeaponBuffNativeExtra))
    {
        int mimic = EntRefToEntIndex(g_WeaponBuffStickyMimic[entity]);
        g_WeaponBuffStickyMimic[entity] = 0;
        if (mimic > MaxClients && IsValidEntity(mimic))
        {
            RequestFrame(WeaponBuffs_RemoveEntityFrame, EntIndexToEntRef(mimic));
        }
        if (g_WeaponBuffNativeExtra[entity] && g_WeaponBuffProjectileDebug.BoolValue)
        {
            WeaponBuffs_DebugLog("[AP projectile] extra entity=%d removed after %.3f seconds",
                entity, GetGameTime() - g_WeaponBuffNativeExtraBorn[entity]);
        }
        g_WeaponBuffNativeExtra[entity] = false;
    }
}

static void WeaponBuffs_RemoveEntityFrame(any reference)
{
    int entity = EntRefToEntIndex(reference);
    if (entity > MaxClients && IsValidEntity(entity))
    {
        RemoveEntity(entity);
    }
}

public Action OnPlayerRunCmd(int client, int &buttons, int &impulse, float velocity[3],
    float angles[3], int &weapon)
{
    bool pressed = (buttons & IN_ATTACK2) != 0
        && (g_WeaponBuffLastButtons[client] & IN_ATTACK2) == 0;
    g_WeaponBuffLastButtons[client] = buttons;
    if (!pressed || !MvM_IsPlayer(client))
    {
        return Plugin_Continue;
    }
    for (int projectile = MaxClients + 1; projectile < sizeof(g_WeaponBuffStickyMimic); projectile++)
    {
        int mimic = EntRefToEntIndex(g_WeaponBuffStickyMimic[projectile]);
        if (mimic <= MaxClients || !IsValidEntity(mimic))
        {
            continue;
        }
        int owner = IsValidEntity(projectile) ? WeaponBuffs_EntityOwner(projectile) : -1;
        if (owner == client)
        {
            AcceptEntityInput(mimic, "DetonateStickies");
        }
    }
    return Plugin_Continue;
}

public void WeaponBuffs_ProjectileSpawned(int entity)
{
    SDKUnhook(entity, SDKHook_SpawnPost, WeaponBuffs_ProjectileSpawned);
    if (g_WeaponBuffProjectileDebug.BoolValue && IsValidEntity(entity))
    {
        char classname[64];
        GetEntityClassname(entity, classname, sizeof(classname));
        WeaponBuffs_DebugLog("[AP projectile] spawn-post entity=%d class=%s", entity, classname);
    }
    RequestFrame(WeaponBuffs_FireExtraProjectiles, EntIndexToEntRef(entity));
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
    WeaponBuffs_DebugLog("[AP projectile] %s entity=%d owner=%d launcher=%d team=%d collision=%d damage=%.1f origin=%.1f %.1f %.1f velocity=%.1f %.1f %.1f",
        label, entity, owner, launcher, team, collision, damage, origin[0], origin[1], origin[2],
        velocity[0], velocity[1], velocity[2]);
}

static int WeaponBuffs_MimicType(int source, const char[] classname)
{
    if (StrEqual(classname, "tf_projectile_rocket", false)) return 0;
    if (StrEqual(classname, "tf_projectile_pipe", false))
    {
        // Both names can represent the same CTFGrenadePipebombProjectile.
        // m_iType=1/2 is a sticky; zero is a timed grenade.
        if (HasEntProp(source, Prop_Send, "m_iType")
            && GetEntProp(source, Prop_Send, "m_iType") != 0)
        {
            return 3;
        }
        return 1;
    }
    if (StrEqual(classname, "tf_projectile_arrow", false)) return 2;
    if (StrEqual(classname, "tf_projectile_pipe_remote", false)) return 3;
    return -1;
}

static void WeaponBuffs_CopySendInt(int source, int extra, const char[] property)
{
    if (HasEntProp(source, Prop_Send, property) && HasEntProp(extra, Prop_Send, property))
    {
        SetEntProp(extra, Prop_Send, property, GetEntProp(source, Prop_Send, property));
    }
}

static void WeaponBuffs_CopySendEntity(int source, int extra, const char[] property)
{
    if (HasEntProp(source, Prop_Send, property) && HasEntProp(extra, Prop_Send, property))
    {
        SetEntPropEnt(extra, Prop_Send, property, GetEntPropEnt(source, Prop_Send, property));
    }
}

static void WeaponBuffs_CopyDataFloat(int source, int extra, const char[] property)
{
    if (HasEntProp(source, Prop_Data, property) && HasEntProp(extra, Prop_Data, property))
    {
        SetEntPropFloat(extra, Prop_Data, property, GetEntPropFloat(source, Prop_Data, property));
    }
}

static int WeaponBuffs_CloneProjectile(int source, const char[] classname, int owner,
    int launcher, const float origin[3], const float angles[3], const float velocity[3])
{
    g_WeaponBuffFiringExtra = true;
    int extra = CreateEntityByName(classname);
    g_WeaponBuffFiringExtra = false;
    if (extra == -1)
    {
        LogError("Could not create %s for projectile-count.", classname);
        return -1;
    }

    // Projectile type selects the correct model and behavior for arrows,
    // healing bolts and the festive variants, and must be set before Spawn.
    WeaponBuffs_CopySendInt(source, extra, "m_iProjectileType");
    if (HasEntProp(extra, Prop_Send, "m_hOwnerEntity"))
    {
        SetEntPropEnt(extra, Prop_Send, "m_hOwnerEntity", owner);
    }
    SetEntProp(extra, Prop_Send, "m_iTeamNum", GetClientTeam(owner));
    // Flare Spawn() reads both launcher handles to choose projectile speed,
    // heat-seeking behavior and other weapon-specific behavior.
    if (HasEntProp(extra, Prop_Send, "m_hLauncher"))
    {
        SetEntPropEnt(extra, Prop_Send, "m_hLauncher", launcher);
    }
    if (HasEntProp(extra, Prop_Send, "m_hOriginalLauncher"))
    {
        SetEntPropEnt(extra, Prop_Send, "m_hOriginalLauncher", launcher);
    }
    // CTFProjectile_Flare::Spawn traces from the owner's eyes to its current
    // origin and removes itself if that path crosses world geometry. Creating
    // at 0,0,0 and moving afterward therefore fizzles every cloned flare.
    TeleportEntity(extra, origin, angles, NULL_VECTOR);
    DispatchSpawn(extra);
    ActivateEntity(extra);

    WeaponBuffs_CopySendEntity(source, extra, "m_hLauncher");
    WeaponBuffs_CopySendEntity(source, extra, "m_hOriginalLauncher");
    WeaponBuffs_CopySendInt(source, extra, "m_bCritical");
    WeaponBuffs_CopySendInt(source, extra, "m_bArrowAlight");
    WeaponBuffs_CopySendInt(source, extra, "m_bChargedShot");
    WeaponBuffs_CopySendInt(source, extra, "m_CollisionGroup");
    WeaponBuffs_CopyDataFloat(source, extra, "m_flDamage");
    WeaponBuffs_CopyDataFloat(source, extra, "m_flDamageRadius");
    WeaponBuffs_CopyDataFloat(source, extra, "m_flDetonateTime");
    WeaponBuffs_CopyDataFloat(source, extra, "m_flCreationTime");
    WeaponBuffs_CopyDataFloat(source, extra, "m_flInitTime");
    if (HasEntProp(extra, Prop_Send, "m_hLauncher"))
    {
        SetEntPropEnt(extra, Prop_Send, "m_hLauncher", launcher);
    }
    if (HasEntProp(extra, Prop_Send, "m_hOriginalLauncher"))
    {
        SetEntPropEnt(extra, Prop_Send, "m_hOriginalLauncher", launcher);
    }
    if (HasEntProp(extra, Prop_Send, "m_vInitialVelocity"))
    {
        SetEntPropVector(extra, Prop_Send, "m_vInitialVelocity", velocity);
    }
    SetEntityMoveType(extra, GetEntityMoveType(source));
    SetEntityGravity(extra, GetEntityGravity(source));
    TeleportEntity(extra, origin, angles, velocity);
    return extra;
}

static int WeaponBuffs_FireMimicProjectile(int mimicType, int owner,
    int launcher, float speed, float damage, bool critical, const float origin[3],
    const float angles[3], const float velocity[3])
{
    int mimic = CreateEntityByName("tf_point_weapon_mimic");
    if (mimic == -1)
    {
        LogError("Could not create tf_point_weapon_mimic for projectile-count.");
        return -1;
    }
    char rawType[4];
    IntToString(mimicType, rawType, sizeof(rawType));
    DispatchKeyValue(mimic, "WeaponType", rawType);
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
    if (extra <= MaxClients || !IsValidEntity(extra))
    {
        RemoveEntity(mimic);
        return -1;
    }
    if (mimicType == 3)
    {
        // The mimic owns the private list required to detonate the sticky.
        // Keep it until the projectile dies and forward secondary fire to it.
        g_WeaponBuffStickyMimic[extra] = EntIndexToEntRef(mimic);
    }
    else
    {
        RemoveEntity(mimic);
    }
    SetEntPropEnt(extra, Prop_Send, "m_hOwnerEntity", owner);
    if (HasEntProp(extra, Prop_Send, "m_hLauncher"))
    {
        SetEntPropEnt(extra, Prop_Send, "m_hLauncher", launcher);
    }
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
    return extra;
}

static int WeaponBuffs_EntityOwner(int entity)
{
    if (entity > 0 && entity <= MaxClients)
    {
        return entity;
    }
    if (entity > MaxClients && IsValidEntity(entity)
        && HasEntProp(entity, Prop_Send, "m_hOwnerEntity"))
    {
        return GetEntPropEnt(entity, Prop_Send, "m_hOwnerEntity");
    }
    return -1;
}

// TF2 projectile families do not agree about m_hOwnerEntity: depending on the
// weapon, it can be the player or the launcher. Normalize both shapes here.
static void WeaponBuffs_ResolveProjectile(int source, int &owner, int &launcher)
{
    owner = WeaponBuffs_EntityOwner(source);
    launcher = HasEntProp(source, Prop_Send, "m_hLauncher")
        ? GetEntPropEnt(source, Prop_Send, "m_hLauncher") : -1;
    if ((launcher <= MaxClients || !IsValidEntity(launcher))
        && HasEntProp(source, Prop_Send, "m_hOriginalLauncher"))
    {
        launcher = GetEntPropEnt(source, Prop_Send, "m_hOriginalLauncher");
    }

    // Flares and several grenade-derived projectiles expose the weapon as the
    // owner rather than through m_hLauncher.
    if (WeaponBuffs_ForEntity(owner) >= 0)
    {
        launcher = owner;
        owner = WeaponBuffs_EntityOwner(launcher);
    }
    if (!MvM_IsPlayer(owner) && launcher > MaxClients && IsValidEntity(launcher))
    {
        int launcherOwner = WeaponBuffs_EntityOwner(launcher);
        if (MvM_IsPlayer(launcherOwner))
        {
            owner = launcherOwner;
        }
    }
    if ((launcher <= MaxClients || !IsValidEntity(launcher)) && MvM_IsPlayer(owner))
    {
        launcher = GetEntPropEnt(owner, Prop_Send, "m_hActiveWeapon");
    }
}

static void WeaponBuffs_FireExtraProjectiles(any reference)
{
    int source = EntRefToEntIndex(reference);
    if (source <= MaxClients || !IsValidEntity(source))
    {
        return;
    }
    int owner, launcher;
    WeaponBuffs_ResolveProjectile(source, owner, launcher);
    if (!MvM_IsPlayer(owner) || launcher <= MaxClients || !IsValidEntity(launcher))
    {
        if (g_WeaponBuffProjectileDebug.BoolValue)
        {
            char rejectedClass[64];
            GetEntityClassname(source, rejectedClass, sizeof(rejectedClass));
            WeaponBuffs_DebugLog("[AP projectile] rejected entity=%d class=%s owner=%d launcher=%d (unresolved owner/launcher)",
                source, rejectedClass, owner, launcher);
        }
        return;
    }
    int weapon = WeaponBuffs_ForEntity(launcher);
    if (weapon < 0)
    {
        if (g_WeaponBuffProjectileDebug.BoolValue)
        {
            char rejectedClass[64], launcherClass[64];
            GetEntityClassname(source, rejectedClass, sizeof(rejectedClass));
            GetEntityClassname(launcher, launcherClass, sizeof(launcherClass));
            WeaponBuffs_DebugLog("[AP projectile] rejected entity=%d class=%s owner=%d launcher=%d class=%s (launcher not catalogued)",
                source, rejectedClass, owner, launcher, launcherClass);
        }
        return;
    }
    int copies = g_WeaponEffectLevels[weapon][ProjectileCountEffect];
    if (copies <= 0)
    {
        if (g_WeaponBuffProjectileDebug.BoolValue)
        {
            char rejectedClass[64];
            GetEntityClassname(source, rejectedClass, sizeof(rejectedClass));
            WeaponBuffs_DebugLog("[AP projectile] ignored entity=%d class=%s owner=%d launcher=%d catalog=%d projectile-count=0",
                source, rejectedClass, owner, launcher, weapon);
        }
        return;
    }

    WeaponBuffs_LogProjectile("original", source);
    float origin[3], sourceVelocity[3];
    GetEntPropVector(source, Prop_Data, "m_vecAbsOrigin", origin);
    GetEntPropVector(source, Prop_Data, "m_vecAbsVelocity", sourceVelocity);
    float absSpeedSquared = sourceVelocity[0] * sourceVelocity[0]
        + sourceVelocity[1] * sourceVelocity[1]
        + sourceVelocity[2] * sourceVelocity[2];
    // Pipebombs, stickies and flares derive from the grenade projectile base.
    // During this deferred callback their physics velocity can read as zero,
    // while this networked launch vector still contains the real shot speed.
    if (absSpeedSquared < 1.0 && HasEntProp(source, Prop_Send, "m_vInitialVelocity"))
    {
        GetEntPropVector(source, Prop_Send, "m_vInitialVelocity", sourceVelocity);
    }
    float speed = SquareRoot(sourceVelocity[0] * sourceVelocity[0]
        + sourceVelocity[1] * sourceVelocity[1]
        + sourceVelocity[2] * sourceVelocity[2]);
    float damage = HasEntProp(source, Prop_Data, "m_flDamage")
        ? GetEntPropFloat(source, Prop_Data, "m_flDamage") : 90.0;
    bool critical = HasEntProp(source, Prop_Send, "m_bCritical")
        && GetEntProp(source, Prop_Send, "m_bCritical") != 0;
    char classname[64];
    GetEntityClassname(source, classname, sizeof(classname));
    int mimicType = WeaponBuffs_MimicType(source, classname);

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

        int extra = mimicType >= 0
            ? WeaponBuffs_FireMimicProjectile(mimicType, owner, launcher,
                speed, damage, critical, origin, angles, velocity)
            : WeaponBuffs_CloneProjectile(source, classname, owner, launcher,
                origin, angles, velocity);
        if (extra <= MaxClients || !IsValidEntity(extra))
        {
            if (g_WeaponBuffProjectileDebug.BoolValue)
            {
                WeaponBuffs_DebugLog("[AP projectile] backend failed for %s (entity %d)", classname, extra);
            }
            continue;
        }
        WeaponBuffs_LogProjectile(mimicType >= 0 ? "mimic-extra" : "clone-extra", extra);
    }
}

public void WeaponBuffs_OnTakeDamagePost(int victim, int attacker, int inflictor,
    float damage, int damagetype, int weapon, const float damageForce[3],
    const float damagePosition[3], int damagecustom)
{
    if (victim < 1 || victim > MaxClients || attacker < 1 || attacker > MaxClients
        || victim == attacker || !IsClientInGame(victim) || !IsClientInGame(attacker)
        || GetClientTeam(victim) == GetClientTeam(attacker) || (damagetype & DMG_BURN) != 0)
    {
        return;
    }
    if ((weapon <= MaxClients || !IsValidEntity(weapon)) && inflictor > MaxClients
        && IsValidEntity(inflictor) && HasEntProp(inflictor, Prop_Send, "m_hLauncher"))
    {
        weapon = GetEntPropEnt(inflictor, Prop_Send, "m_hLauncher");
    }
    if ((weapon <= MaxClients || !IsValidEntity(weapon)) && IsPlayerAlive(attacker))
    {
        weapon = GetEntPropEnt(attacker, Prop_Send, "m_hActiveWeapon");
    }
    int catalog = WeaponBuffs_ForEntity(weapon);
    if (catalog >= 0 && g_WeaponEffectLevels[catalog][IgniteEffect] > 0)
    {
        TF2_IgnitePlayer(victim, attacker);
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
    ReplyToCommand(client, "[AP] Projectile debug %s; backend=mimic plus family clone.",
        g_WeaponBuffProjectileDebug.BoolValue ? "ON" : "OFF");
    if (args == 0 && g_WeaponBuffProjectileLineHeld > 0)
    {
        int first = (g_WeaponBuffProjectileLineNext - g_WeaponBuffProjectileLineHeld
            + ProjectileDebugLineCount) % ProjectileDebugLineCount;
        for (int offset = 0; offset < g_WeaponBuffProjectileLineHeld; offset++)
        {
            int line = (first + offset) % ProjectileDebugLineCount;
            ReplyToCommand(client, "%s", g_WeaponBuffProjectileLines[line]);
        }
    }
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
