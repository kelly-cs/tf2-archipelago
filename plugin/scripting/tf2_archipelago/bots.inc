/**
 * The defender bots, as this plugin sees them.
 *
 * The bots belong to OfficerSpy's mod. What lives here is what a run needs
 * from them and the mod does not do: a bot steps aside for a human, and what
 * the bots buy can be read in the chat, since the game no longer lets a
 * player inspect a teammate's upgrades.
 */

ConVar g_CvarBotUpgradesChat;

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
    g_UpgradeNames = new ArrayList(ByteCountToCells(64));
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
    int size = DefenderTeamSizeDefault;
    ConVar teamSize = FindConVar("tf_mvm_defenders_team_size");
    if (teamSize != null && teamSize.IntValue > 0)
    {
        size = teamSize.IntValue;
    }
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
