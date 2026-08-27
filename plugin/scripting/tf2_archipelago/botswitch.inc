/**
 * The Bot Switcher, in the game.
 *
 * The launcher has a tab for this and a player halfway through a mission is not
 * in front of it. !ap bots opens the same team as a menu: a seat, then a class,
 * and the lineup is retyped.
 *
 * No state of its own. The team lives in the bot mod's own convar, which is
 * what the launcher writes and what the mod reads, so the menu edits that
 * string and nothing here can disagree with either of them.
 */

// The lineup convar of OfficerSpy's mod, or null on a server running without it.
static ConVar BotComposition()
{
    return FindConVar("sm_redbots_manager_team_composition");
}

/* The mod's spelling of a class, which is not this plugin's for the Heavy.
 *
 * g_ClassKeys says "heavy" because that is what an Archipelago item is called.
 * The mod says "heavyweapons", the way the game's own tf_bot_add does, and a
 * composition it cannot parse leaves the seat to the lineup mode instead. */
static char[] BotClassKey(TFClassType class)
{
    char key[16];
    if (class == TFClass_Heavy)
    {
        strcopy(key, sizeof(key), "heavyweapons");
        return key;
    }
    strcopy(key, sizeof(key), g_ClassKeys[view_as<int>(class)]);
    return key;
}

// What a menu calls a class, which is not what a convar calls it. Indexed by
// TFClassType, like g_ClassKeys.
static char g_BotClassNames[view_as<int>(TFClass_Engineer) + 1][] = {
    "",
    "Scout",
    "Sniper",
    "Soldier",
    "Demoman",
    "Medic",
    "Heavy",
    "Pyro",
    "Spy",
    "Engineer",
};

/* The menu's name for a class the lineup spells in the mod's way, or an empty
string for anything the mod would not parse either.

A lineup can be typed into the console by hand, so this is the one place that
reads it back and it has to survive a typo rather than print it. */
static char[] BotClassName(const char[] key)
{
    char name[16];
    if (key[0] == '\0')
    {
        return name;
    }
    for (int class = view_as<int>(TFClass_Scout); class <= view_as<int>(TFClass_Engineer); class++)
    {
        if (StrEqual(key, BotClassKey(view_as<TFClassType>(class)), false))
        {
            strcopy(name, sizeof(name), g_BotClassNames[class]);
            return name;
        }
    }
    return name;
}

// How many seats RED holds, which is how many the menu offers.
static int BotSeatCount()
{
    ConVar size = FindConVar("tf_mvm_defenders_team_size");
    return size == null ? 6 : size.IntValue;
}

// The class the lineup puts in this seat, counted from 1, or an empty string
// when it leaves the seat to the mod.
static void BotSeatClass(int seat, char[] buffer, int maxlength)
{
    buffer[0] = '\0';

    ConVar composition = BotComposition();
    if (composition == null)
    {
        return;
    }
    char list[256];
    composition.GetString(list, sizeof(list));

    char seats[MAXPLAYERS + 1][32];
    int total = ExplodeString(list, ",", seats, sizeof(seats), sizeof(seats[]));
    if (seat < 1 || seat > total)
    {
        return;
    }
    TrimString(seats[seat - 1]);
    strcopy(buffer, maxlength, seats[seat - 1]);
}

/* Put a class in one seat and leave every other seat as it stands.

The list is padded out to the seat before it is written, or seat five of a list
naming two would land in seat three. A blank entry is a seat the mod draws,
which is what the padding means and what an empty class writes. */
static void BotSeatSet(int seat, const char[] class)
{
    ConVar composition = BotComposition();
    if (composition == null)
    {
        return;
    }
    char list[256];
    composition.GetString(list, sizeof(list));

    char seats[MAXPLAYERS + 1][32];
    int total = ExplodeString(list, ",", seats, sizeof(seats), sizeof(seats[]));

    char out[256];
    int count = seat > total ? seat : total;
    for (int i = 0; i < count && i < sizeof(seats); i++)
    {
        char entry[32];
        if (i < total)
        {
            strcopy(entry, sizeof(entry), seats[i]);
            TrimString(entry);
        }
        if (i == seat - 1)
        {
            strcopy(entry, sizeof(entry), class);
        }
        Format(out, sizeof(out), "%s%s%s", out, i == 0 ? "" : ",", entry);
    }
    composition.SetString(out);
}

