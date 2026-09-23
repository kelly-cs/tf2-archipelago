/**
 * Isolated MvM wave smoke probe. Install only in a disposable test server.
 * It does not advance waves: only the game's mvm_wave_complete event passes a
 * test. Enemy bots and tanks are defeated 15-25 seconds after first sighting.
 */
#pragma semicolon 1
#pragma newdecls required

#include <sourcemod>
#include <sdktools>
#include <sdkhooks>
#include <tf2>
#include <tf2_stocks>

public Plugin myinfo =
{
    name = "TF2 MvM wave probe",
    author = "tf2-archipelago",
    description = "Disposable, unattended wave-completion smoke test",
    version = "0.1.0"
};

#define PROBE_TICK 0.25
#define PROBE_MAX_TANKS 64

enum ProbeState
{
    Probe_Idle,
    Probe_Armed,
    Probe_Running,
    Probe_Passed,
    Probe_Failed
};

ProbeState g_State;
int g_ExpectedWave;
int g_ObservedWave;
int g_Seed;
int g_Defender;
TFClassType g_DefenderClass = TFClass_Scout;
int g_PlayerTeam = 2;
int g_EnemyTeam = 3;
int g_SpawnSerial;
int g_BotKills;
int g_TankKills;
int g_BotSpawns;
int g_TankSpawns;
int g_KillAttempts;
int g_InitialEnemies;
char g_FailureReason[32];
int g_BotUserId[MAXPLAYERS + 1];
bool g_BotKillPending[MAXPLAYERS + 1];
float g_BotDeadline[MAXPLAYERS + 1];
int g_TankRef[PROBE_MAX_TANKS];
int g_TankIndex[PROBE_MAX_TANKS];
bool g_TankKillPending[PROBE_MAX_TANKS];
float g_TankDeadline[PROBE_MAX_TANKS];
float g_ArmedAt;
float g_StartedAt;

public void OnPluginStart()
{
    RegAdminCmd("sm_waveprobe_arm", Command_Arm, ADMFLAG_ROOT,
        "Arm a wave test: sm_waveprobe_arm <wave> [seed]");
    RegAdminCmd("sm_waveprobe_status", Command_Status, ADMFLAG_ROOT,
        "Print machine-readable wave test state");
    RegAdminCmd("sm_waveprobe_debug", Command_Debug, ADMFLAG_ROOT,
        "Print a snapshot of the wave and living invaders");
    RegAdminCmd("sm_waveprobe_reset", Command_Reset, ADMFLAG_ROOT,
        "Stop the wave test, retaining its fake player client");
    RegAdminCmd("sm_waveprobe_wake", Command_Wake, ADMFLAG_ROOT,
        "Create a fake player: sm_waveprobe_wake <red|blue>");
    HookEvent("mvm_begin_wave", Event_BeginWave);
    HookEvent("mvm_wave_complete", Event_WaveComplete);
    HookEvent("mvm_wave_failed", Event_WaveFailed);
    HookEvent("player_spawn", Event_PlayerSpawn);
    HookEvent("player_death", Event_PlayerDeath);
    CreateTimer(PROBE_TICK, Timer_Probe, _, TIMER_REPEAT);
}

public void OnMapStart()
{
    ResetProbe();
}

public void OnPluginEnd()
{
    if (g_Defender > 0 && IsClientConnected(g_Defender))
    {
        KickClient(g_Defender, "wave probe stopped");
    }
}

public void OnClientDisconnect(int client)
{
    if (client == g_Defender) g_Defender = 0;
    g_BotUserId[client] = 0;
    g_BotKillPending[client] = false;
    g_BotDeadline[client] = 0.0;
}

public void OnClientPutInServer(int client)
{
    SDKHook(client, SDKHook_OnTakeDamage, DefenderDamage);
}

