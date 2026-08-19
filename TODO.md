# TODO

What the first play-test asked for. The order below is the order to do the work
in.

Two players played that session: EZKSupernova and Cowser the Khelinace. Most
of this list comes from their reports.

## 1. Fork the bots mod. Done

The mod is `m-this/tf2-mvm-bots`. Its `tf2ap` branch is upstream tag 1.5.5
plus our three commits, and `DEFENDERBOTS_VERSION=1.5.5-tf2ap.1` names the tag
of that branch. `deploy/patches/defenderbots/` is gone.

To take a new upstream release: rebase `tf2ap` on the new upstream tag, tag
the result, and bump `DEFENDERBOTS_VERSION`.

## 2. Say what a wave failure does

Two players read the same paragraph and asked the same question. The paragraph
is in `docs/en/what-the-randomizer-changes.md`, and it says that a wiped team
replays the wave, and that a failed wave reports nothing.

The question was whether a team wipe fails the wave at all, because in normal
MvM it does not. It does not here either. The text must separate three steps:

- The team runs out of respawns, and the game declares the wave lost. That
  rule belongs to the game, and this project does not change it.
- A lost wave reports no check. Nothing else happens.
- Death Link adds one thing. A lost wave sends a death to the multiworld, and
  a death that arrives kills RED, which loses the wave you play.

The same confusion sits in `docs/en/setup/shape-of-the-run.md` under
`MVM_DEATH_LINK`, and in `docs/en/spec.md`. Fix all three, and the French
copies.

## 3. Medic, Engineer and Spy need a different first slot. Done

`Class.SlotOrder` in `gamedata/classes.go` gives each class its own order, and
`g_SlotOrder` in the plugin holds the same table. A test in `gamedata` compares
the two. The Medic opens the Medigun first, the Engineer the Wrench, the Spy
the Knife.

The plugin counts the slots the run holds rather than reading which key
arrived. The bridge and the item pool did not change, so an already generated
seed still means the same thing.

The Engineer's two PDA slots were never locked, so the Wrench brings them.

## 4. Choose the start mission and the start class. Done

`start_mission` and `start_class` are YAML options, `MVM_START_MISSION` and
`MVM_START_CLASS` in the environment. `random` is the default and keeps the old
draw. Naming the Final Boss mission as the start stops generation: clearing it
wins at once, which is not a run.

The Windows launcher has one menu for both. Picking a mission sets the seed's
start mission and the server's boot mission together, which is the only way
the two cannot disagree.

The compose stack now writes its player file with `deploy/player-yaml.py`
rather than a shell heredoc. The heredoc never wrote `excluded_missions` at
all: the variable reached the container and no line of YAML carried it, so the
run drew the mission anyway. A popfile the tables do not know is now an error
before generation.

## 5. Choose the classes of the bots. Done

`sm_redbots_manager_team_composition` takes the classes the bots fill RED
with, in order. `SRCDS_BOT_TEAM_COMP` in the environment, and six menus on the
launcher's Bots tab, one per seat. The fork carries it, `1.5.5-tf2ap.3`.

The list names the whole team, not what to add. Every top-up counts the bots
already on RED against the list first. A seat that empties mid-wave gets the
class that left it.

A team named this way beats the blacklist, and beats the mod's lineup mode.

## 6. The spectator bug. Done, but not reproduced

EZKSupernova reported this. A few minutes away from the keyboard moved the
player to spectator. The server put a bot in their place and refused them RED.
Then it added one more bot, and RED held seven members.

Three faults, one per symptom. Each fix comes from reading the code, not from
a repro on a live server. Watch for this one again.

- Moved to spectator: Team Fortress 2 does that after `mp_idlemaxtime`
  minutes, and nothing turned it off. `server.cfg` now sets `mp_idlemaxtime 0`
  and `mp_idledealmethod 0`, in the container and in the launcher.
- Refused RED: `Bots_MakeRoom` frees a seat when a human arrives, and
  `OnClientPutInServer` was the only thing that called it. A spectator who
  comes back does not connect. A command listener on `jointeam`, `autoteam`
  and `joinclass` now calls it too.
- Seven members: the mod flags a bot as its own on that bot's first spawn.
  Its top-up timer runs every second in the window before that. The timer
  counted the pending bot as absent and added one more. Fixed in the fork,
  `1.5.5-tf2ap.2`.

## 7. Keep the cash after a failure. Done, one half of it

A bundle now pays between waves, at the upgrade station, and not before. A
wave the team loses takes it back to where the wave began, and money paid into
that wave goes back with it. Waiting costs nothing: the upgrade station is
where the money is spent.

The plugin holds an effect it cannot apply and does not acknowledge it, so the
bridge keeps sending it. That also carries the money across a plugin reload,
and it fixes the bundle that used to arrive on an empty server and pay nobody.
Effects stay ordered behind the held one. State grants past it are applied
anyway, because applying one twice changes nothing.

The other half is not done, on purpose. Carrying credits from one mission into
the next means telling the money a bundle paid apart from the money the team
earned. No property of the game says which is which. Re-paying both is free
money every mission, and MvM clears the upgrades at the end of a mission by
design. A bundle that arrives and cannot be paid still waits, so it reaches
the next mission that way.

Nobody reproduced this on a live server. The change is safe under either
rule. If the game does not claw the money back, the bundle simply arrives at
the upgrade station rather than during the wave.

## 8. More checks

The only locations are a cleared wave and a cleared mission. A tank kill and a
giant kill are discrete events, which is what a location needs. `CONTEXT.md`
already names both.

More checks mean more items. The pool is tickets, classes, three weapon slots
and filler, so more checks mean more `Cash Bundle`. So do this work together
with a real item: bot templates, canteens or upgrades.

## 9. A native Linux build

`docker compose` runs the whole stack on Linux today, and `tf2ap.exe` is the
Windows path. Players want a Linux binary that does what the launcher does:
it installs the server, runs the bridge, and holds the log.

The launcher is Go. The Windows-only parts sit in `launcher/internal/winproc`,
already split by build tag. The UI needs the same split.

Decide one question first: is the Linux answer the same window, or a terminal
front end over the same runtime?

## 10. Balance for fewer than six players

Valve tunes every wave for six defenders. The bots exist so that a smaller
team can win. Nobody measured whether they do. A solo run against an Advanced
mission is the case to check.

`tf_mvm_defenders_team_size` and the skill of the bots are the two settings.
Measure before you tune: play a mission solo, note which waves fail, and only
then change a number.

## 11. Credits

Thank EZKSupernova and Cowser the Khelinace in the README for the first
play-test.
