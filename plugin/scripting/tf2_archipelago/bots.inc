/**
 * The defender bots, as this plugin sees them.
 *
 * The bots belong to OfficerSpy's mod. What lives here is what a run needs
 * from them and the mod does not do: a bot steps aside for a human, and what
 * the bots buy can be read in the chat, since the game no longer lets a
 * player inspect a teammate's upgrades.
 */

ConVar g_CvarBotUpgradesChat;
ConVar g_CvarBotsWaitForPlayers;
ConVar g_CvarBotsBackfill;

// Whether the hold is on, so it is said once rather than twice a second.
bool g_Holding;

// How often the ready hold runs. The mod sets a bot ready from several places
// and the hold takes it away again, so it is a poll rather than a gate: one
// gate per call site is one call site away from being missed.
#define BotReadyInterval 0.5

// The upgrade names, in the order scripts/items/mvm_upgrades.txt lists them.
// That order is the index the game's MVM_Upgrade command carries, and the
// file is the only place the names live.
ArrayList g_UpgradeNames;

// Wide enough for any line of mvm_upgrades.txt.
#define UpgradeLineMax 256

// What tf_mvm_defenders_team_size says when the game will not.
#define DefenderTeamSizeDefault 6

void Bots_Init()
{
    g_CvarBotUpgradesChat = CreateConVar("tf2ap_bot_upgrades_chat", "0",
        "Write what the defender bots buy at the upgrade station to the chat.", _, true, 0.0, true, 1.0);
    g_CvarBotsWaitForPlayers = CreateConVar("tf2ap_bots_wait_for_players", "1",
        "Hold the bots unready until every player on RED is ready.", _, true, 0.0, true, 1.0);
    g_CvarBotsBackfill = CreateConVar("tf2ap_bots_backfill", "1",
        "Put a bot back on RED when a player leaves between waves.", _, true, 0.0, true, 1.0);
    g_UpgradeNames = new ArrayList(ByteCountToCells(64));

    // No TIMER_FLAG_NO_MAPCHANGE: the hold must outlive every map change.
    CreateTimer(BotReadyInterval, Timer_HoldBotsReady, _, TIMER_REPEAT);
}

void Bots_OnMapStart()
{
    Bots_LoadUpgradeNames();
}

// Every "attribute" line, in file order. Read through the game's own file
// system, because the file lives inside a VPK rather than on disk.
static void Bots_LoadUpgradeNames()
{
    g_UpgradeNames.Clear();
    File file = OpenFile("scripts/items/mvm_upgrades.txt", "r", true);
    if (file == null)
    {
        AP_Debug("The plugin cannot read mvm_upgrades.txt. Bot purchases show as numbers.");
        return;
    }
    char line[UpgradeLineMax];
    while (file.ReadLine(line, sizeof(line)))
    {
        char key[32], value[64];
        if (!SplitQuoted(line, key, sizeof(key), value, sizeof(value)))
        {
            continue;
        }
        if (StrEqual(key, "attribute", false))
        {
            g_UpgradeNames.PushString(value);
        }
    }
    delete file;
    AP_Debug("The plugin read %d upgrade names.", g_UpgradeNames.Length);
}

// "key"   "value" from one KeyValues line. False for anything else.
static bool SplitQuoted(const char[] line, char[] key, int keyMax, char[] value, int valueMax)
{
    int quotes[4];
    int found = 0;
    for (int index = 0; line[index] != '\0' && found < 4; index++)
    {
        if (line[index] == '"')
        {
            quotes[found++] = index;
        }
    }
    if (found < 4)
    {
        return false;
    }
    int keyLength = quotes[1] - quotes[0];
    int valueLength = quotes[3] - quotes[2];
    strcopy(key, keyLength < keyMax ? keyLength : keyMax, line[quotes[0] + 1]);
    strcopy(value, valueLength < valueMax ? valueLength : valueMax, line[quotes[2] + 1]);
    return true;
}

