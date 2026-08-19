# Install on Windows

This is the easiest way to run a Mann vs Archipelago server. One file, no
Docker, no clone, no compiler.

Download `tf2ap.exe` from the
[latest release](https://github.com/m-this/tf2-archipelago/releases/latest) and
run it.

## What tf2ap.exe does

Double-click it and a window opens. The window holds everything an evening
needs:

- **Start**, **Stop** and **Restart**, so the server goes up and down without a
  terminal. A light beside them is red when the server is stopped, amber
  while it starts, and green when it runs.
- The **Join** line, under the buttons: every address of this machine on the
  game port, which is what your friends type after `connect`. Over Steam it
  also shows the address Steam assigned. **Copy** puts the line on the
  clipboard.
- A **Log** tab, where the game server, the bridge and the installer write.
  This is what you read when something looks wrong. The **rcon** box under it
  sends a command to the server and prints the answer in the log.
  `sm_ap_status` is the one to know.
- A **Session** tab. It shows whether the bridge is connected to the room,
  and how many checks and items the run holds. It lists the missions of the
  run, unlocked, cleared or locked. **Play this mission** loads the selected
  one.
- **Settings**, for the room address, the missions, the bots and the shape of
  the run.

The first Start installs SteamCMD, the TF2 dedicated server, Metamod:Source,
SourceMod, the plugin and the defender bots. The game server is about 14 GB.
This is the long part, and it happens once. The log tracks it.

Closing the window stops the server. Every later run takes seconds, and the
launcher saves your answers in `%APPDATA%\tf2ap\config.json`.

## What you need

| Thing | What you need |
| --- | --- |
| Windows | 10 or 11, 64-bit. |
| Disk | About 20 GB free. The game server is about 14 GB. |
| Memory | 4 GB for six players. |
| Processor | Two cores. |
| Network | One port your friends can reach, UDP and TCP. The default is 27015. |

You do not need Docker. You do not need the Steam client. You do not need a
Steam account for the server.

## The Archipelago session

The launcher runs the TF2 server. The multiworld session is separate. Mann vs
Machine is not one of the games that ship with Archipelago. The generator is
Python, so it stays with the official app.

1. Install the official
   [Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases) app.
   The launcher finds it where the installer puts it.
2. In the launcher, open **Settings**, set the player options, and press
   **Generate seed**. The launcher puts the world file into the app and writes
   the player file. It then runs the generator and opens the folder with the
   archive.
3. Upload that archive at
   [archipelago.gg/uploads](https://archipelago.gg/uploads) to open a room.

The room page shows an address such as `archipelago.gg:12345`. Paste it into
the **Archipelago room** tab.

The player file is also at `%USERPROFILE%\tf2-archipelago\tf2.yaml`, and
**Open tf2.yaml** shows it. That is for anybody who wants to generate in the
app by hand.

See [Create the session](create-the-session.md) for the detail of each step.
It also shows how to host the session on your own machine.

## The first start

Double-click `tf2ap.exe`, or run it from a terminal:

```
tf2ap.exe
```

A terminal window opens. The first start goes like this:

1. The launcher asks for the room address. Paste the line from your room page,
   such as `archipelago.gg:12345`, and press Enter. This comes first, so a
   typo costs you a second rather than a download.
2. The launcher installs SteamCMD. Seconds.
3. The launcher installs the TF2 dedicated server. This is the 14 GB download.
   The time depends on your connection.
4. The launcher starts the server. The log says
   `connected to archipelago slot=tf2`, and the room page says
   `tf2 (Team #1) playing Team Fortress 2 Mann vs Machine has joined`.

Your friends connect from the developer console:

```
connect your.server.address:27015
```

The **Join** line under the buttons shows the addresses to give out. See
[Invite your friends](invite-your-friends.md) for the rest.

## Who can join

The **Game server** tab has a **Who can join** choice:

| Choice | What it means |
| --- | --- |
| This machine and the local network only | The default. No Steam login, no token, nobody outside your network. |
| Over Steam, no port to open | Steam relays the traffic. Valve assigns the server an address such as `169.254.13.42:20232`, and your friends connect to that from anywhere. Nothing to forward. The address is a new one at every start; the Join line shows it once the server has it. Needs a login token. |
| Over a port forwarded on the router | The classic setup: forward the game port to this machine and give out your public address. Needs a login token. |

A Game Server Login Token comes from
[steamcommunity.com/dev/managegameservers](https://steamcommunity.com/dev/managegameservers),
for app 440. Paste it in **Login token**.

## The missions

The **Missions** tab has two things:

- **Start mission**, as map - mission. The server loads it first. If the run
  has not unlocked it, the plugin moves to the first mission the run has.
- The missions the run can draw. Untick one to keep it out of every seed you
  generate from here. Caliginous Caper is one wave of 666 robots and an hour
  on its own. The tier on the Player options tab still applies.

Once the run is going, the server moves from a cleared mission to the next
unlocked one on its own. See [Chat commands](../play/chat-commands.md).

## The bots on your team

The server fills the RED team with bots that play. Team Fortress 2 balances
every wave for six defenders, so two people can win a run. The bots pick
classes, fight, buy their own upgrades and ready themselves.

The **Bots** tab turns them off, or fills RED to fewer than six for a harder
run. It unticks the classes the bots must not play, and picks a loadout preset
per class. It can also write what they buy in the chat.
[The bots on your team](../play/defender-bots.md) says what they do and what
they are bad at.

## Try it without Archipelago

Settings has a **Test mode** box. With it on, the launcher serves a multiworld
of one on your own machine. It makes up a seed from your player options and
hands out an unlock for every wave you clear.

It also plays the other players: they find things, send you things and die.
Every line of that reaches the log and the game's chat.

The run starts the way a generated one does: one class, one weapon slot and
one mission. The server locks what a real seed locks, from the first wave.
What the simulated players send you is filler. The unlocks come from the waves
you clear, and from nothing else.

Nothing leaves the machine, and you need no room and no seed. Use it to try the
server out, and to play-test when something looks wrong.

## When you need help with it

Settings has two buttons for that.

**Debug logs** writes `debug-logs-<date>.zip` next to the game files, then
opens the folder. It holds the launcher log, SourceMod's logs, the server
console, the player file and your settings, and it leaves every password out.
Send that file to whoever helps you.

**Repair** throws away SteamCMD and the mods and fetches them again. It stops
the server first. It keeps the game files and the run, so no 14 GB comes down
again and the checks stay.

The **Player options** tab also has **Open tf2.yaml**, which writes the player
file from what is on screen and opens it.

## The commands

The window covers an evening. These are for a script, or for a setting the
window does not show. Run them from a terminal: the exe attaches to it.

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

## Settings from the environment

Every setting also reads an environment variable, under the name
`deploy/.env.example` uses. A variable wins over the saved file for that run,
and the launcher never writes it back. Start a server with no questions from a
shortcut or a `.bat` file:

```bat
set AP_ROOM=archipelago.gg:12345
set SRCDS_BOT_TEAM_SIZE=4
tf2ap.exe
```

`AP_ROOM` carries the whole address. `AP_HOST`, `AP_PORT` and `AP_TLS` set the
three parts on their own, for a `.env` file shared with the Docker stack.

`tf2ap.exe -env` prints every name it reads and marks the ones your environment
already sets.

## Where it keeps things

| Path | Holds |
| --- | --- |
| `%USERPROFILE%\tf2-archipelago\` | The game files, SourceMod and SteamCMD. Keep this. |
| `%USERPROFILE%\tf2-archipelago\tf2.yaml` | The player file for the Archipelago app. |
| `%USERPROFILE%\tf2-archipelago\bridge-state\` | The checks and unlocks of the run. |
| `%APPDATA%\tf2ap\config.json` | Your saved settings. |

`TF2AP_INSTALL_ROOT` moves the first three, for a second disk.

## When something goes wrong

The launcher writes its log to the terminal. Scroll up to read what the game
server and the bridge said.

The port is the usual cause. Forward 27015, or the port you set, to your
machine on your router, and open it in the Windows firewall. UDP carries the
game, so a closed UDP port stops everybody from joining.

See [Troubleshooting](../operate/troubleshooting.md) for the rest.

## The other way to run it

If you have a Linux machine with Docker, see [Install](install.md). The two run
the same software. The launcher holds the same bridge and the same plugin,
packaged for Windows without Docker.