static void ResetProbe()
{
    g_State = Probe_Idle;
    g_ExpectedWave = 0;
    g_ObservedWave = 0;
    g_SpawnSerial = 0;
    g_BotKills = 0;
    g_TankKills = 0;
    g_BotSpawns = 0;
    g_TankSpawns = 0;
    g_KillAttempts = 0;
    g_InitialEnemies = 0;
    g_FailureReason[0] = '\0';
    g_ArmedAt = 0.0;
    g_StartedAt = 0.0;
    for (int client = 1; client <= MaxClients; client++)
    {
        g_BotUserId[client] = 0;
        g_BotKillPending[client] = false;
        g_BotDeadline[client] = 0.0;
    }
    for (int i = 0; i < PROBE_MAX_TANKS; i++)
    {
        g_TankRef[i] = INVALID_ENT_REFERENCE;
        g_TankIndex[i] = -1;
        g_TankKillPending[i] = false;
        g_TankDeadline[i] = 0.0;
    }
}

public Action Command_Arm(int client, int argc)
{
    if (argc < 1)
    {
        ReplyToCommand(client, "usage: sm_waveprobe_arm <wave> [seed]");
        return Plugin_Handled;
    }
    char arg[24];
    GetCmdArg(1, arg, sizeof(arg));
    int wave = StringToInt(arg);
    if (wave < 1)
    {
        ReplyToCommand(client, "wave must be positive");
        return Plugin_Handled;
    }
    ResetProbe();
    g_ExpectedWave = wave;
    g_Seed = 1;
    if (argc >= 2)
    {
        GetCmdArg(2, arg, sizeof(arg));
        g_Seed = StringToInt(arg);
    }
    g_State = Probe_Armed;
    g_ArmedAt = GetGameTime();
    EnsureDefender();
    ReplyToCommand(client, "WAVEPROBE armed wave=%d seed=%d", wave, g_Seed);
    return Plugin_Handled;
}

public Action Command_Reset(int client, int argc)
{
    ResetProbe();
    ReplyToCommand(client, "WAVEPROBE reset");
    return Plugin_Handled;
}

public Action Command_Wake(int client, int argc)
{
    if (argc < 1)
    {
        ReplyToCommand(client, "usage: sm_waveprobe_wake <red|blue> [scout|medic]");
        return Plugin_Handled;
    }
    char team[16];
    GetCmdArg(1, team, sizeof(team));
    if (StrEqual(team, "blue", false))
    {
        g_PlayerTeam = 3;
        g_EnemyTeam = 2;
    }
    else if (StrEqual(team, "red", false))
    {
        g_PlayerTeam = 2;
        g_EnemyTeam = 3;
    }
    else
    {
        ReplyToCommand(client, "usage: sm_waveprobe_wake <red|blue> [scout|medic]");
        return Plugin_Handled;
    }
    g_DefenderClass = TFClass_Scout;
    if (argc >= 2)
    {
        char playerClass[16];
        GetCmdArg(2, playerClass, sizeof(playerClass));
        if (StrEqual(playerClass, "medic", false))
        {
            g_DefenderClass = TFClass_Medic;
        }
        else if (!StrEqual(playerClass, "scout", false))
        {
            ReplyToCommand(client, "usage: sm_waveprobe_wake <red|blue> [scout|medic]");
            return Plugin_Handled;
        }
    }
    EnsureDefender();
    ReplyToCommand(client, "WAVEPROBE defender=%d playerteam=%d enemyteam=%d class=%d",
        g_Defender, g_PlayerTeam, g_EnemyTeam, view_as<int>(g_DefenderClass));
    return Plugin_Handled;
}