// The game handles the purchase; this only says what it was. Runs for the
// mod's bots, whose purchases are what nobody can see any more.
public void OnClientCommandKeyValues_Post(int client, KeyValues kv)
{
    if (!g_CvarBotUpgradesChat.BoolValue || client <= 0 || !IsFakeClient(client))
    {
        return;
    }
    char name[32];
    kv.GetSectionName(name, sizeof(name));
    if (!StrEqual(name, "MVM_Upgrade"))
    {
        return;
    }
    if (!kv.JumpToKey("upgrade"))
    {
        return;
    }
    int slot = kv.GetNum("itemslot", -1);
    int index = kv.GetNum("upgrade", -1);
    int count = kv.GetNum("count", 1);
    kv.GoBack();

    // The mod asks for zero charges when a bot's canteen is already full. The
    // game does nothing with it, so neither does the chat.
    if (count == 0)
    {
        return;
    }

    char upgrade[64];
    if (index >= 0 && index < g_UpgradeNames.Length)
    {
        g_UpgradeNames.GetString(index, upgrade, sizeof(upgrade));
    }
    else
    {
        FormatEx(upgrade, sizeof(upgrade), "upgrade %d", index);
    }
    if (count < 0)
    {
        AP_Announce("%N sold %s (%s).", client, upgrade, SlotLabel(slot));
        return;
    }
    AP_Announce("%N bought %s (%s).", client, upgrade, SlotLabel(slot));
}

// The loadout slot MVM_Upgrade carries, which is not the weapon slot: the
// canteen is slot 9, and slot 4 is the Engineer's buildings and the Spy's
// sapper. Everything above three used to read as a canteen, which is how an
// Engineer buying sentry damage filled the chat with canteens.
static char[] SlotLabel(int slot)
{
    char label[16];
    switch (slot)
    {
        case -1: strcopy(label, sizeof(label), "player");
        case 0: strcopy(label, sizeof(label), "primary");
        case 1: strcopy(label, sizeof(label), "secondary");
        case 2: strcopy(label, sizeof(label), "melee");
        case 3: strcopy(label, sizeof(label), "utility");
        case 4: strcopy(label, sizeof(label), "building");
        case 5: strcopy(label, sizeof(label), "pda");
        case 6: strcopy(label, sizeof(label), "watch");
        case 9: strcopy(label, sizeof(label), "canteen");
        default: FormatEx(label, sizeof(label), "slot %d", slot);
    }
    return label;
}

// A human arrived and RED is full of bots. One bot leaves so the human has a
// seat: the mod fills the team but never empties it, and the game refuses a
// player a full team. The mod tops the team back up when a wave begins.
void Bots_MakeRoom()
{
    if (!MvM_IsActive())
    {
        return;
    }
    int size = DefenderTeamSize();
    int onRed = 0;
    int bot = 0;
    for (int client = 1; client <= MaxClients; client++)
    {
        if (!IsClientInGame(client) || GetClientTeam(client) != TeamRed)
        {
            continue;
        }
        onRed++;
        if (IsFakeClient(client))
        {
            bot = client;
        }
    }
    if (onRed < size || bot == 0)
    {
        return;
    }
    AP_Debug("RED is full at %d. A bot leaves for the player who arrived.", onRed);
    KickClient(bot, "Making room for a player");
}

// How many seats RED holds. The game's own number, and six when it will not say.
static int DefenderTeamSize()
{
    ConVar teamSize = FindConVar("tf_mvm_defenders_team_size");
    if (teamSize != null && teamSize.IntValue > 0)
    {
        return teamSize.IntValue;
    }
    return DefenderTeamSizeDefault;
}

// UNVERIFIED m_bPlayerReady on the game rules, which is what the F4 sets and
// what the mod itself reads. Absent, nothing can be held, so nothing is.
static bool IsReady(int client)
{
    if (FindSendPropInfo("CTFGameRulesProxy", "m_bPlayerReady") == -1)
    {
        return true;
    }
    return GameRules_GetProp("m_bPlayerReady", 1, client) != 0;
}

