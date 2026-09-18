# The first session

## What a player sees

Eight seconds after joining, each player gets this in the chat:

```text
[AP] This server runs an Archipelago randomizer.
[AP] The run locks the classes and the weapon slots until it finds them. All players share the unlocks.
[AP] Mission: mvm_decoy. Each wave you clear is a check.
[AP] Unlocked classes: scout, medic
[AP] Unlocked slots: primary
[AP] Type !ap to speak to the multiworld. Examples: !ap hint Class: Scout and !ap missing.
```

The two unlock lines show the state of the run at that moment. A player who
joins late sees what the team already found.

## During a wave

- Bots fill RED to six when the wave begins. See
  [The bots on your team](defender-bots.md).
- The class menu refuses a locked class, with a line in the chat.
- Locked weapon slots stay empty at every spawn and every resupply.
- The upgrade station shows the buffs on your loadout. `!ap_buffs` shows
  them again.
- Each cleared wave writes `[AP] Wave 3 cleared.` in the chat.
- Each received item writes `[AP] Unlocked: Class: Pyro` or
  `[AP] The run received 200 credits for 4 player(s).`
- The other players of the multiworld talk in the same chat.
- Anything that goes wrong is written in red.

## Who does what

- **The host** connects like everybody else. The host also runs the launcher,
  switches missions and reads the logs.
- **The players** clear waves. They have nothing to configure. Their commands
  are `!ap` and `!apchat`. See [Chat commands](chat-commands.md).

## The first check

The first cleared wave proves the whole chain works. Watch for three things,
in this order:

1. `[AP] Wave 1 cleared.` in the game chat. The plugin saw the wave.
2. `check recorded` in the launcher log. The check is on disk.
3. `tf2 sent <item> to <somebody>` on the room page. The multiworld has it.
   The [campaign tracker](tracker.md) shows the check as a green box.

If step 1 does not happen, see [Troubleshooting](../operate/troubleshooting.md).
If step 1 happens and step 2 does not, the chat says so in red.

## Which mission plays

The run decides, not the map cycle.

- The server starts on the **Start mission** from the settings.
- If the loaded mission is not part of the run, the server moves to the first
  unlocked mission that is not cleared. It does the same when the run has not
  unlocked the loaded mission.
- When the team clears a mission, the server loads the next unlocked mission
  after 30 seconds. When the team has cleared every unlocked mission, it
  replays one until a ticket opens another.
- The host switches missions with **Play** in the missions table of the
  **Play** tab, or with `!mission` in the chat.

## Ending the evening

Press **Stop**, or `make down` with Docker. The run stays on disk. The next
start continues the same run, with the same checks and unlocks. A run can sit
for a week.

## Watching what the server does

The plugin writes every game event to the console and the SourceMod log. To
see the same lines in the game chat, type this in the rcon box under the log
on the **Play** tab:

```text
tf2ap_debug 2
```

With Docker, use `make rcon`, or the console password from `.env` in the
game's developer console:

```text
rcon_password your-console-password
rcon tf2ap_debug 2
```

Next: [Chat commands](chat-commands.md).
