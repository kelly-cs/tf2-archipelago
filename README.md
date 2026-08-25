# Mann vs Archipelago

An [Archipelago](https://archipelago.gg) randomizer for Team Fortress 2, in
Mann vs Machine mode. The classes, the weapon slots and the missions start
locked. The team clears waves to unlock them. Everybody on the server shares
the same unlocks.

<p align="center">
  <a href="https://github.com/m-this/tf2-archipelago/releases/latest/download/tf2ap.exe">
    <img alt="Download tf2ap.exe for Windows" src="https://img.shields.io/badge/Download-tf2ap.exe%20for%20Windows-2ea44f?style=for-the-badge&logo=windows&logoColor=white">
  </a>
  <a href="https://github.com/m-this/tf2-archipelago/releases/latest/download/tf2ap-linux-amd64">
    <img alt="Download tf2ap-linux-amd64 for Linux" src="https://img.shields.io/badge/Download-tf2ap--linux--amd64-1b1f27?style=for-the-badge&logo=linux&logoColor=white">
  </a>
</p>

> [!NOTE]
> **Windows will warn you about `tf2ap.exe`, and it is a false positive.** It
> unpacks archives and starts a game server without a code signature, which
> looks like a virus to a scanner.
> Check it rather than trust it: every release publishes `SHA256SUMS` and a
> VirusTotal report, and `make launcher` builds the same exe on your machine.

<p align="center">
  <img alt="The launcher: the run, Start and Stop, a Join button" src="docs/images/launcher-main.png" width="820">
</p>

## What talks to Archipelago

- **Location** (check): a wave cleared, a mission cleared, the first tank
  destroyed in a mission, the first giant killed in one.
- **Item** (unlock): a mercenary class, a weapon slot, a mission ticket. All
  shared: one unlocked class or slot opens for every player on the server.
- **Filler**: `Cash Bundle`, 200 credits paid to everyone on RED.
- **DeathLink**: off by default. A wave your team loses sends a death to the
  multiworld; a death that arrives kills your team, which loses the wave you
  are on.

See [What the randomizer changes](./docs/en/what-the-randomizer-changes.md)
for the detail, and [Archipelago for MvM players](./docs/en/archipelago-for-mvm-players.md)
for the vocabulary.

## Windows

Download `tf2ap.exe` and run it. One file: no Docker, no clone, no compiler.
SmartScreen will stop you the first time: click **More info**, then **Run
anyway**. The note at the top of this page says why it does that and how to
check the file first. It opens a window and asks for the address of your
Archipelago room. Then it installs the rest: SteamCMD, the game server,
SourceMod, the plugin, and the bots that fill your team.

The window holds the log, **Start** and **Stop**, an **rcon** box, and
**Settings**. Settings has six tabs: the run, the missions it may draw, the
room, the game server, the bots, and who can join. It also makes the seed for
you with the Archipelago app, writes the player file, bundles the logs to send
when something looks wrong, and puts every answer back to its default. A
**Test mode** plays without any room at all.

<p align="center">
  <img alt="Settings: the player options tab" src="docs/images/launcher-settings.png" width="740">
</p>

[The Windows guide](./docs/en/setup/install-windows.md) takes you from the
download to the first wave.

## Linux

Download `tf2ap-linux-amd64` and run it. Same program, same interface as the
window above in everything but how it is drawn: a full-screen terminal
interface, with the log, the run's missions, the rcon line and the same six
tabs of settings. It is Bubble Tea, which is pure Go, so it needs nothing
installed and works over SSH.

It is on Windows too. `tf2ap.exe -tui` opens it instead of the window, which is
what to use over SSH or in a terminal you already have open. `-console` on
either platform prints the log and nothing else, which is what a service wants.

```sh
curl -fsSLO https://github.com/m-this/tf2-archipelago/releases/latest/download/tf2ap-linux-amd64
chmod +x tf2ap-linux-amd64
./tf2ap-linux-amd64
```

Every answer either interface asks for is also a variable, and `-status` says
what the run is set to before it starts:

<p align="center">
  <img alt="tf2ap-linux-amd64 -status" src="docs/images/linux-status.svg" width="740">
</p>

[The Linux guide](./docs/en/setup/install-linux.md) has the rest.

## Docker

For any machine with Docker:

```sh
cp deploy/.env.example .env   # SRCDS_RCONPW, AP_HOST and AP_PORT have no default
make seed                     # upload the file to archipelago.gg, open a room
make up
make logs
```

`make seed` generates the multiworld here, because archipelago.gg generates
only the games that ship with Archipelago. It hosts the file all the same, so
the stack runs no Archipelago server of its own.
[`COMPOSE_PROFILES=selfhost`](./docs/en/setup/create-the-session.md) brings one
back. `TF2AP_TEST_MODE=1` plays without a room at all.

Every release attaches a `compose.yaml` that pulls published images instead of
building them. See
[`docs/en/setup/install.md`](./docs/en/setup/install.md#without-the-repository).

The first start downloads about 14 GB of game files.

Custom MvM packs can be overlaid without modifying the container image. The
[community content guide](./community-content/README.md) covers custom BSPs,
population files, stable Archipelago IDs, validation, and the rebuild/relaunch
sequence. In the standalone launcher, community downloads happen only through
**Download Selected Community Assets**. Use **Use Local Community Assets** for
full pack ZIPs you already have; **Start** never downloads community content.

## The bots on your team

Valve balances every wave for six players. The server fills the RED team with
bots that play. They pick classes, fight, buy their own upgrades and ready
themselves, so two people can win a run. See
[The bots on your team](./docs/en/play/defender-bots.md).

## Why MvM

MvM is the only TF2 mode with built-in progress:

- A mission is an ordered series of waves.
- A team clears a wave, or fails it.
- A shop sells upgrades that persist.
- A round is an ordered series of missions.

This structure maps directly onto Archipelago's regions and locations.
Classic TF2 has none of it.

## How it works

Three processes. One source of truth.

```
  gamedata/ (Go)  ──generates──>  apworld/tf2_mvm/data/*.json
        │                                   │
        │ built into                        │ read at generation
        v                                   v
    bridge (Go)  <──websocket──>  Archipelago server (archipelago.gg)
        ^
        │ HTTP + JSON on 127.0.0.1
        v
  SourceMod plugin  (inside the game server)
```

The SourceMod plugin is the only part that sees the game. The Go bridge is
the only part that speaks Archipelago. `gamedata/`, in Go, is the only part
that knows what a mission or a weapon is. It exports that data as JSON for
the Python apworld. The Windows launcher runs the bridge in-process next to
the game server; the compose stack runs them as two containers.

Players connect with a stock TF2 client. They install nothing.

## Directory tree

| Directory | Language | Role |
| --- | --- | --- |
| [`gamedata/`](./gamedata/) | Go | Source of truth: maps, missions, waves, weapons, upgrades, robots, and the IDs. Exports the JSON. |
| [`bridge/`](./bridge/) | Go | Archipelago client. WebSocket, reconnection, a durable queue, and a loopback HTTP API for the plugin. |
| [`fakeroom/`](./fakeroom/) | Go | The multiworld of one that test mode serves, with simulated players. |
| [`apworld/`](./apworld/) | Python | A thin apworld. It reads the exported JSON and sets the regions, the rules, and the YAML options. |
| [`plugin/`](./plugin/) | SourcePawn | Detects the objectives. Applies the unlocks. |
| [`launcher/`](./launcher/) | Go | The Windows exe: a window over the bridge and the game server, the installer, the seed generator's driver. |
| [`deploy/`](./deploy/) | Compose, shell | The images, the compose files, and the build of the defender bots. |
| [`docs/`](./docs/) | Markdown | The book, in English and French. Spec, ADRs, prior art, and the original Discord thread. |

## Development

```sh
make check        # everything CI runs
make integration  # real Archipelago and bridge, driven the way the plugin drives them
make launcher     # cross-compile tf2ap.exe into dist/
make bots         # stage the defender bots the image and the exe carry
```

## Documentation

`docs/` is a book, in English and in French. GitHub Pages publishes it at
[m-this.github.io/tf2-archipelago](https://m-this.github.io/tf2-archipelago/),
and `make docs` builds and serves the English version on `127.0.0.1:8081`.

- [`docs/en/SUMMARY.md`](./docs/en/SUMMARY.md): the table of contents.
- [`docs/en/archipelago-for-mvm-players.md`](./docs/en/archipelago-for-mvm-players.md):
  for a player new to a multiworld. The vocabulary and the chat commands.
- [`docs/en/spec.md`](./docs/en/spec.md): the design. Scope, locations,
  items, and goals.
- [`docs/en/adr/`](./docs/en/adr/): the decisions, and why the alternatives
  lost.
- [`docs/en/discord-mvm-thread.md`](./docs/en/discord-mvm-thread.md): the
  original Archipelago Discord thread, copied word for word. **Damonj17**
  and **Roseburst** wrote the design.
- [`docs/en/prior-art.md`](./docs/en/prior-art.md): what exists already,
  most of all the fork
  [ALPHAMARIOX/TF2-MvM-Archipelago](https://github.com/ArchipelagoMW/Archipelago/compare/main...ALPHAMARIOX:TF2-MvM-Archipelago:main).
- [`CONTEXT.md`](./CONTEXT.md): the glossary. Archipelago's and MvM's
  vocabularies share words but not their meanings; this file fixes both.

## Code signing policy

Free code signing for open source projects, provided by
[SignPath.org](https://signpath.org), with a certificate from the
[SignPath Foundation](https://signpath.org).

> [!WARNING]
> The application is open, not granted. Nothing is signed yet: `tf2ap.exe` ships
> unsigned, and Windows warns about it until that changes.

## Licence

This repository is MIT. See [LICENSE](./LICENSE).

What it ships is not all MIT. The defender bots are GPL-3.0, and so is
[our fork](https://github.com/m-this/tf2-mvm-bots) of them. `tf2ap.exe` and
`tf2-defender-bots.zip` both carry their compiled plugins, and the fork is
where that source lives. Every other project in the bot stack keeps its own
terms. [Defender bots](./docs/en/play/defender-bots.md) names each one.

The launcher's icon is the Team Fortress crosshair, traced from the
[Team Fortress 2 wordmark](https://commons.wikimedia.org/wiki/File:Team-Fortress-2-logo.png)
on Wikimedia Commons, which is public domain as a work below the threshold of
originality. The mark is still Valve's trademark. This project is a fan project
and is not affiliated with or endorsed by Valve Corporation.

## Credits

The design comes from the Archipelago Discord thread.

- **Damonj17** set the premise and the shape of the items and the checks.
- **Roseburst** wrote the entire YAML options schema.
- **ALPHAMARIOX**'s fork is where the starting data tables come from.
- **adeleine64DS**, **Amazia**, **Snolid Ice**, **mudkipslike**,
  **TheBreadstick**, **CrystalClear** and **Pixel Silzavon** contributed.
- **SwagDoll420** and **EZKSupernova** play-tested the runs the issue list
  comes from. Most of what the bots do came from their reports: the class
  blacklist, the excluded missions, and the bots' purchases in the chat.

The bots are [OfficerSpy's MvM Defender TFBots](https://github.com/OfficerSpy/TF2-MvM-Defender-TFBots),
carried in [a fork](https://github.com/m-this/tf2-mvm-bots).
