# Community maps and missions

## The downloaded Potato archives and `tf2ap.exe`

This build directly recognizes these official full Potato packs by filename:

```text
archive-assets.zip
mlarchive-assets.zip
```

In `tf2ap.exe`, open **Settings → Missions**, choose the folder that will hold
the ZIPs, and tick **Potato Archive** and/or **Moonlight Archive**. Press
**Start**. If a selected ZIP is absent, the launcher downloads it from Potato's
official archive host into that folder, validates the completed ZIP, and then
streams it into the dedicated server. Existing downloads are reused; later
starts skip unchanged archives. Cancelling cannot leave a partial file that the
next run mistakes for a complete pack.

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

The catalog enables missions that use stock server features and have the
required map, population, and navigation files. Entries missing bot navigation
remain visible as unavailable and cannot be drawn by a seed. The launcher
strips the archive's `tf/download/` prefix so assets arrive under the correct
TF2 game directory on Windows and Linux.

## Build a new Windows launcher

Run the build from WSL/Linux at the repository root. Building the standalone
Windows launcher does not require Docker:

```sh
make community-check COMMUNITY_CONTENT="archive-assets.zip mlarchive-assets.zip"
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

## Start a server with a Potato map

1. Run `dist\tf2ap.exe`.
2. Open **Settings → Missions**.
3. Choose the folder for the supported ZIPs and tick both archives for the full
   set. The launcher downloads either full pack if it is not already there.
4. Choose any available community mission as the start mission.
5. Tick the community missions wanted in the pool and untick unwanted Valve
   missions. Save.
6. For local play, enable **Test mode**; otherwise press **Generate seed**,
   upload the generated archive to Archipelago, and enter the new room.
7. Press **Start**. The first run installs TF2 and extracts the selected asset
   pack.
8. In the server console or RCON, run `sm_ap_status`. In game, `!mission`
   lists the run and `!mission <number>` switches missions once their tickets
   are unlocked.

Generate a new seed after changing the registered mission pool; an existing
room cannot acquire the new stable mission IDs.

This directory is a content-pack overlay. Put the pack's normal `tf/` tree
under `community-content/tf/`; do not flatten it:

```text
community-content/
└── tf/
    ├── maps/mvm_underground_rc3.bsp
    ├── maps/mvm_underground_rc3.nav
    └── scripts/population/mvm_underground_rc3_welcometomymine.pop
```

The server copies this tree into TF2's game directory on startup. Maps are
downloaded by joining clients.
Population files and `.nav` files are server-side.

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
      "has_giant": true
    }
  ]
}
```

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

`community-check` fails early if a registered BSP, navigation mesh, or
population file is absent. `make export` bakes the new stable IDs into the apworld. All
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

- Vanilla `.bsp`, `.nav`, `.pop`, packed VScript, and normal TF2 assets work
  with the layout above. Selectable missions use Valve's upgrade table.
- Missions that require additional server extensions are not included in the
  supported catalog.
