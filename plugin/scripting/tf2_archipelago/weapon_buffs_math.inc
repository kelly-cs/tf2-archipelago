/**
 * The decisions in weapon_buffs.inc that touch nothing but their arguments.
 *
 * Split out so they can be compiled on their own. weapon_buffs.inc opens with
 * sourcemod, sdkhooks, tf2 and ripext, so nothing in it could be run without a
 * game server, and it was checked instead by reading its source for substrings.
 * That catches a rename and misses a changed sum, which is the wrong way round
 * for both: inverting the cooldown floor below so every cooldown collapses to
 * the minimum leaves every watched string in place.
 *
 * This file includes nothing. gamedata/spdriver_test.go compiles it against
 * weapon_buffs_data.inc and runs it under SourcePawn's standalone VM, so the
 * answers here are measured rather than read.
 *
 * A function belongs here when it needs nothing but its arguments and the
 * generated tables: no entity property, no SDKCall, no TF2 lookup, not even
 * RoundToFloor or StrEqual, which the standalone VM does not bind, and no
 * constant out of a SourceMod include either. WeaponBuffs_IsDamageOverTime is
 * pure arithmetic and still does not qualify, because TF_CUSTOM_BLEEDING comes
 * from tf2_stocks; it stays next door and is checked by reading, which is the
 * best that can be done for it.
 *
 * Nothing here is static. SourcePawn scopes a static function to its own file,
 * so these would be invisible to the file they were cut out of. The
 * WeaponBuffs_ prefix is what keeps them apart in the one flat namespace.
 */

#define MoveSpeedEffect 24
#define JumpHeightEffect 25
#define ActiveHealthRegenEffect 66
#define HealOnKillEffect 6
#define CritsOnKillEffect 7
#define SpeedOnKillEffect 22
#define MinicritsOnKillEffect 70
#define BaseHealthOnKillEffect 80
#define ProjectileDestructionBaseCooldown 2.0
#define ProjectileDestructionCooldownStep 0.25
#define ProjectileDestructionMinimumCooldown 0.5

bool WeaponBuffs_IsPassiveEffect(int effect)
{
    return effect == MoveSpeedEffect || effect == JumpHeightEffect
        || effect == ActiveHealthRegenEffect;
}

bool WeaponBuffs_IsOnKillEffect(int effect)
{
    return effect == HealOnKillEffect || effect == CritsOnKillEffect
        || effect == SpeedOnKillEffect || effect == MinicritsOnKillEffect
        || effect == BaseHealthOnKillEffect;
}

float WeaponBuffs_ProjectileDestructionCooldown(int levels)
{
    float cooldown = ProjectileDestructionBaseCooldown
        - float(levels - 1) * ProjectileDestructionCooldownStep;
    return cooldown > ProjectileDestructionMinimumCooldown
        ? cooldown : ProjectileDestructionMinimumCooldown;
}

int WeaponBuffs_ForDefinition(int definition)
{
    for (int index = 0; index < sizeof(g_WeaponByDefinition); index++)
    {
        if (g_WeaponByDefinition[index][0] == definition)
        {
            return g_WeaponByDefinition[index][1];
        }
    }
    return -1;
}