static void EnsureDefender()
{
    if (g_Defender > 0 && IsClientConnected(g_Defender))
    {
        if (!IsClientInGame(g_Defender)) return;
        if (GetClientTeam(g_Defender) != g_PlayerTeam)
            ChangeClientTeam(g_Defender, g_PlayerTeam);
        if (TF2_GetPlayerClass(g_Defender) != g_DefenderClass)
        {
            TF2_SetPlayerClass(g_Defender, g_DefenderClass);
            TF2_RespawnPlayer(g_Defender);
        }
        return;
    }
    g_Defender = CreateFakeClient("Wave Probe Player");
    if (g_Defender == 0)
    {
        g_State = Probe_Failed;
        strcopy(g_FailureReason, sizeof(g_FailureReason), "defender_create");
        LogError("WAVEPROBE cannot create a fake player client");
    }
}

public Action DefenderDamage(int victim, int &attacker, int &inflictor,
    float &damage, int &damageType)
{
    // The probe measures invader wave progression. Protecting RED also keeps
    // authored escort NPCs such as Remedic's Chief Medic alive while we kill
    // BLU robots, so a survival objective does not invalidate that measure.
    if ((g_State == Probe_Armed || g_State == Probe_Running)
        && IsClientInGame(victim) && GetClientTeam(victim) == g_PlayerTeam)
    {
        return Plugin_Handled;
    }
    return Plugin_Continue;
}

static void StateName(ProbeState state, char[] buffer, int length)
{
    switch (state)
    {
        case Probe_Armed: strcopy(buffer, length, "armed");
        case Probe_Running: strcopy(buffer, length, "running");
        case Probe_Passed: strcopy(buffer, length, "passed");
        case Probe_Failed: strcopy(buffer, length, "failed");
        default: strcopy(buffer, length, "idle");
    }
}

public Action Command_Status(int client, int argc)
{
    char map[64];
    char state[16];
    char pop[PLATFORM_MAX_PATH];
    int maxWaves = 0;
    int gameWave = 0;
    int redClients = 0;
    int blueClients = 0;
    int alive = 0;
    int remaining = -1;
    GetCurrentMap(map, sizeof(map));
    StateName(g_State, state, sizeof(state));
    strcopy(pop, sizeof(pop), "unknown");
    int resource = FindEntityByClassname(-1, "tf_objective_resource");
    if (resource != -1)
    {
        if (HasEntProp(resource, Prop_Send, "m_iszMvMPopfileName"))
        {
            GetEntPropString(resource, Prop_Send, "m_iszMvMPopfileName", pop, sizeof(pop));
        }
        if (HasEntProp(resource, Prop_Send, "m_nMannVsMachineMaxWaveCount"))
        {
            maxWaves = GetEntProp(resource, Prop_Send, "m_nMannVsMachineMaxWaveCount");
        }
        if (HasEntProp(resource, Prop_Send, "m_nMannVsMachineWaveCount"))
        {
            gameWave = GetEntProp(resource, Prop_Send, "m_nMannVsMachineWaveCount");
        }
        if (HasEntProp(resource, Prop_Send, "m_nMannVsMachineWaveEnemyCount"))
        {
            remaining = GetEntProp(resource, Prop_Send, "m_nMannVsMachineWaveEnemyCount");
            if (g_State == Probe_Running && remaining > g_InitialEnemies)
            {
                g_InitialEnemies = remaining;
            }
        }
    }
    ReplaceString(pop, sizeof(pop), "scripts/population/", "");
    ReplaceString(pop, sizeof(pop), ".pop", "");
    for (int player = 1; player <= MaxClients; player++)
    {
        if (!IsClientInGame(player))
        {
            continue;
        }
        if (GetClientTeam(player) == 2) redClients++;
        if (GetClientTeam(player) == 3) blueClients++;
        if (GetClientTeam(player) == g_EnemyTeam && IsPlayerAlive(player)) alive++;
    }
    int tank = -1;
    while ((tank = FindEntityByClassname(tank, "tank_boss")) != -1) alive++;
    float progress = -1.0;
    if (g_InitialEnemies > 0 && remaining >= 0)
    {
        progress = 100.0 * float(g_InitialEnemies - remaining) / float(g_InitialEnemies);
        if (progress < 0.0) progress = 0.0;
        if (progress > 100.0) progress = 100.0;
    }
    ReplyToCommand(client,
        "WAVEPROBE state=%s reason=%s map=%s pop=%s max=%d gamewave=%d expected=%d observed=%d botspawns=%d tankspawns=%d bots=%d tanks=%d attempts=%d alive=%d remaining=%d initial=%d progress=%.1f red=%d blue=%d defender=%d defteam=%d defclass=%d playerteam=%d enemyteam=%d elapsed=%.1f",
        state, g_FailureReason[0] == '\0' ? "none" : g_FailureReason,
        map, pop, maxWaves, gameWave, g_ExpectedWave,
        g_ObservedWave, g_BotSpawns, g_TankSpawns, g_BotKills, g_TankKills,
        g_KillAttempts, alive, remaining, g_InitialEnemies, progress,
        redClients, blueClients,
        g_Defender, g_Defender > 0 && IsClientInGame(g_Defender) ? GetClientTeam(g_Defender) : 0,
        g_Defender > 0 && IsClientInGame(g_Defender) ? view_as<int>(TF2_GetPlayerClass(g_Defender)) : 0,
        g_PlayerTeam, g_EnemyTeam,
        GetGameTime() - (g_StartedAt > 0.0 ? g_StartedAt : g_ArmedAt));
    return Plugin_Handled;
}