/* The seats, as a menu.
 *
 * The mod's own convar change hook is what acts on a pick: it reseats between
 * waves and holds the change until the break when a wave is running, so nothing
 * here has to know which of the two it is. */
void BotSwitch_Open(int client)
{
    if (BotComposition() == null)
    {
        AP_PrintToClient(client, "This server runs without the defender bots.");
        return;
    }

    Menu menu = new Menu(BotSwitch_SeatChosen);
    menu.SetTitle("Bot Switcher: pick a seat");

    for (int seat = 1; seat <= BotSeatCount(); seat++)
    {
        char class[32];
        BotSeatClass(seat, class, sizeof(class));

        char name[16]; strcopy(name, sizeof(name), BotClassName(class));

        char item[64];
        Format(item, sizeof(item), "Seat %d: %s", seat, name[0] == '\0' ? "the mod picks" : name);

        char choice[8];
        IntToString(seat, choice, sizeof(choice));
        menu.AddItem(choice, item);
    }
    menu.ExitButton = true;
    menu.Display(client, MENU_TIME_FOREVER);
}

public int BotSwitch_SeatChosen(Menu menu, MenuAction action, int client, int item)
{
    if (action == MenuAction_End)
    {
        delete menu;
        return 0;
    }
    if (action != MenuAction_Select)
    {
        return 0;
    }
    char choice[8];
    menu.GetItem(item, choice, sizeof(choice));

    Menu classes = new Menu(BotSwitch_ClassChosen);
    classes.SetTitle("Seat %s plays", choice);

    // The seat number rides on every item: a menu handler is handed the item it
    // was given and nothing else, and holding it per client is a second place
    // for the two to disagree.
    for (int class = view_as<int>(TFClass_Scout); class <= view_as<int>(TFClass_Engineer); class++)
    {
        char key[32];
        Format(key, sizeof(key), "%s:%s", choice, BotClassKey(view_as<TFClassType>(class)));
        classes.AddItem(key, g_BotClassNames[class]);
    }
    char drawn[32];
    Format(drawn, sizeof(drawn), "%s:", choice);
    classes.AddItem(drawn, "let the mod pick");

    classes.ExitBackButton = true;
    classes.Display(client, MENU_TIME_FOREVER);
    return 0;
}

public int BotSwitch_ClassChosen(Menu menu, MenuAction action, int client, int item)
{
    if (action == MenuAction_End)
    {
        delete menu;
        return 0;
    }
    if (action == MenuAction_Cancel && item == MenuCancel_ExitBack)
    {
        BotSwitch_Open(client);
        return 0;
    }
    if (action != MenuAction_Select)
    {
        return 0;
    }
    char key[40];
    menu.GetItem(item, key, sizeof(key));

    /* Both halves or neither. "let the mod pick" is the seat and an empty
    class, which is two parts and not one, so a single part is an item nothing
    here wrote. */
    char parts[2][32];
    if (ExplodeString(key, ":", parts, sizeof(parts), sizeof(parts[])) < 2)
    {
        return 0;
    }
    int seat = StringToInt(parts[0]);
    BotSeatSet(seat, parts[1]);

    char name[16]; strcopy(name, sizeof(name), BotClassName(parts[1]));

    AP_PrintToClient(client, "Seat %d plays %s.", seat, name[0] == '\0' ? "whatever the mod picks" : name);
    BotSwitch_Open(client);
    return 0;
}