/**
 * The bots do not start the wave. The players do.
 *
 * A bot readies up as soon as it has done its own job, which on a team of one
 * player and five bots means the wave begins while that player is still in the
 * upgrade station. So a bot that is ready is made unready again for as long as
 * any player on RED is not.
 *
 * Nobody on RED is the same answer for a stronger reason: an empty server
 * would otherwise play the run by itself, wave after wave, and check off
 * locations nobody was there for.
 *
 * A poll and not a command listener. The mod lets its own bots through that
 * listener on purpose, and it sets them ready from four places; taking the
 * ready away wherever it came from is one rule instead of four.
 */
public Action Timer_HoldBotsReady(Handle timer)
{
    if (!g_CvarBotsWaitForPlayers.BoolValue || !MvM_IsActive())
    {
        return Plugin_Continue;
    }
    // A wave that is already running has nothing left to hold.
    if (GameRules_GetRoundState() != RoundState_BetweenRounds)
    {
        return Plugin_Continue;
    }
    if (Bots_PlayersReady())
    {
        if (g_Holding)
        {
            g_Holding = false;
            AP_Debug("Every player on RED is ready. The bots are let go.");
        }
        return Plugin_Continue;
    }
    // Edge-triggered: this runs twice a second, and a line each time would be
    // the only thing in the log by the end of the evening.
    if (!g_Holding)
    {
        g_Holding = true;
        AP_Debug("The bots are held unready: %d player(s) on RED, not all ready.", Bots_PlayerCount());
    }
    for (int client = 1; client <= MaxClients; client++)
    {
        if (!IsClientInGame(client) || !IsFakeClient(client)
            || GetClientTeam(client) != TeamRed || !IsReady(client))
        {
            continue;
        }
        // The same command the F4 sends, so the game does its own bookkeeping.
        FakeClientCommand(client, "tournament_player_readystate 0");
    }
    return Plugin_Continue;
}

// How many players are on RED, bots excluded.
static int Bots_PlayerCount()
{
    int players = 0;
    for (int client = 1; client <= MaxClients; client++)
    {
        if (MvM_IsPlayer(client))
        {
            players++;
        }
    }
    return players;
}

// Whether every player on RED is ready. False for an empty RED: no player is
// not the same as every player.
static bool Bots_PlayersReady()
{
    int players = 0;
    for (int client = 1; client <= MaxClients; client++)
    {
        if (!MvM_IsPlayer(client))
        {
            continue;
        }
        players++;
        if (!IsReady(client))
        {
            return false;
        }
    }
    return players > 0;
}

// Whether the client leaving was a player. Read in _Post, where the client is
// already gone and nothing can be asked about it any more.
bool g_Leaving[MAXPLAYERS + 1];

void Bots_OnClientLeaving(int client)
{
    g_Leaving[client] = IsClientInGame(client) && !IsFakeClient(client);
}

void Bots_OnClientLeft(int client)
{
    if (!g_Leaving[client])
    {
        return;
    }
    g_Leaving[client] = false;
    Bots_Backfill();
}

/**
 * A player left, so a bot takes the seat back.
 *
 * The mod fills RED when a wave begins and tops it up every second while the
 * wave runs. Between waves it does neither, which is where a player who leaves
 * during the upgrade station leaves a wave balanced for six to five bodies.
 *
 * Only between waves, for that reason: during a wave the mod is already doing
 * it, and asking twice puts two bots in one seat.
 */
static void Bots_Backfill()
{
    if (!g_CvarBotsBackfill.BoolValue || !MvM_IsActive())
    {
        return;
    }
    if (GameRules_GetRoundState() != RoundState_BetweenRounds)
    {
        return;
    }
    // Mode 0 leaves the bots to an admin, and the mod being absent leaves them
    // to nobody. Neither wants this plugin adding any.
    ConVar mode = FindConVar("sm_redbots_manager_mode");
    if (mode == null || mode.IntValue == 0)
    {
        return;
    }
    int onRed = 0;
    for (int client = 1; client <= MaxClients; client++)
    {
        if (IsClientInGame(client) && GetClientTeam(client) == TeamRed)
        {
            onRed++;
        }
    }
    int missing = DefenderTeamSize() - onRed;
    if (missing <= 0)
    {
        return;
    }
    AP_Debug("RED is down to %d between waves. %d bot(s) take the empty seats.", onRed, missing);
    ServerCommand("sm_addbots %d", missing);
}