public Action Command_Debug(int client, int argc)
{
    int resource = FindEntityByClassname(-1, "tf_objective_resource");
    int wave = resource == -1 ? -1 : GetEntProp(resource, Prop_Send, "m_nMannVsMachineWaveCount");
    int remaining = resource == -1 ? -1 : GetEntProp(resource, Prop_Send, "m_nMannVsMachineWaveEnemyCount");
    ConVar timescale = FindConVar("host_timescale");
    ReplyToCommand(client,
        "WAVEPROBE_DEBUG state=%d wave=%d expected=%d game=%.1f engine=%.1f timescale=%.1f remaining=%d spawns=%d kills=%d attempts=%d defender=%d",
        view_as<int>(g_State), wave, g_ExpectedWave, GetGameTime(), GetEngineTime(),
        timescale == null ? -1.0 : timescale.FloatValue, remaining,
        g_BotSpawns + g_TankSpawns, g_BotKills + g_TankKills, g_KillAttempts, g_Defender);
    int listed;
    int alive;
    for (int bot = 1; bot <= MaxClients; bot++)
    {
        if (!IsClientInGame(bot) || GetClientTeam(bot) != g_EnemyTeam || !IsPlayerAlive(bot))
        {
            continue;
        }
        alive++;
        if (listed++ >= 12) continue;
        char name[64];
        float origin[3];
        GetClientName(bot, name, sizeof(name));
        GetClientAbsOrigin(bot, origin);
        ReplyToCommand(client,
            "WAVEPROBE_BOT client=%d userid=%d class=%d hp=%d timer=%d deadline=%.1f origin=%.0f,%.0f,%.0f name=%s",
            bot, GetClientUserId(bot), view_as<int>(TF2_GetPlayerClass(bot)), GetClientHealth(bot),
            IsScriptedTimerBot(bot), g_BotDeadline[bot] - GetGameTime(),
            origin[0], origin[1], origin[2], name);
    }
    ReplyToCommand(client, "WAVEPROBE_DEBUG_END alive=%d listed=%d", alive, listed < 12 ? listed : 12);
    return Plugin_Handled;
}

