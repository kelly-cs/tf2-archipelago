# launcher

Go. The all-in-one Windows and Linux launcher. It embeds the compiled plugin,
ripext and the MvM defender bots. It installs SteamCMD, the TF2 dedicated
server, Metamod:Source and SourceMod, then runs the bridge in-process next to
the SRCDS subprocess.

One executable, no Docker and no clone. Its browser interface is identical on
both platforms.

## Build

```sh
make launcher
```

Cross-compiles `tf2ap.exe` into `dist/`. The target stages the bots, fetches
the ripext Windows zip into the embed dir, copies the compiled plugin, and
injects the pinned versions from `deploy/env/versions.env` with `-ldflags`.

The Missions tab can download missing official Potato/Moonlight full asset
ZIPs into a player-selected cache folder and install them into SRCDS. It labels Valve and Potato missions
separately in both the pool table and start-mission menu. See the
[community-content guide](../community-content/README.md) for the verified
archive layout, custom-upgrade findings, build commands, and RafMod boundary.

## Layout

| Package | Holds |
| --- | --- |
| `cmd/tf2ap` | Entrypoint: flags, the guided flow, subcommands |
| `internal/assets` | The embedded files and the injected version strings |
| `internal/settings` | The saved configuration, the environment overlay, the player YAML |
| `internal/installer` | SteamCMD, TF2 server, Metamod, SourceMod, ripext, plugin, bots |
| `internal/srcdsconfig` | Renders `server.cfg`, `admins_simple.ini`, `tf2_archipelago.cfg` |
| `internal/runtime` | The `srcds.exe` subprocess and the in-process bridge, interleaved |
| `internal/webui` | The cross-platform browser UI: logs, controls, settings, session and RCON |
| `internal/generate` | Drives the Archipelago app's generator: installs the apworld, writes the player file, runs it |
| `internal/debugbundle` | The zip a play-tester sends: logs, settings without passwords, player file |
| `../fakeroom` | The multiworld of one that test mode serves, shared with the bridge |
| `internal/rcon` | Source RCON client, shared by the command box and `cmd/rcon`, which `make rcon` runs |
| `internal/runshape` | The run's choices, counted from `gamedata` |
| `internal/ui` | Console prompts, and the console the Windows build attaches to |

## Configuration

The no-args run asks one question, the room address, and only on a first run.
Everything else carries a default, and `settings.NewRconPassword` invents the
one secret rather than asking for it. `-configure` opens the full list.

Three layers, in order. The defaults, then `%APPDATA%\tf2ap\config.json`, then
the environment. `settings.ApplyEnv` reads the names `deploy/.env.example`
already uses, so a compose operator's file works here unchanged. An environment
value is never written back: an override for one run must not become the saved
answer.

## The browser and the console

With no arguments the launcher serves its embedded interface on a random
loopback port and opens it in the desktop browser. Windows and Linux draw the
same HTML over the same HTTP handlers. `runtime.Supervisor` owns the processes
behind Start and Stop, and Server-Sent Events carry its log and state changes
to the page. The server listens only on `127.0.0.1`.

`-console` keeps the prompt-and-log flow for a headless machine. The Windows
exe still links with `-H windowsgui`; `ui.AttachConsole` gives flags their
output back when a terminal started them.

## How it fits

The launcher imports `bridge/` and runs it in-process. The `srcds.exe`
subprocess shares the machine's loopback with the bridge, so the plugin reaches
it at `127.0.0.1:24680` exactly as it does under `network_mode: service:srcds`
in compose.

Seed generation stays with the Archipelago app: the generator is Python, and a
bundled Python breaks the one-exe promise. What the launcher does is drive that
app. Generate seed finds it where the installer put it and installs the
embedded apworld into `custom_worlds`. It writes the player file into a folder
of its own, runs `ArchipelagoGenerate.exe`, and opens the folder the archive
landed in. The player uploads that archive to archipelago.gg.

## The embeds

`internal/assets/embedded/` holds:

- `tf2_archipelago.smx` (gitignored, copied by `make launcher-assets`)
- `sm-ripext-windows.zip` (gitignored, fetched by `make launcher-assets`)
- `defender-bots-windows.zip` (gitignored, built by `make bots`, `.so` stripped)
- `tf2_mvm.apworld` (gitignored, built without Docker by `make apworld-package`)
- `tf2_archipelago.cfg` (committed)
- `server.cfg.tmpl` (committed)

The first three are build artefacts, and a hand `go build` needs placeholders
in their place. Such a build also leaves the version strings empty, so
`assets.RequireVersions()` stops the installer rather than let it guess a
version.
