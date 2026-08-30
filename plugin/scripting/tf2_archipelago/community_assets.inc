/**
 * Loose client files for supported community missions.
 *
 * A BSP makes Source request itself and anything packed inside it. ClassIcon
 * values live in population files, which are server-side, so their loose VMT
 * and VTF files never enter the client download table on their own.
 */

#include "community_assets_data.inc"

void CommunityAssets_OnMapStart()
{
    char map[PLATFORM_MAX_PATH];
    GetCurrentMap(map, sizeof(map));

    int registered = 0;
    for (int index = 0; index < CommunityClientAssetCount; index++)
    {
        if (!StrEqual(map, g_CommunityClientAssetMaps[index], false))
        {
            continue;
        }
        if (!FileExists(g_CommunityClientAssetPaths[index], true))
        {
            AP_Error("Community client asset is missing: %s", g_CommunityClientAssetPaths[index]);
            continue;
        }
        AddFileToDownloadsTable(g_CommunityClientAssetPaths[index]);
        registered++;
    }
    if (registered > 0)
    {
        AP_Debug("Registered %d loose community client assets for %s.", registered, map);
    }
}
