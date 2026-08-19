/**
 * Reading the game.
 *
 * The entity properties and event fields here are UNVERIFIED: what the
 * community documents, checked against nothing but a compiler.
 *
 * So nothing is read without asking whether it exists. A missing property is a
 * warning and a fallback, never a native error: SourceMod turns a bad property
 * name into a runtime error on every call, burying the console mid-wave.
 */

// Players are on RED in Mann vs Machine; BLU is the robots, and they are bots.
#define TeamRed 2

// Keeps each missing-property complaint to one line rather than one per frame.
StringMap g_WarnedProps;

void MvM_Init()
{
    g_WarnedProps = new StringMap();
}

static bool HasProp(int entity, const char[] property)
{
    if (entity > 0 && HasEntProp(entity, Prop_Send, property))
    {
        return true;
    }
    bool warned;
    if (!g_WarnedProps.GetValue(property, warned))
    {
        g_WarnedProps.SetValue(property, true);
        AP_Error("This server has no %s. The plugin works without it.", property);
    }
    return false;
}

static int FindObjectiveResource()
{
    return FindEntityByClassname(-1, "tf_objective_resource");
}

// The game rules property is the real answer; the mvm_ map prefix is the fallback if it is absent.
stock bool MvM_IsActive()
{
    if (FindSendPropInfo("CTFGameRulesProxy", "m_bPlayingMannVsMachine") != -1)
    {
        return GameRules_GetProp("m_bPlayingMannVsMachine") != 0;
    }
    char map[PLATFORM_MAX_PATH];
    GetCurrentMap(map, sizeof(map));
    return strncmp(map, "mvm_", 4, false) == 0;
}

// Humans on RED only: the robots are bots on BLU, and stripping one would take the wave apart.
stock bool MvM_IsPlayer(int client)
{
    return client > 0 && client <= MaxClients
        && IsClientInGame(client)
        && !IsFakeClient(client)
        && GetClientTeam(client) == TeamRed;
}

// UNVERIFIED m_iszMvMPopfileName: scripts/population/mvm_decoy.pop, trimmed to mvm_decoy.
stock bool MvM_PopFile(char[] buffer, int maxlength)
{
    buffer[0] = '\0';

    int resource = FindObjectiveResource();
    if (resource != -1 && HasProp(resource, "m_iszMvMPopfileName"))
    {
        char raw[PLATFORM_MAX_PATH];
        if (GetEntPropString(resource, Prop_Send, "m_iszMvMPopfileName", raw, sizeof(raw)) > 0)
        {
            TrimPopFileName(raw, sizeof(raw));
            strcopy(buffer, maxlength, raw);
        }
    }
    if (buffer[0] == '\0')
    {
        GetCurrentMap(buffer, maxlength);
    }
    return buffer[0] != '\0';
}

static void TrimPopFileName(char[] name, int maxlength)
{
    int lastSlash = FindCharInString(name, '/', true);
    if (lastSlash == -1)
    {
        lastSlash = FindCharInString(name, '\\', true);
    }
    if (lastSlash != -1)
    {
        strcopy(name, maxlength, name[lastSlash + 1]);
    }
    int extension = FindCharInString(name, '.', true);
    if (extension != -1)
    {
        name[extension] = '\0';
    }
}

// UNVERIFIED m_nMannVsMachineWaveCount: the cross-check, mvm_begin_wave is the primary source.
stock int MvM_WaveFromGame()
{
    int resource = FindObjectiveResource();
    if (resource == -1 || !HasProp(resource, "m_nMannVsMachineWaveCount"))
    {
        return 0;
    }
    return GetEntProp(resource, Prop_Send, "m_nMannVsMachineWaveCount");
}

// UNVERIFIED m_nMannVsMachineMaxWaveCount. Zero when the game will not say.
stock int MvM_MaxWavesFromGame()
{
    int resource = FindObjectiveResource();
    if (resource == -1 || !HasProp(resource, "m_nMannVsMachineMaxWaveCount"))
    {
        return 0;
    }
    return GetEntProp(resource, Prop_Send, "m_nMannVsMachineMaxWaveCount");
}

// Whether credits paid now would survive. A wave that is lost takes the team
// back to where the wave began, and money paid into a wave goes back with it,
// which is what a play-test saw a Cash Bundle do. Between waves is where the
// upgrade station is anyway, so nothing is lost by waiting for it.
//
// Nobody on RED is the other unpayable moment: a bundle that arrives with an
// empty server used to pay nobody and was gone.
stock bool MvM_CanPayCredits()
{
    if (!MvM_IsActive() || GameRules_GetRoundState() != RoundState_BetweenRounds)
    {
        return false;
    }
    for (int client = 1; client <= MaxClients; client++)
    {
        if (MvM_IsPlayer(client) && HasProp(client, "m_nCurrency"))
        {
            return true;
        }
    }
    return false;
}

// UNVERIFIED m_nCurrency on CTFPlayer. Credits are per player, so a bundle pays
// each of them fully. Returns how many players it paid, zero for none: the
// caller holds the grant back rather than acknowledging a payment nobody got.
stock int MvM_GrantCredits(int amount)
{
    if (amount <= 0 || !MvM_IsActive())
    {
        return 0;
    }
    int paid = 0;
    for (int client = 1; client <= MaxClients; client++)
    {
        if (!MvM_IsPlayer(client) || !HasProp(client, "m_nCurrency"))
        {
            continue;
        }
        int held = GetEntProp(client, Prop_Send, "m_nCurrency");
        SetEntProp(client, Prop_Send, "m_nCurrency", held + amount);
        paid++;
    }
    if (paid > 0)
    {
        AP_Announce("The run received %d credits for %d player(s).", amount, paid);
    }
    return paid;
}

// Everyone on RED, bots included. The kill is the whole effect: no wave-failed
// event is fired, and an undefended hatch is what loses the wave.
stock int MvM_KillTeam()
{
    int killed = 0;
    for (int client = 1; client <= MaxClients; client++)
    {
        if (IsClientInGame(client) && GetClientTeam(client) == TeamRed && IsPlayerAlive(client))
        {
            ForcePlayerSuicide(client);
            killed++;
        }
    }
    return killed;
}
