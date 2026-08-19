/**
 * The unlock set, and what it does to players.
 *
 * The bridge owns the set. The plugin holds a copy for as long as the map
 * lasts, asks again after every reload or map change, and never writes back.
 */

enum
{
    Slot_Primary = 0,
    Slot_Secondary,
    Slot_Melee,
    Slot_Count,
};

// False until the bridge answers once; enforcement is a no-op until then, too much kit beats none.
bool g_HaveUnlocks;

bool g_ClassUnlocked[view_as<int>(TFClass_Engineer) + 1];
bool g_SlotUnlocked[Slot_Count];
StringMap g_UnlockedMissions;

// Indexed by TFClassType, whose order is not the class menu's order. Checked against tf2.inc.
char g_ClassKeys[view_as<int>(TFClass_Engineer) + 1][] = {
    "",
    "scout",
    "sniper",
    "soldier",
    "demoman",
    "medic",
    "heavy",
    "pyro",
    "spy",
    "engineer",
};

char g_SlotKeys[Slot_Count][] = { "primary", "secondary", "melee" };

void Unlocks_Init()
{
    g_UnlockedMissions = new StringMap();
}

void Unlocks_Clear()
{
    g_HaveUnlocks = false;
    for (int class = 0; class < sizeof(g_ClassUnlocked); class++)
    {
        g_ClassUnlocked[class] = false;
    }
    for (int slot = 0; slot < Slot_Count; slot++)
    {
        g_SlotUnlocked[slot] = false;
    }
    g_UnlockedMissions.Clear();
}

TFClassType Unlocks_ClassFromKey(const char[] key)
{
    for (int class = 1; class < sizeof(g_ClassKeys); class++)
    {
        if (StrEqual(g_ClassKeys[class], key, false))
        {
            return view_as<TFClassType>(class);
        }
    }
    return TFClass_Unknown;
}

int Unlocks_SlotFromKey(const char[] key)
{
    for (int slot = 0; slot < Slot_Count; slot++)
    {
        if (StrEqual(g_SlotKeys[slot], key, false))
        {
            return slot;
        }
    }
    return -1;
}

void Unlocks_GrantClass(const char[] key)
{
    TFClassType class = Unlocks_ClassFromKey(key);
    if (class == TFClass_Unknown)
    {
        LogError("The bridge granted the class %s. This plugin does not know it.", key);
        return;
    }
    g_ClassUnlocked[view_as<int>(class)] = true;
}

void Unlocks_GrantSlot(const char[] key)
{
    int slot = Unlocks_SlotFromKey(key);
    if (slot == -1)
    {
        LogError("The bridge granted the weapon slot %s. This plugin does not know it.", key);
        return;
    }
    g_SlotUnlocked[slot] = true;
}

void Unlocks_GrantMission(const char[] popFile)
{
    g_UnlockedMissions.SetValue(popFile, true);
}

bool Unlocks_HasMission(const char[] popFile)
{
    bool held;
    return g_UnlockedMissions.GetValue(popFile, held);
}

// Classes and slots are enforced apart. A set with neither cannot be right,
// the generator guarantees one of each, so an empty one is read as "the bridge
// has not told us yet" rather than "take everything away". Holding them apart
// is what stops a run with classes but no slot yet from enforcing nothing,
// which is what a play-test saw: every class pickable for the first four waves.
bool Unlocks_ClassesEnforced()
{
    if (!g_HaveUnlocks)
    {
        return false;
    }
    for (int class = 1; class < sizeof(g_ClassUnlocked); class++)
    {
        if (g_ClassUnlocked[class])
        {
            return true;
        }
    }
    return false;
}

bool Unlocks_SlotsEnforced()
{
    if (!g_HaveUnlocks)
    {
        return false;
    }
    for (int slot = 0; slot < Slot_Count; slot++)
    {
        if (g_SlotUnlocked[slot])
        {
            return true;
        }
    }
    return false;
}

