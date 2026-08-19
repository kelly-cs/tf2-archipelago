# Install on Windows

The easiest way to run a Mann vs Archipelago server. One file. No Docker, no
clone, no compiler.

Download `tf2ap.exe` from the
[latest release](https://github.com/m-this/tf2-archipelago/releases/latest)
and run it.

## What happens

A window opens and asks for your Archipelago room address. Then it installs
everything: SteamCMD, the TF2 dedicated server, SourceMod, the plugin, and
the bots that fill your team. The game server is about 14 GB, and the first
start takes a while because of it. Every later start takes seconds.

The window has:

- **Start**, **Stop**, **Restart**. A light beside them shows red, amber or
  green.
- **Join**, under the buttons: the addresses your friends connect to.
  **Copy** puts one on the clipboard.
- A **Log** tab and an **rcon** box, for when something looks wrong.
- A **Session** tab: connection status, checks, items, and the missions of
  the run. **Play this mission** loads the one you pick.
- **Settings**, for the room, the missions, the bots and the shape of the
  run.

Closing the window stops the server. Your answers are saved for next time.

## What you need

| Thing | What you need |
| --- | --- |
| Windows | 10 or 11, 64-bit |
| Disk | About 20 GB free |
| Memory | 4 GB for six players |
| Processor | Two cores |
| Network | Nothing, for friends on the same network or over Steam. One forwarded port only if you pick that route. |

No Docker, no Steam client, no Steam account for the server itself.

## The Archipelago session

The launcher runs the TF2 server. The multiworld session is separate. Mann
vs Machine is not one of the games that ship with Archipelago, so the seed
generator stays with the official app.

1. Install the official
   [Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases) app.
   The launcher finds it on its own in the usual places.
2. In the launcher, open **Settings**, set the player options, and press
   **Generate seed**. It writes the player file and opens the folder with
   the generated archive.
3. Upload that archive at
   [archipelago.gg/uploads](https://archipelago.gg/uploads) to open a room,
   and paste the room address (like `archipelago.gg:12345`) into the
   launcher.

If the launcher can't find the Archipelago app, **Generate seed** says so.
Point it at the app's folder in **Settings → Player options → Archipelago
app**.

See [Create the session](create-the-session.md) for the full detail,
including hosting the session yourself.

## Inviting friends

Your friends connect from the developer console:

```
connect your.server.address:27015
```

The **Join** line under the buttons shows the addresses to give out. See
[Invite your friends](invite-your-friends.md) for reaching people outside
your network.

## The bots on your team

Valve balances every wave for six players. The server fills empty RED seats
with bots that play: they pick classes, fight, and buy their own upgrades.
The **Bots** tab turns them off, shrinks the team for a harder run, or shapes
which classes they play. See
[The bots on your team](../play/defender-bots.md).

## Try it without Archipelago

**Test mode**, in Settings, runs a multiworld of one on your own machine —
no room, no seed, nothing leaves your computer. Use it to try the server
out or to check something before a real run.

## When you need help

**Debug logs**, in Settings, bundles everything useful into one file: the
launcher log, the server console, your settings, no passwords. Send it to
whoever is helping you.

**Repair** reinstalls the mods without touching the game files or your run.

See [Troubleshooting](../operate/troubleshooting.md) for the rest, and
[Install with Docker](install.md) if you'd rather run this on Linux.

---

## Reference

### Commands

Run these from a terminal. The window opens on its own otherwise.

| Command | What it does |
| --- | --- |
| `tf2ap.exe` | Open the window |
| `tf2ap.exe -room <host:port>` | Set the room address, then open the window |
| `tf2ap.exe -console` | Run in the terminal, with no window |
| `tf2ap.exe -configure` | Edit every setting in the terminal, then exit |
| `tf2ap.exe -install` | Install or repair the server, then exit |
| `tf2ap.exe -status` | Show the settings and the install state |
| `tf2ap.exe -yaml <path>` | Write the Archipelago player file, then exit |
| `tf2ap.exe -env` | List the environment variables, then exit |
| `tf2ap.exe -version` | Print the version and the pinned tool versions |

### Settings from the environment

Every setting also reads an environment variable, named the way
`deploy/.env.example` names it. A variable wins over the saved file for that
run:

```bat
set AP_ROOM=archipelago.gg:12345
set SRCDS_BOT_TEAM_SIZE=4
tf2ap.exe
```

`tf2ap.exe -env` prints every name it reads and marks the ones already set.

### Where it keeps things

| Path | Holds |
| --- | --- |
| `%USERPROFILE%\tf2-archipelago\` | The game files, SourceMod and SteamCMD |
| `%USERPROFILE%\tf2-archipelago\tf2.yaml` | The player file |
| `%USERPROFILE%\tf2-archipelago\bridge-state\` | The checks and unlocks of the run |
| `%APPDATA%\tf2ap\config.json` | Your saved settings |
| `%LOCALAPPDATA%\Programs\Archipelago\` | The Archipelago app, if installed there |

`TF2AP_INSTALL_ROOT` moves the first three, for a second disk.
