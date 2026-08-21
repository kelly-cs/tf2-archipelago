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
#define TeamBlue 3

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

// UNVERIFIED m_bIsMiniBoss on CTFPlayer, which is what the game calls a giant.
// The defender bot mod reads the same property to pick its targets, so it is
// the community's answer rather than only ours.
//
// Robots are fake clients on BLU. A human is never a giant, and neither is an
// allied bot on RED, so the team is checked before the property is.
stock bool MvM_IsGiant(int client)
{
    if (client <= 0 || client > MaxClients || !IsClientInGame(client)
        || !IsFakeClient(client) || GetClientTeam(client) != TeamBlue)
    {
        return false;
    }
    if (!HasProp(client, "m_bIsMiniBoss"))
    {
        return false;
    }
    return GetEntProp(client, Prop_Send, "m_bIsMiniBoss") != 0;
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
        // Alive, because the money arrives as a pack to walk over and a dead
        // player walks nowhere. Between waves everybody is alive; the check is
        // for the moment after a wave is lost and before the respawn.
        if (MvM_IsPlayer(client) && IsPlayerAlive(client))
        {
            return true;
        }
    }
    return false;
}

/**
 * Pay a bundle into the run.
 *
 * Credits are per player, so a bundle pays each of them fully. Returns how many
 * players it paid, zero for none: the caller holds the grant back rather than
 * acknowledging a payment nobody got.
 *
 * The money goes in as a currency pack the player picks up, which is the path
 * the game itself uses when a robot drops cash. Writing m_nCurrency is the
 * obvious way and it is the wrong one: it puts a number on the player that the
 * game's own bookkeeping never saw, and two things read that bookkeeping rather
 * than the number.
 *
 * The refund button is one. It does not hand back what was spent, it restores
 * the balance the game recorded at wave start: 1200 on screen, 400 recorded,
 * press refund and 800 of bundles are gone. Losing a wave is the other, and it
 * restores the same record, which is where the negative balances came from.
 *
 * A pack is collected by walking over it, so it is dropped at the player's own
 * feet and only between waves, where they are standing at the upgrade station
 * anyway. m_bDistributed stays off: that flag means the team has already been
 * paid, and a pack carrying it pays nobody.
 */
stock int MvM_GrantCredits(int amount)
{
    if (amount <= 0 || !MvM_IsActive())
    {
        return 0;
    }
    int paid = 0;
    for (int client = 1; client <= MaxClients; client++)
    {
        if (!MvM_IsPlayer(client) || !IsPlayerAlive(client))
        {
            continue;
        }
        if (MvM_DropCredits(client, amount))
        {
            paid++;
        }
    }
    if (paid > 0)
    {
        AP_Announce("The run received %d credits for %d player(s).", amount, paid);
    }
    return paid;
}

// One pack, at one player's feet. False when the game will not make one, which
// leaves the grant unpaid and unacknowledged rather than half applied.
static bool MvM_DropCredits(int client, int amount)
{
    int pack = CreateEntityByName("item_currencypack_custom");
    if (pack == -1)
    {
        AP_Error("This server cannot make a currency pack. The bundle waits.");
        return false;
    }
    /* m_nAmount is a data property and not a networked one.
     *
     * Asked for as a send property it throws, and a throw here aborts the
     * whole callback: the bundle was never paid, never acknowledged, and asked
     * for again on the next poll, which left one dead pack per attempt behind
     * it. A play-test made two hundred and forty of them in ten minutes.
     *
     * So the property is checked before it is written, and a pack that cannot
     * be told what it is worth is removed rather than left standing.
     */
    if (!HasEntProp(pack, Prop_Data, "m_nAmount"))
    {
        RemoveEntity(pack);
        AP_Error("This server's currency packs have no m_nAmount. The bundle waits.");
        return false;
    }
    // Before the spawn: the amount is read as the pack comes up, and a pack
    // that spawned with the default is worth what the game felt like.
    SetEntProp(pack, Prop_Data, "m_nAmount", amount);
    DispatchSpawn(pack);
    if (!IsValidEntity(pack))
    {
        AP_Error("The currency pack did not survive its spawn. The bundle waits.");
        return false;
    }

    // At the feet and a little above them, so it lands rather than falling
    // through whatever the player is standing on.
    float origin[3];
    GetClientAbsOrigin(client, origin);
    origin[2] += 8.0;
    TeleportEntity(pack, origin, NULL_VECTOR, NULL_VECTOR);
    AP_Debug("A pack worth %d credits is at %N's feet.", amount, client);
    return true;
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