// Runs as soon as the bridge answers: a grant that waits for the next spawn looks like no grant.
void Unlocks_ApplyToAll()
{
    if (!g_HaveUnlocks || !MvM_IsActive())
    {
        return;
    }
    for (int client = 1; client <= MaxClients; client++)
    {
        if (!MvM_IsPlayer(client))
        {
            continue;
        }
        Unlocks_EnforceClass(client);
        Unlocks_EnforceSlots(client);
    }
}

// Between waves the player is respawned on the spot, which is what the game
// itself does for a class change there. Mid-wave it waits for the next spawn:
// a forced respawn then hands back the life the death cost.
void Unlocks_EnforceClass(int client)
{
    if (!Unlocks_ClassesEnforced() || !MvM_IsPlayer(client))
    {
        return;
    }
    TFClassType current = TF2_GetPlayerClass(client);
    if (current == TFClass_Unknown || g_ClassUnlocked[view_as<int>(current)])
    {
        return;
    }
    TFClassType allowed = Unlocks_FirstAllowedClass();
    if (allowed == TFClass_Unknown)
    {
        // Nothing to move them to; leave them alone rather than as nothing.
        return;
    }
    TF2_SetPlayerClass(client, allowed, false, true);
    if (GameRules_GetRoundState() == RoundState_BetweenRounds && IsPlayerAlive(client))
    {
        // Not from inside the spawn event that brought us here.
        RequestFrame(Frame_Respawn, GetClientUserId(client));
        AP_PrintToClient(client, "The run did not unlock %s. You play %s.",
            g_ClassKeys[view_as<int>(current)], g_ClassKeys[view_as<int>(allowed)]);
        return;
    }
    AP_PrintToClient(client, "The run did not unlock %s. You spawn as %s next time.",
        g_ClassKeys[view_as<int>(current)], g_ClassKeys[view_as<int>(allowed)]);
}

static void Frame_Respawn(any userid)
{
    int client = GetClientOfUserId(userid);
    if (MvM_IsPlayer(client) && GameRules_GetRoundState() == RoundState_BetweenRounds)
    {
        TF2_RespawnPlayer(client);
    }
}

TFClassType Unlocks_FirstAllowedClass()
{
    for (int class = 1; class < sizeof(g_ClassUnlocked); class++)
    {
        if (g_ClassUnlocked[class])
        {
            return view_as<TFClassType>(class);
        }
    }
    return TFClass_Unknown;
}

// Every inventory application: spawn, resupply locker, and the upgrade station's loadout changes.
void Unlocks_EnforceSlots(int client)
{
    if (!Unlocks_SlotsEnforced() || !MvM_IsPlayer(client) || !IsPlayerAlive(client))
    {
        return;
    }
    bool removedActive = false;
    int active = GetEntPropEnt(client, Prop_Send, "m_hActiveWeapon");

    for (int slot = 0; slot < Slot_Count; slot++)
    {
        if (g_SlotUnlocked[slot])
        {
            continue;
        }
        if (active != -1 && GetPlayerWeaponSlot(client, slot) == active)
        {
            removedActive = true;
        }
        TF2_RemoveWeaponSlot(client, slot);
    }
    if (removedActive)
    {
        SwitchToAnyWeapon(client);
    }
}

// Removing the held weapon leaves no viewmodel and no attack, and the player cannot switch back.
static void SwitchToAnyWeapon(int client)
{
    for (int slot = 0; slot < Slot_Count; slot++)
    {
        if (!g_SlotUnlocked[slot])
        {
            continue;
        }
        int weapon = GetPlayerWeaponSlot(client, slot);
        if (weapon == -1)
        {
            continue;
        }
        char classname[64];
        if (GetEntityClassname(weapon, classname, sizeof(classname)))
        {
            FakeClientCommand(client, "use %s", classname);
            return;
        }
    }
    AP_PrintToClient(client, "The run locked every weapon slot. You hold nothing yet.");
}