public void Event_BeginWave(Event event, const char[] name, bool dontBroadcast)
{
    if (g_State != Probe_Armed)
    {
        return;
    }
    g_ObservedWave = event.GetInt("wave_index") + 1;
    if (g_ObservedWave != g_ExpectedWave)
    {
        g_State = Probe_Failed;
        strcopy(g_FailureReason, sizeof(g_FailureReason), "wrong_wave");
        LogError("WAVEPROBE wrong wave: expected %d, started %d",
            g_ExpectedWave, g_ObservedWave);
        return;
    }
    if (g_Defender == 0 || !IsClientInGame(g_Defender)
        || GetClientTeam(g_Defender) != g_PlayerTeam)
    {
        g_State = Probe_Failed;
        strcopy(g_FailureReason, sizeof(g_FailureReason), "defender_missing");
        LogError("WAVEPROBE fake player is not on expected team %d", g_PlayerTeam);
        return;
    }
    g_State = Probe_Running;
    g_StartedAt = GetGameTime();
    int resource = FindEntityByClassname(-1, "tf_objective_resource");
    if (resource != -1 && HasEntProp(resource, Prop_Send, "m_nMannVsMachineWaveEnemyCount"))
    {
        g_InitialEnemies = GetEntProp(resource, Prop_Send, "m_nMannVsMachineWaveEnemyCount");
    }
    LogMessage("WAVEPROBE started wave=%d seed=%d", g_ObservedWave, g_Seed);
}

public void Event_WaveComplete(Event event, const char[] name, bool dontBroadcast)
{
    if (g_State != Probe_Running)
    {
        return;
    }
    g_State = Probe_Passed;
    LogMessage("WAVEPROBE passed wave=%d seconds=%.1f bots=%d tanks=%d",
        g_ObservedWave, GetGameTime() - g_StartedAt, g_BotKills, g_TankKills);
}

public void Event_WaveFailed(Event event, const char[] name, bool dontBroadcast)
{
    if (g_State == Probe_Running)
    {
        g_State = Probe_Failed;
        strcopy(g_FailureReason, sizeof(g_FailureReason), "wave_failed");
        LogError("WAVEPROBE game reported wave failed");
    }
}

public void Event_PlayerDeath(Event event, const char[] name, bool dontBroadcast)
{
    int bot = GetClientOfUserId(event.GetInt("userid"));
    if (bot > 0 && g_BotKillPending[bot])
    {
        g_BotKills++;
        g_BotKillPending[bot] = false;
    }
    if (bot > 0)
    {
        g_BotUserId[bot] = 0;
        g_BotDeadline[bot] = 0.0;
    }
}

public void Event_PlayerSpawn(Event event, const char[] name, bool dontBroadcast)
{
    int bot = GetClientOfUserId(event.GetInt("userid"));
    if (bot > 0) RegisterEnemyBot(bot);
}

static void RegisterEnemyBot(int bot)
{
    if (g_State != Probe_Running || bot == g_Defender || !IsClientInGame(bot)
        || GetClientTeam(bot) != g_EnemyTeam)
    {
        return;
    }
    int userid = GetClientUserId(bot);
    if (g_BotUserId[bot] == userid) return;
    g_BotUserId[bot] = userid;
    g_BotDeadline[bot] = GetGameTime() + KillDelay();
    g_BotSpawns++;
    g_BotKillPending[bot] = false;
    CreateTimer(0.1, Timer_MarkScriptedBot, userid, TIMER_FLAG_NO_MAPCHANGE);
}

// SigMod's TimerBot is a wave clock, not an enemy the player is meant to
// defeat. Its "timer" tag is visible to VScript, so mark it through the same
// RunScriptCode input the authored missions use and leave it alive while the
// real room groups are cleared. The game retires it at wave completion.
public Action Timer_MarkScriptedBot(Handle timer, any userid)
{
    int bot = GetClientOfUserId(userid);
    if (bot > 0 && IsClientInGame(bot) && GetClientTeam(bot) == g_EnemyTeam)
    {
        SetVariantString("if (self.HasBotTag(\"timer\")) self.AcceptInput(\"AddOutput\", \"targetname waveprobe_timer\", null, null)");
        AcceptEntityInput(bot, "RunScriptCode");
    }
    return Plugin_Stop;
}

