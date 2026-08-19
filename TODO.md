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

## 4. Choose the start mission and the start class

`generate_early` in `apworld/tf2_mvm/__init__.py` takes the easiest mission
drawn as the start mission. It takes the start classes from `random.sample`.
The player chooses neither.

So the server boots a map that is often not the start mission, and the player
must find out how to change it.

Add two YAML options: the start mission, and the start class. Both keep the
current random draw as their default. The launcher already writes
`tf2ap_start_mission`, so the option and the launcher must agree.

## 5. Choose the classes of the bots

`sm_redbots_manager_class_blacklist` says which classes the bots never play.
`configs/defenderbots/loadout.cfg` says what they carry. Neither says what the
team is.

One draw gave a team of three Spies and two Scouts on an Advanced mission.
Another team had no Engineer and lost wave 1 of Quarry twice. The bots ran out
of ammo against the tank, and the small robots swarmed them, because no sentry
covered the team.

Add a team composition: an ordered list of the classes the bots fill from. A
player who asks for an Engineer and a Medic then gets them. This change
belongs in the fork.

## 6. The spectator bug

EZKSupernova reported this. A few minutes away from the keyboard moved the
player to spectator. The server put a bot in their place and refused them RED.
Then it added one more bot, and RED held seven members.

Two faults, and maybe one cause. `Bots_MakeRoom` in
`plugin/scripting/tf2_archipelago/bots.inc` kicks one bot when a human arrives
and RED is full. It counts RED against `tf_mvm_defenders_team_size`. It does
not run when a human comes back from spectator. And nothing caps the team at
six.

Reproduce it first. Idle a client into spectator with the bots on, and watch
the count on RED.

## 7. Keep the cash after a failure

A `Cash Bundle` pays 200 credits per player. The team spends them at the
upgrade station, and the end of the mission takes the upgrades back. A lost
wave then costs the team what the bundle bought.

Two requests. The first one is small:

- A bundle survives a lost wave. The team keeps what it held when the wave
  started.
- A bundle survives the mission that received it. This one is larger. MvM
  resets the upgrades at the end of a mission, so the bridge must hold a
  balance and pay it in on the next one.

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
