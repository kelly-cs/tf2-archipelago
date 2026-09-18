# Chat commands

You type everything on this page in the normal Team Fortress 2 chat. There is
no client to install and no second window to keep open.

## For the players

| Type this | What the server does |
| --- | --- |
| `!ap` | Print the help |
| `!ap status` | Print the mission, the wave, the unlocked classes and slots, and whether the bridge is connected |
| `!mission` | List the missions of the run: the one that plays, the cleared ones, the locked ones |
| `!ap missing` | List the checks nobody found yet |
| `!ap checked` | List the checks already found |
| `!ap remaining` | List what is left, if the room allows it before the end |
| `!ap players` | List the players in the multiworld |
| `!ap hint Class: Scout` | Ask where an item is |
| `!ap hint_location Doe's Doom Wave 3` | Ask what a check holds |
| `!ap options` | Print the options of the session |
| `!ap help` | Print the help of the room |
| `!apchat nice one` | Talk to the other players of the multiworld |
| `!ap_buffs` | Show the buffs on your loadout |
| `!ap unlock mission` | Test mode only. Hand over the next mission ticket. |

The game server itself answers `!ap status` and `!mission`, so they work when
the room does not answer. Every other `!ap` command goes to the room, which
answers in the chat. Team chat works the same as all chat.

Hints cost hint points, which the session earns from checks. Type the whole
item name, with its prefix. `!ap hint Scout` does not find `Class: Scout`.
The room answers with the name it thinks you meant, so the second try works.

## Commands that are refused

The list above is the whole list. The server refuses any other command:

```text
[AP] That multiworld command cannot be sent from the game. It cannot be undone.
```

Every allowed command only reads. The missing ones change the run, and
nothing undoes them. `!release`, for example, hands every remaining item of
this server to the other players. One line from one player ends the run for
everybody.

## Other refusals

| The chat says | Why |
| --- | --- |
| `Wait a moment before speaking to the multiworld again.` | One player can speak once every three seconds |
| `Too much is going to the multiworld. Wait a moment.` | Five lines at once for the whole server, then one every three seconds |
| `That line is too long for the multiworld.` | A line is at most 300 characters |
| `The bridge has no connection to the multiworld. It refused your line.` | The room is unreachable right now |

A refused line is never queued.

## For the admin

An admin is a Steam id in **Admins by Steam id**, on the **Game server**
settings page. In `.env`, it is `SRCDS_ADMIN_STEAMIDS`. Either form works:
the 17-digit id from a profile URL, or `STEAM_0:1:...`.

| Type this | What the server does |
| --- | --- |
| `!mission 3` | Switch to the third mission of the list |
| `!mission mvm_decoy_intermediate` | Switch to a mission by file name |
| `!ap bots` | Open the bot team as a menu. Pick a seat, then a class. |

The server tells a player who is not an admin no. It refuses a mission the
run has not unlocked for everybody.

A bot change made during a wave applies at the next break, because a bot
removed during a wave drops its buildings.

## For the host

The same commands, and a few more, run from the remote console. In the
launcher, type them in the rcon box under the log on the **Play** tab. With Docker, use
`make rcon`.

| Command | What it does |
| --- | --- |
| `sm_ap_status` | Print the mission, the wave, the game events the server sends, the unlocks, the missions, and the last error |
| `sm_ap_mission` | List the missions of the run. With an argument, switch to one |
| `sm_ap_resync` | Ask the bridge for the unlock set again |
| `sm_ap_buffs` | Show the buffs on the current loadout |

The console is the server itself, so it reaches every command whatever
**Admins by Steam id** says.

### Debug and test commands

These need root admin access. Test buffs last until the plugin, the map or
the run state reloads.

| Command | What it does |
| --- | --- |
| `sm_ap_buff_test <1-80\|effect-key\|all> [levels]` | Add effects to your active weapon. Example: `sm_ap_buff_test projectile-count 3` |
| `sm_ap_buff_give <target> <1-80\|effect-key\|all> [levels]` | Add effects to another RED player's active weapon |
| `sm_ap_buff_slot <target> <1\|2\|3\|primary\|secondary\|melee> <1-80\|effect-key\|all> [levels]` | Add effects to an equipped item by slot |
| `sm_ap_projectile_debug on` | Turn on projectile diagnostics |
| `sm_ap_projectile_debug` | Print the 24 most recent diagnostic lines |
| `sm_ap_projectile_debug off` | Turn off projectile diagnostics |
| `sm_ap_unlock_override <on\|off>` | Allow every class and slot for now, or restore the run's locks |
| `sm_ap_bundle [credits]` | Pay a test Cash Bundle, 200 credits by default |
| `sm_ap_report wave_cleared [wave]` | Report a cleared wave by hand |
| `sm_ap_report mission_cleared` | Report a cleared mission by hand |
| `sm_ap_report death` | Report a lost wave by hand |

In the chat, drop the `sm_` prefix: `!ap_buff_test projectile-count 3`. From
the game's own console, send the command to the server with
`cmd sm_ap_buff_test projectile-count 3`.

`sm_ap_report` with no wave number uses the current wave. Reporting the same
check twice does nothing: the bridge identifies a check by its place.

## The console variables

Set one for the session with `tf2ap_debug 2` in the command box. To keep it
across restarts, edit `cfg/sourcemod/tf2_archipelago.cfg` in the game files.

| Variable | Default | What it does |
| --- | --- | --- |
| `tf2ap_announce` | `1` | Write the cleared waves and the received items in the chat |
| `tf2ap_chat` | `1` | Write what the rest of the multiworld says in the chat |
| `tf2ap_debug` | `1` | `0` writes nothing. `1` writes every bridge call and game event to the console and the SourceMod log. `2` writes them to the chat as well. |
| `tf2ap_bridge_url` | `http://127.0.0.1:24680` | Where the bridge is. Do not change it. |
| `tf2ap_start_mission` | empty | The mission the server starts on. The launcher writes it from **Start mission**. |
| `tf2ap_next_mission_delay` | `30` | Seconds between a mission clear and the next mission. `0` leaves the game's own cycle to it. |
| `tf2ap_bot_upgrades_chat` | `0` | Write what the bots buy at the upgrade station in the chat |
| `tf2ap_bots_wait_for_players` | `1` | Hold the bots unready until every player on RED is ready |
| `tf2ap_bots_backfill` | `1` | Put a bot back on RED when a player leaves between waves |

Errors reach the chat whatever `tf2ap_announce` says.

Next: [The bots on your team](defender-bots.md).
