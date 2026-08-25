/** Weapon-specific useful items received from the multiworld. */

StringMap g_UnlockedWeaponBuffs;

bool g_WeaponBuffApplied[2049];
int g_WeaponBuffAppliedRef[2049];
int g_WeaponBuffAppliedIndex[2049];
float g_WeaponBuffAppliedValue[2049];

void WeaponBuffs_Init()
{
    g_UnlockedWeaponBuffs = new StringMap();
}

void WeaponBuffs_Clear()
{
    g_UnlockedWeaponBuffs.Clear();
}

bool WeaponBuffs_Grant(const char[] key)
{
    for (int index = 0; index < sizeof(g_WeaponBuffKeys); index++)
    {
        if (StrEqual(g_WeaponBuffKeys[index], key))
        {
            g_UnlockedWeaponBuffs.SetValue(key, true);
            return true;
        }
    }
    LogError("The bridge granted weapon buff %s. This plugin does not know it.", key);
    return false;
}

static bool WeaponBuffs_IsUnlocked(int index)
{
    bool unlocked;
    return g_UnlockedWeaponBuffs.GetValue(g_WeaponBuffKeys[index], unlocked);
}

static int WeaponBuffs_ForDefinition(int definition)
{
    for (int index = 0; index < sizeof(g_WeaponBuffByDefinition); index++)
    {
        if (g_WeaponBuffByDefinition[index][0] == definition)
        {
            return g_WeaponBuffByDefinition[index][1];
        }
    }
    return -1;
}

static int WeaponBuffs_ForEntity(int weapon)
{
    if (weapon <= MaxClients || !IsValidEntity(weapon) || !HasEntProp(weapon, Prop_Send, "m_iItemDefinitionIndex"))
    {
        return -1;
    }
    int definition = GetEntProp(weapon, Prop_Send, "m_iItemDefinitionIndex");
    int buff = WeaponBuffs_ForDefinition(definition);
    return buff >= 0 && WeaponBuffs_IsUnlocked(buff) ? buff : -1;
}

// A station upgrade and an Archipelago buff can share an attribute. The
// station's value is read first and composed with ours. Remembering the result
// makes a second inventory event idempotent instead of multiplying it again.
static void WeaponBuffs_ApplyEntity(int weapon, int buff)
{
    int reference = EntIndexToEntRef(weapon);
    Address attribute = TF2Attrib_GetByName(weapon, g_WeaponBuffAttributes[buff]);
    float current = g_WeaponBuffAdditive[buff] ? 0.0 : 1.0;
    if (attribute != Address_Null)
    {
        current = TF2Attrib_GetValue(attribute);
    }
    if (g_WeaponBuffApplied[weapon]
        && g_WeaponBuffAppliedRef[weapon] == reference
        && g_WeaponBuffAppliedIndex[weapon] == buff
        && FloatAbs(current - g_WeaponBuffAppliedValue[weapon]) < 0.0001)
    {
        return;
    }

    float value = g_WeaponBuffAdditive[buff]
        ? current + g_WeaponBuffValues[buff]
        : current * g_WeaponBuffValues[buff];
    if (!TF2Attrib_SetByName(weapon, g_WeaponBuffAttributes[buff], value))
    {
        LogError("Could not apply %s to %s.", g_WeaponBuffAttributes[buff], g_WeaponBuffWeapons[buff]);
        return;
    }
    g_WeaponBuffApplied[weapon] = true;
    g_WeaponBuffAppliedRef[weapon] = reference;
    g_WeaponBuffAppliedIndex[weapon] = buff;
    g_WeaponBuffAppliedValue[weapon] = value;
}

void WeaponBuffs_Apply(int client)
{
    if (!g_HaveUnlocks || !MvM_IsPlayer(client) || !IsPlayerAlive(client))
    {
        return;
    }
    for (int slot = 0; slot <= 5; slot++)
    {
        int weapon = GetPlayerWeaponSlot(client, slot);
        int buff = WeaponBuffs_ForEntity(weapon);
        if (buff >= 0)
        {
            WeaponBuffs_ApplyEntity(weapon, buff);
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
        int buff = WeaponBuffs_ForEntity(GetPlayerWeaponSlot(client, slot));
        if (buff < 0)
        {
            continue;
        }
        char line[128];
        FormatEx(line, sizeof(line), "%s: %s", g_WeaponBuffWeapons[buff], g_WeaponBuffDescriptions[buff]);
        menu.AddItem(g_WeaponBuffKeys[buff], line);
        listed++;
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

public Action Command_TestWeaponBuff(int client, int args)
{
    if (!MvM_IsPlayer(client) || !IsPlayerAlive(client))
    {
        ReplyToCommand(client, "[AP] Join RED and spawn before testing a weapon buff.");
        return Plugin_Handled;
    }
    int weapon = GetEntPropEnt(client, Prop_Send, "m_hActiveWeapon");
    int definition = weapon > MaxClients && IsValidEntity(weapon)
        ? GetEntProp(weapon, Prop_Send, "m_iItemDefinitionIndex") : -1;
    int buff = WeaponBuffs_ForDefinition(definition);
    if (buff < 0)
    {
        ReplyToCommand(client, "[AP] The active item definition %d has no weapon buff.", definition);
        return Plugin_Handled;
    }
    g_UnlockedWeaponBuffs.SetValue(g_WeaponBuffKeys[buff], true);
    WeaponBuffs_ApplyEntity(weapon, buff);
    ReplyToCommand(client, "[AP] Test buff: %s — %s.",
        g_WeaponBuffWeapons[buff], g_WeaponBuffDescriptions[buff]);
    WeaponBuffs_ShowMenu(client);
    return Plugin_Handled;
}
