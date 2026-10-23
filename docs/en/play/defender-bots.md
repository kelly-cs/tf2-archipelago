# The bots on your team

Team Fortress 2 balances every Mann vs Machine wave for six players on RED.
With two, the robots get through. So the server fills the empty seats with
bots. Nothing to install, nothing to type.

## What they do

- They join RED when a wave begins and stay for the rest of it.
- They pick classes, fight, and buy their own upgrades between waves.
- They ready themselves, so a wave starts when **you** press F4.
- A bot that dies comes back within a second.
- When a friend joins a full team, one bot leaves and the friend takes the
  seat.

They are not human. They spot spies late, and they never do the clever thing
your friend does. They make a wave winnable, and that is what they are for.

## The settings

Launcher page: **Settings**, then **Bots**.

| Launcher | `.env` | Default | What it does |
| --- | --- | --- | --- |
| Fill RED with bots | `SRCDS_BOTS` | on | Off keeps the bots off the field until an admin runs `!addbots`. |
| Fill RED to | `SRCDS_BOT_TEAM_SIZE` | `6` | How many players the bots fill RED to, humans included. |
| One menu per seat | `SRCDS_BOT_TEAM_COMP` | `engineer,medic,heavyweapons,soldier,demoman` | The classes the bots play, in the order the seats fill. |
| Classes, one tick per class | `SRCDS_BOT_CLASS_BLACKLIST` | empty | The classes the mod draws from when a seat is not named. Untick the ones the bots never play. |
| One loadout per class | `SRCDS_BOT_LOADOUTS` | empty | What a bot of each class carries. Empty is stock weapons. |
| Say what they buy | `TF2AP_BOT_UPGRADES_CHAT` | off | Write each upgrade a bot buys in the chat. |
| Cosmetic items | `SRCDS_BOT_HATS` | on | A random hat on every bot. |
| Unusual effects | `SRCDS_BOT_HAT_EFFECTS` | off | A random unusual effect on that hat. |

Every setting here applies at the next map load. **Restart** is the sure way.

### Fewer bots, or none

- Lower **Fill RED to** for a harder run. At `4`, three friends get one bot.
- Turn off **Fill RED with bots** when six of you play.

Change these against a record, not a memory. `wave_failures` in the bridge's
health page names every wave the team lost, worst first. See
[Troubleshooting](../operate/troubleshooting.md#ask-the-bridge).

### The team

Humans take seats before the bots do, so put the classes you cannot do without
first. A team shorter than the empty seats leaves the rest to the mod, which
draws from the classes you allowed.

Bots are poor snipers and spies. The class names are the mod's: `scout`,
`soldier`, `pyro`, `demoman`, `heavyweapons`, `engineer`, `medic`, `sniper`,
`spy`.

### Loadouts

The presets cover the common builds. To make your own:

1. Open **Settings**, then **Loadouts**.
2. Pick a class, then a weapon per slot.
3. Type a name and press **Save**.

The loadout then appears in the weapon menus of that class on the **Bots**
page. Remove a loadout, and any seat that named it plays stock.

### Looks

A bot draws a hat its class can wear and keeps it for the mission, so the hat
tells one Heavy from another. Unusual effects are off by default because six
particle effects on screen for a whole wave is a lot. Neither changes how a
bot plays.

## Changing the team mid-mission

1. Open the **Bots** tab of the launcher.
2. Set the seats.
3. Press **Apply to the running server**.

The mod replaces only the seats whose class changed. The wave continues, and
the bots keep the money they earned. In the game, `!ap bots` opens the same
team as a menu for an admin.

A change made during a wave applies at the next break. A bot removed during a
wave drops its buildings.

## A short team

Valve tunes every wave for six defenders. Two settings help a team that is
short of them:

- **Weapon buffs**, on the **Rewards** page, make the team stronger. They are
  on by default. See [Run options](../setup/shape-of-the-run.md#rewards).
- **Robot health (%)**, on the **Balancing** page, scales the health of every
  robot. At 50, a play-test cleared three waves in eight where the same team
  had cleared none.

## How they play

- A bot holds its distance by what it carries. A Brass Beast closes in, a
  Tomislav holds a lane.
- A bot switches to a weapon that still has ammo rather than walk at a robot
  with an empty one.
- An Engineer nests near the hatch, not at the robots' spawn door.
  `sm_redbots_manager_engineer_nest_depth` says how far up the bomb path it can
  build, as a fraction of the path. The default is `0.4`.
- At the upgrade station, a bot buys damage first, for the weapon in its
  hands. A Medic buys healing, an Engineer buys the sentry. Resistances come
  last.

## Who wrote them

The bots are [OfficerSpy's MvM Defender TFBots][mod], GPL-3.0, plus five
dependencies: CBaseNPC, Actions, TF2Attributes, TF Econ Data and TF2Utils.
The mod ships from [m-this/tf2-mvm-bots-go][fork], where the bots' decisions
are written in Go and the SourcePawn is generated from it.

Report a bot that walks into a wall to that repository, not to this one.

[mod]: https://github.com/OfficerSpy/TF2-MvM-Defender-TFBots
[fork]: https://github.com/m-this/tf2-mvm-bots-go

## On a server that is not this one

Every release attaches `tf2-defender-bots.zip`. It holds the plugins, the
extensions for Linux and Windows, the gamedata and the navigation hints. The
zip roots at `addons/`, so one unzip into `tf/` installs it.

Then set these in `server.cfg`. The launcher and the Docker image do this for
you:

```text
sm_redbots_manager_mode 2
sm_redbots_manager_defender_team_size 6
sm_redbots_manager_min_players -1
```

`mode 2` spawns the bots when a wave begins. `min_players -1` matters: the
mod's own ready-up gate defaults to 3 and counts RED before the wave, where a
solo player has no bots yet. Left on, it blocks the F4 that spawns them.

Next: [Troubleshooting](../operate/troubleshooting.md).
