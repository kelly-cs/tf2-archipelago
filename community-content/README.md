# Community maps, missions, and upgrades

## The downloaded Potato archives and `tf2ap.exe`

This build directly recognizes these official full Potato packs:

```text
C:\Users\Admin\tf2\archive-assets.zip
C:\Users\Admin\tf2\mlarchive-assets.zip
```

In `tf2ap.exe`, open **Settings → Missions**, set **Asset pack folder** to
`C:\Users\Admin\tf2`, and tick **Potato Archive** and/or **Moonlight Archive**.
Press **Download Selected Community Assets** to fetch only the checked packs.
This button is the only operation that downloads community content. **Start**
never downloads a community file: it validates and installs selected local
ZIPs, and reports a missing or invalid ZIP instead. Existing valid downloads
are reused. A cancelled download leaves only a disposable `.partial` file.

Community mission and start-mission rows are not populated merely because a
pack checkbox is selected. They appear only after the matching ZIP exists and
validates. Maps without a bot `.nav` then appear in red as unavailable and
cannot be added to a seed.

### Use asset ZIPs you already have

Keep the original full-with-maps ZIP intact; do not extract or rename it. Put
either or both exact filenames directly in one folder:

```text
C:\Users\Admin\tf2\archive-assets.zip
C:\Users\Admin\tf2\mlarchive-assets.zip
```

Then press **Use Local Community Assets**, choose that folder, and let the
launcher validate the files. It checks the matching pack boxes and refreshes
the mission list without contacting the network. On the terminal interface,
type the folder into **Asset pack folder** first and activate the same action.

The ZIP may have Potato's normal `tf/download/...` root or a direct `tf/...`
root. The launcher reads it in place and installs its contents beneath the
dedicated server's `tf/` directory; it does not move, rewrite, or delete the
source archive.

The recognized downloads are:

- `https://dlarchive.potato.tf/archive-assets.zip`
- `https://dlml.potato.tf/mlarchive-assets.zip`

The `-no-maps.zip` alternatives are deliberately not used: they omit the
BSP/NAV files required by this catalog.

The mission table has an explicit **Source** column (`Valve` or
`Potato Archive`), and every start-mission choice has the same source prefix.
This build installs every asset in the selected archives and offers one
conservative, stock-syntax mission on each of 19 community maps:

| Pack | Portable maps |
| --- | --- |
| Potato Archive (15) | Condemned, Downpour, Frostwynd, Heatrock, Hideout, Kelly, Lotus, Null, Oilrig, Oxidize RC3, Radar, Redstone Ridge, Snowpine, Teien, Transmission |
| Moonlight Archive (4) | Area 52, Autumnull, Oxidize RR18, Yiresa |

This was checked against the actual ZIP central directories and population
files. The selection deliberately rejects missions containing obvious
RafMod-only population directives. Bogland and Cyberia are shown in red but
locked out of the pool because
the current full archives contain their BSPs but no NAV files; that avoids maps
where the included defender bots cannot route. `Anomalous Materials` was loaded with
the current stock Linux SRCDS parser. The same parser and content format are
used by Windows. The launcher strips the archive's
`tf/download/` prefix so those files arrive under `tf/scripts/items/`, and the
Archipelago plugin selects the registered table on every mission switch.

Mobocracy, Sudden Equinox, Trespasser and Trespasser Remaster keep their
Archipelago IDs as tombstones but are not offered by either launcher and can
never be drawn by a newly generated seed. They require RafMod/SigMod features
such as `PointTemplates`, `SpawnTemplate`, `ExtendedUpgrades` or
`CustomWeapon`. The public native extension is currently incompatible with
the current TF2 server and has no Windows build. The 19-map portable set uses
Valve's normal upgrade station table; the four loose custom tables remain
installed but are not assigned to an unrelated mission.

## Build a new Windows launcher

Run the build from WSL/Linux at the repository root. Building the standalone
Windows launcher does not require Docker:

```sh
cd /mnt/b/tf2-archipelago
make community-check COMMUNITY_CONTENT="/mnt/c/Users/Admin/tf2/archive-assets.zip /mnt/c/Users/Admin/tf2/mlarchive-assets.zip"
make export
make plugin
make launcher
```

The usable launcher is `dist/tf2ap.exe`. Do not substitute a plain
`go build`: the release target embeds the compiled SourceMod plugin, apworld,
ripext, and defender bots and injects their pinned versions.

Build the native WSL/Linux launcher from the same catalog with:

```sh
make launcher-linux
```

That produces `dist/tf2ap-linux-amd64`. Both binaries recognize the same ZIP
names and offer the same 19 portable community maps.

## Start a server with a Potato map and custom upgrades

1. Run `dist\tf2ap.exe`.
2. Open **Settings → Missions**.
3. Browse to `C:\Users\Admin\tf2` and tick both archives for the full set.
   Press **Download Selected Community Assets**, or press **Use Local
   Community Assets** if the ZIPs are already in that folder.
4. Choose `[Moonlight Archive] mvm_area_52_rc3 - Anomalous Materials` as the
   start mission. It is the stock-parser smoke-test mission for the portable
   set.
5. Tick the community missions wanted in the pool and untick unwanted Valve
   missions. Save.
6. For a wiring test, enable **Test mode**; otherwise press **Generate seed**,
   upload the generated archive to Archipelago, and enter the new room.
7. Press **Start**. The first run installs TF2 and extracts the selected local
   asset pack; it performs no community download. Watch for the plugin log
   line naming `scripts/items/mvm_upgrades.txt`.
