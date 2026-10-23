/** Remove invaders that never leave their protected spawn area.
 *
 * TFCond_UberchargedHidden is the game's out-of-bounds robot protection. It
 * does not include Medic ÜberCharge or the ordinary post-spawn invulnerability.
 * A continuously protected robot can hold a WaveSpawn open forever even when
 * every reachable enemy is dead. Only a live BLU fake client in an active wave
 * is eligible; RED defenders and humans are never touched.
 */

#define InvaderStallPollInterval 1.0

ConVar g_InvaderStallTimeout;

bool g_InvaderStallWaveRunning;
int g_InvaderStallUserId[MAXPLAYERS + 1];
float g_InvaderStallSince[MAXPLAYERS + 1];

void InvaderStalls_Init()
{
    g_InvaderStallTimeout = CreateConVar("tf2ap_invader_stall_timeout", "120",
        "Seconds a BLU robot may remain protected in spawn during a wave; 0 disables the guard.",
        _, true, 0.0);
    CreateTimer(InvaderStallPollInterval, Timer_InvaderStalls, _, TIMER_REPEAT);
}

void InvaderStalls_Reset()
{
    g_InvaderStallWaveRunning = false;
    for (int client = 1; client <= MaxClients; client++)
    {
        InvaderStalls_ResetClient(client);
    }
}

void InvaderStalls_BeginWave()
{
    InvaderStalls_Reset();
    g_InvaderStallWaveRunning = true;
}

void InvaderStalls_ResetClient(int client)
{
    if (client < 1 || client > MaxClients)
    {
        return;
    }
    g_InvaderStallUserId[client] = 0;
    g_InvaderStallSince[client] = 0.0;
}

public Action Timer_InvaderStalls(Handle timer)
{
    float timeout = g_InvaderStallTimeout.FloatValue;
    if (timeout <= 0.0 || !g_InvaderStallWaveRunning || !MvM_IsActive()
        || GameRules_GetRoundState() != RoundState_RoundRunning)
    {
        return Plugin_Continue;
    }

    float now = GetGameTime();
    for (int client = 1; client <= MaxClients; client++)
    {
        if (!IsClientInGame(client) || !IsFakeClient(client)
            || GetClientTeam(client) != TeamBlue || !IsPlayerAlive(client)
            || !TF2_IsPlayerInCondition(client, TFCond_UberchargedHidden))
        {
            InvaderStalls_ResetClient(client);
            continue;
        }

        int userid = GetClientUserId(client);
        if (g_InvaderStallUserId[client] != userid || g_InvaderStallSince[client] == 0.0)
        {
            g_InvaderStallUserId[client] = userid;
            g_InvaderStallSince[client] = now;
            continue;
        }
        if (now - g_InvaderStallSince[client] < timeout)
        {
            continue;
        }

        char name[MAX_NAME_LENGTH];
        float position[3];
        GetClientName(client, name, sizeof(name));
        GetClientAbsOrigin(client, position);
        LogMessage("Removed BLU robot %s after %.0fs protected in spawn at %.0f %.0f %.0f.",
            name, now - g_InvaderStallSince[client], position[0], position[1], position[2]);
        // A normal death lets the population manager account for this bot and
        // continue its WaveSpawn. A failed suicide can be retried after another
        // timeout without printing the same message every second.
        g_InvaderStallSince[client] = now;
        ForcePlayerSuicide(client);
    }
    return Plugin_Continue;
}
