/**
 * Saying things.
 *
 * There is no client mod, so chat, the server console and the SourceMod log
 * are the whole diagnostic surface (ADR 0002). This plugin gets debugged on a
 * live server by reading chat: a failure nobody can see is blamed on the game.
 */

#define AP_TAG "\x07FFD700[AP]\x01 "
#define AP_TAG_ERROR "\x07FF4040[AP]\x01 "

// How long a line may be before the tag goes in front of it. The engine refuses
// a user message over 255 bytes and drops it without telling anyone, which is a
// failure that looks exactly like a plugin that never ran. Same number as the
// bridge's chat.LineMax, because both sides build a line for this chat.
#define ChatLineMax 200

ConVar g_CvarAnnounce;
ConVar g_CvarDebug;

// What tf2ap_debug is worth. Chat is the loud one, so it is not the default:
// the log costs nothing anybody reads and it is what a debug bundle carries.
#define DebugOff 0
#define DebugLog 1
#define DebugChat 2

void Log_Init()
{
    g_CvarAnnounce = CreateConVar("tf2ap_announce", "1",
        "Write the received items and the unlocks to the chat.", _, true, 0.0, true, 1.0);
    g_CvarDebug = CreateConVar("tf2ap_debug", "1",
        "Every bridge call and game event: 0 writes none, 1 writes them to the console and the log, 2 writes them to the chat as well.",
        _, true, 0.0, true, 2.0);
}

stock void AP_Announce(const char[] format, any ...)
{
    char message[ChatLineMax];
    VFormat(message, sizeof(message), format, 2);
    if (g_CvarAnnounce.BoolValue)
    {
        PrintToChatAll("%s%s", AP_TAG, message);
    }
    PrintToServer("[AP] %s", message);
}

// Reaches chat even when announcements are muted: players notice missing checks before anyone else.
stock void AP_Error(const char[] format, any ...)
{
    char message[ChatLineMax];
    VFormat(message, sizeof(message), format, 2);
    PrintToChatAll("%s%s", AP_TAG_ERROR, message);
    PrintToServer("[AP] error: %s", message);
    LogError("%s", message);
}

/* Every bridge call and every game event.
 *
 * The default writes them to the console and the SourceMod log and not to the
 * chat. Nobody wants to play a wave through a running commentary, and the two
 * files are what a player sends when something goes wrong: a bundle from a
 * default install is diagnosable without asking that player to reproduce the
 * bug a second time with a convar turned on.
 *
 * Cheap enough to leave on, because every call site here is an event: a wave,
 * a grant, a request that failed. None of them is on a timer. */
stock void AP_Debug(const char[] format, any ...)
{
    int level = g_CvarDebug.IntValue;
    if (level <= DebugOff)
    {
        return;
    }
    char message[ChatLineMax];
    VFormat(message, sizeof(message), format, 2);
    if (level >= DebugChat)
    {
        PrintToChatAll("%s\x07AAAAAA%s", AP_TAG, message);
    }
    PrintToServer("[AP] debug: %s", message);
    LogMessage("%s", message);
}

// Answers the console as readily as a player. Client 0 is the server console,
// which is where an rcon command runs, and a player can leave while a request
// is in flight; PrintToChat throws on either and takes the whole callback with
// it.
stock void AP_PrintToClient(int client, const char[] format, any ...)
{
    char message[ChatLineMax];
    VFormat(message, sizeof(message), format, 3);
    if (client > 0 && client <= MaxClients && IsClientInGame(client))
    {
        PrintToChat(client, "%s%s", AP_TAG, message);
        return;
    }
    PrintToServer("[AP] %s", message);
}