8. In the server console or RCON, run `sm_ap_status`. In game, `!mission`
   lists the run and `!mission <number>` switches missions once their tickets
   are unlocked.

Generate a new seed after changing the registered mission pool; an existing
room cannot acquire the new stable mission IDs.

If this WSL installation still has the experimental 2025 SigMod/RafMod build
from an earlier Mobocracy attempt, stop the launcher and disable its autoload
file before using the portable catalog:

```sh
mv ~/tf2-archipelago/tf-dedicated/tf/addons/sourcemod/extensions/sigsegv.autoload \
   ~/tf2-archipelago/tf-dedicated/tf/addons/sourcemod/extensions/sigsegv.autoload.disabled
```

This is reversible: rename it back if a future current-compatible RafMod build
becomes available. A clean Windows install never receives that extension.

This directory is a content-pack overlay. Put the pack's normal `tf/` tree
under `community-content/tf/`; do not flatten it:

```text
community-content/
└── tf/
    ├── maps/mvm_underground_rc3.bsp
    ├── scripts/population/mvm_underground_rc3_welcometomymine.pop
    └── scripts/items/mvm_upgrades_tf2ap_example.txt
```

The server copies this tree into TF2's game directory on startup. Joining
clients receive the active map through Source's direct downloader; the managed
server configuration raises its stock 16 MB limit to the 64 MB engine cap.
Custom upgrade tables are added to Source's
download table and selected automatically whenever Archipelago changes to the
mission that names them. Population files and `.nav` files are server-side.

Prefer a map whose client assets are packed into its BSP. If a content pack
ships loose materials, models, particles, or sounds, it also needs its own
download manifest/SourceMod downloader; merely putting loose client assets in
this directory does not make TF2 send them.

## Register the content

Edit `gamedata/community.json`. IDs are permanent Archipelago identities:
start at 100, never reuse an ID, and keep the manifest with every server and
apworld that can load the resulting seed.

```json
{
  "format_version": 1,
  "maps": [
    {"id": 101, "name": "mvm_underground_rc3"}
  ],
  "missions": [
    {
      "id": 101,
      "pop_file": "mvm_underground_rc3_welcometomymine",
      "name": "WelcomeToMyMine",
      "map_id": 101,
      "difficulty": "intermediate",
      "waves": 12,
      "has_tank": false,
      "has_giant": true,
      "upgrades_file": "scripts/items/mvm_upgrades_tf2ap_example.txt"
    }
  ]
}
```

`upgrades_file` is optional. Omit it to use Valve's
`scripts/items/mvm_upgrades.txt`. It must be a game-relative path named
`scripts/items/mvm_upgrades_*.txt`. The plugin returns to Valve's table when
the next mission does not specify one.

The example uses a real Potato community mission: the TF2 Wiki identifies
[WelcomeToMyMine](https://wiki.teamfortress.com/wiki/WelcomeToMyMine_%28custom_mission%29)
as a 12-wave intermediate mission on Underground and gives its exact popfile
name. Replace every value with the metadata from the pack you actually install;
`examples/community.json` is documentation, not live content.

Validate and regenerate the apworld after every manifest edit:

```sh
make community-check
make export
```

`community-check` fails early if a registered BSP, population file, or upgrade
table is absent. `make export` bakes the new stable IDs into the apworld. All
registered community missions then participate in the existing difficulty,
mission-count, start-mission, and exclusion options alongside Valve missions.
Use `MVM_EXCLUDED_MISSIONS` if a particular custom mission should not be drawn.

## Rebuild and relaunch

Community content changes require an SRCDS restart. Manifest changes also
change the apworld and therefore require a newly generated seed:

```sh
make down
make community-check
make export
make seed
make build
make up
make logs
```

Upload the new file from `seed/` to Archipelago and put that room's host/port
in `.env` before `make up`. Do not reuse an old room after changing the
manifest: its seed does not know the new mission IDs. In `TF2AP_TEST_MODE=1`,
there is no seed or room, so `make seed` can be skipped.

For a content-file-only update that leaves `gamedata/community.json` unchanged,
restart just the game service so the overlay is recopied:

```sh
docker compose --project-directory . \
  --env-file deploy/env/versions.env --env-file .env \
  -f deploy/compose.yml restart srcds
```

## What Potato-style content works

- Vanilla `.bsp`, `.nav`, `.pop`, packed VScript, and standard custom MvM
  upgrade tables work with the layout above.
- TF2 exposes `SetCustomUpgradesFile` on `tf_gamerules`; that is the mechanism
  this plugin uses. The [SourceMod discussion and example](https://forums.alliedmods.net/showthread.php?t=256464)
  also documents pairing that input with `AddFileToDownloadsTable` for loose
  upgrade files.
- Potato's uploader accepts maps, population files, navigation files,
  `mvm_upgrades_*.txt`, VScript, and normal TF2 assets, which is a useful
  reference for the shape of real packs: [Potato content uploader](https://testing.potato.tf/upload.html).
- Missions that depend on Potato's server-only SigMod extensions are not
  automatically compatible. This repository does not install SigMod. Check
  the mission against the [Potato SigMod wiki](https://sigwiki.potato.tf/index.php/Main_Page)
  and either install/pin the required extension separately or choose a
  vanilla/VScript mission.

Custom weapons are the next layer rather than part of this first integration.
Packed map VScript weapons can work as content today, but making individual
custom weapons into Archipelago unlock items needs a stable weapon manifest,
grant/enforcement support in SourceMod, and dependency declarations for any
SigMod attributes. Keeping that separate prevents a map-and-upgrade pack from
silently requiring an uninstalled server mod.