static bool IsScriptedTimerBot(int bot)
{
    char name[64];
    GetEntPropString(bot, Prop_Data, "m_iName", name, sizeof(name));
    return StrEqual(name, "waveprobe_timer");
}

public void OnEntityCreated(int entity, const char[] classname)
{
    if (g_State == Probe_Running && StrEqual(classname, "tank_boss"))
    {
        RegisterTank(entity);
    }
}

static int RegisterTank(int tank)
{
    int ref = EntIndexToEntRef(tank);
    int slot = -1;
    for (int i = 0; i < PROBE_MAX_TANKS; i++)
    {
        if (g_TankRef[i] == ref) return i;
        if (slot < 0 && (g_TankRef[i] == INVALID_ENT_REFERENCE
            || EntRefToEntIndex(g_TankRef[i]) == INVALID_ENT_REFERENCE))
        {
            slot = i;
        }
    }
    if (slot < 0)
    {
        g_State = Probe_Failed;
        strcopy(g_FailureReason, sizeof(g_FailureReason), "tank_overflow");
        LogError("WAVEPROBE more than %d active tanks", PROBE_MAX_TANKS);
        return -1;
    }
    g_TankRef[slot] = ref;
    g_TankIndex[slot] = tank;
    g_TankDeadline[slot] = GetGameTime() + KillDelay();
    g_TankSpawns++;
    g_TankKillPending[slot] = false;
    return slot;
}

public void OnEntityDestroyed(int entity)
{
    for (int i = 0; i < PROBE_MAX_TANKS; i++)
    {
        if (g_TankKillPending[i] && g_TankIndex[i] == entity)
        {
            g_TankKills++;
            g_TankKillPending[i] = false;
            g_TankRef[i] = INVALID_ENT_REFERENCE;
            g_TankIndex[i] = -1;
            return;
        }
    }
}

// The spawn ordinal and seed determine a stable 15-25 second delay. Random
// runtime state would make a failed mission harder to reproduce.
static float KillDelay()
{
    g_SpawnSerial++;
    int value = (g_SpawnSerial * 1103515245 + g_Seed * 12345) & 0x7fffffff;
    return 15.0 + float(value % 1001) / 100.0;
}

public Action Timer_Probe(Handle timer)
{
    if (g_Defender > 0) EnsureDefender();
    if (g_State != Probe_Armed && g_State != Probe_Running)
    {
        return Plugin_Continue;
    }
    EnsureDefender();
    if (g_Defender > 0 && IsClientInGame(g_Defender)
        && !IsPlayerAlive(g_Defender))
    {
        TF2_RespawnPlayer(g_Defender);
    }
    if (g_State != Probe_Running)
    {
        return Plugin_Continue;
    }

    float now = GetGameTime();
    for (int bot = 1; bot <= MaxClients; bot++)
    {
        if (bot == g_Defender || !IsClientInGame(bot)
            || GetClientTeam(bot) != g_EnemyTeam || !IsPlayerAlive(bot))
        {
            g_BotUserId[bot] = 0;
            g_BotKillPending[bot] = false;
            continue;
        }
        RegisterEnemyBot(bot);
        if (!IsScriptedTimerBot(bot) && now >= g_BotDeadline[bot])
        {
            g_BotKillPending[bot] = true;
            g_KillAttempts++;
            ForcePlayerSuicide(bot);
            g_BotDeadline[bot] = now + 5.0;
        }
    }

    int tank = -1;
    while ((tank = FindEntityByClassname(tank, "tank_boss")) != -1)
    {
        int slot = RegisterTank(tank);
        if (slot < 0) return Plugin_Continue;
        if (now >= g_TankDeadline[slot])
        {
            g_TankKillPending[slot] = true;
            g_KillAttempts++;
            SDKHooks_TakeDamage(tank, g_Defender, g_Defender, 1000000.0);
            g_TankDeadline[slot] = now + 5.0;
        }
    }
    return Plugin_Continue;
}
