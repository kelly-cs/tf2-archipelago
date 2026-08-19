# TODO

What the first play-test asked for, and what came of it. Every item says where
it stands.

Two players played that session: EZKSupernova and Cowser the Khelinace. Most
of this list comes from their reports.

Only item 9, a native Linux build, is untouched. Item 8 has half of it, and
says why the other half waits.

## 1. Fork the bots mod. Done

The mod is `m-this/tf2-mvm-bots`. Its `tf2ap` branch is upstream tag 1.5.5
plus our commits, and `DEFENDERBOTS_VERSION` names a tag of that branch.
`deploy/patches/defenderbots/` is gone.

The branch carries five changes. The upgrade station crash fix, the server
loadout and the class blacklist came from the old patches. The count of bots
that have not spawned and the team composition are new.

To take a new upstream release: rebase `tf2ap` on the new upstream tag, tag
the result, and bump `DEFENDERBOTS_VERSION`.

## 2. Say what a wave failure does. Done

Two players read the same paragraph and asked the same question: does a team
wipe fail the wave? In normal MvM it does not, and it does not here either.
The docs said it did.

What is true, and what the pages now say:

- A wave ends when the robots deploy the bomb. A team wipe on its own does
  not end it: the game respawns the team and the wave goes on.
- A lost wave reports no check, and the randomizer adds no penalty.
- Outbound, DeathLink sends the game's own `mvm_wave_failed`.
- Inbound, the plugin only kills RED. It fires no wave-failed event, so an
  undefended hatch is what loses the wave.

Fixed in `what-the-randomizer-changes.md`, `shape-of-the-run.md`, `spec.md`,
`CONTEXT.md`, the French copies, and the two plugin comments that carried the
same wrong claim. One spelling throughout: DeathLink.

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

## 8. More checks. Tanks done, giants blocked

The first tank the team destroys in a mission is a check: 26 more locations,
one per mission that holds a tank. The plugin reads the game's own
`mvm_tank_destroyed_by_players`, so the event is not a guess.

`Mission.HasTank` decides which missions get one. It comes from the wiki's
tank health table, which lists every tank spawn by mission and wave. The three
missions absent from it are Mannhattan's, which run on gates. A wrong `true`
there is a location nobody can reach and a run nobody can finish, so a test
names the three.

The export is format version 2. A bridge older than this one refuses a seed
generated with it, which is what that number is for.

Giants are not done, and not guessed. The same question has no answer this
project can source. No wiki page lists which missions hold a giant. The game
fires no event for one either. So the plugin has to read a netprop off a dead
robot to tell a giant from any other robot. Settle both before a giant check
exists.

More checks mean more items. The pool is tickets, classes, three weapon slots
and filler, so 26 more checks mean 26 more `Cash Bundle` across a full roster.
A real item, bot templates or canteens or upgrades, is the next thing this
needs, and it is its own piece of work.

## 9. A native Linux build

`docker compose` runs the whole stack on Linux today, and `tf2ap.exe` is the
Windows path. Players want a Linux binary that does what the launcher does:
it installs the server, runs the bridge, and holds the log.

The launcher is Go. The Windows-only parts sit in `launcher/internal/winproc`,
already split by build tag. The UI needs the same split.

Decide one question first: is the Linux answer the same window, or a terminal
front end over the same runtime?

## 10. Balance for fewer than six players. The measurement exists now

A lost wave used to leave no trace. The plugin reported it, the bridge turned
it into a Death Link, and a seed with Death Link off dropped it. So "which
waves stopped us" had no answer, and there was nothing to tune against.

The bridge now counts every lost wave by mission and wave, whatever the seed
asked for. `wave_failures` in `/healthz` lists them worst first, and
`tf2ap_wave_lost_total` is the same series for a dashboard. It counts from the
last restart of the bridge, and no part of the run depends on it.

The tuning itself is still a human's, and it needs evenings rather than a
commit. Play a solo run against an Advanced mission, read which wave stopped
it, and only then change `SRCDS_BOT_TEAM_SIZE` or the team composition.

## 11. Credits

Thank EZKSupernova and Cowser the Khelinace in the README for the first
play-test.
