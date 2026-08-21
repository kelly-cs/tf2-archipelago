# TODO

What the first play-test asked for, and what came of it. Every item says where
it stands.

Two players played that session: EZKSupernova and Cowser the Khelinace. Most
of this list comes from their reports.

Items 1 to 11 are done. The two that say so name what they left behind. Item 9
grew a second half since: Linux gets a window too, and that has not been
written yet. Item 12 is new, from the second play-test, and open. Item 13 is
not from a play-test: it is what players hit before the game even starts.

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

## 8. More checks. Done

The first tank the team destroys in a mission is a check, and so is the first
giant it kills. 55 more locations: 26 tanks and 29 giants. The table has 265
locations now, against 210 before.

`Mission.HasTank` and `Mission.HasGiant` decide which missions get which. Both
come from the wiki's own mission list, which is the source of the whole table:

	https://wiki.teamfortress.com/w/index.php?title=Template:List_of_MVM_Missions&action=edit

Three missions hold no tank, all of them Mannhattan's, which run on gates
instead. Every mission holds a giant. A wrong `true` is a location nobody can
reach and a run nobody can finish, so a test pins both facts.

That list also agreed with every wave count, every tier and every map already
in the table. Those facts now have a second source, where they had one.

The plugin reads the game's own `mvm_tank_destroyed_by_players` for a tank. The
game fires no event for a giant. So the giant check rides on `player_death`
and reads `m_bIsMiniBoss` off the victim. The defender bot mod reads the same
property to pick its own targets. A wave is hundreds of robot deaths, so the
handler tests the cheapest thing first and stops for good once the check is in.

The export is format version 2. A bridge older than this one refuses a seed
generated with it, which is what that number is for.

More checks mean more items. The pool is tickets, classes, three weapon slots
and filler, so 55 more checks mean 55 more `Cash Bundle` across a full roster.
A real item, bot templates or canteens or upgrades, is the next thing this
needs, and it is its own piece of work.

## 9. A native Linux build. Done

`tf2ap-linux-amd64`, one static binary, in every release beside `tf2ap.exe`.
`make launcher-linux` builds it.

The question the note asked was whether Linux gets the window or a terminal.
It got the terminal: walk is a Win32 binding, and the console flow the compose
stack already uses was there to take. So the answer cost nothing.

It is the wrong answer, and the next piece of work is the other one: the same
window on both, which means the same code on both, which means leaving walk
for a toolkit that builds for Linux. That is not free. Fyne and Gio both need
cgo, so `tf2ap.exe` stops cross-compiling from a Linux box with
`CGO_ENABLED=0` and needs mingw in the build and in CI; the Linux build needs
X11, GL and Wayland headers; and the binary roughly doubles. The 1841 lines of
walk in `launcher/internal/gui` all get rewritten. Fyne is the closer fit: its
widgets line up with what the dialog already uses, where Gio's immediate mode
would hand-roll six tabs of form state.

Until then the pictures of the window come from Wine. `make window-captures`
runs `tf2ap.exe` on a virtual display and photographs it, so what the README
shows is what the last build draws.

What it did cost was the installer, which knew only Windows. SteamCMD and both
AlliedModders drops ship a zip for Windows and a tarball for Linux. So the
unpacker reads the bytes rather than the file name, and takes either.

The launcher embeds one platform's mod binaries per build, chosen by a build
tag. SourceMod loads the .so or the .dll and ignores the other, so both in one
binary is half of every download wasted.

Valve's tarball has no directory entry for `linux32/`, only files under it. A
synthetic archive in a test did not catch that; unpacking the real one did.

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

## 12. A bundle spent on upgrades leaves the money negative. Open

From the second play-test, and it is the other end of item 7. A bundle pays at
the upgrade station now, which is where the money is meant to be spent. Spend
it, then lose the wave, and the count goes negative. The upgrades bought with
it stay.

What the game does on a lost wave is put every player's credits back to what
it recorded at the start of the wave. `MvM_GrantCredits` in
`plugin/scripting/tf2_archipelago/mvm.inc` writes `m_nCurrency` straight onto
the player, so the bundle never reached whatever the game restores from. The
restore then hands back a number that does not include the bundle, the
upgrades were paid for out of money the game does not believe existed, and the
difference shows up as a negative balance. Nothing un-buys the upgrades,
because the game has no reason to think anything was wrong.

Writing the property is the fault. The money has to go in through the path the
game itself uses to hand out credits, so that the same bookkeeping that
records a wave-start balance records the bundle too. That is
`CTFGameRules::DistributeCurrencyAmount`, or `CTFPlayer::AddCurrency` beneath
it, and reaching either means a signature in `gamedata/` and an SDK call.
Neither is exposed to a plugin today, which is what "we might need another
server mod for this" amounts to: either our own gamedata for those functions,
or a mod that already carries them.

Before writing any of it, reproduce and read the numbers: grant a bundle
between waves, note the balance, spend part of it, lose the wave, note the
balance again. That says whether the restore is a set to a recorded value or a
refund of what was spent, and the two want different fixes.

The negative balance is the visible half. The invisible half is that a bundle
paid this way is probably not counted anywhere else the game counts credits
either, the end-of-mission tally included.

## 13. Sign the Windows exe. Open

Players report SmartScreen blocking `tf2ap.exe` and Defender quarantining it.
Nothing is wrong with the binary. The launcher unpacks embedded archives into
the TF2 directory, writes Metamod's and SourceMod's DLLs there, downloads a
game server and starts it. Behaviourally that is a dropper, and an unsigned
binary gives a scanner nothing to weigh against the heuristic. Go makes it
worse only in that a static Go binary is a shape malware uses a lot; rewriting
in another language would change nothing.

The fix is a code signature. The application to the SignPath Foundation, which
signs open-source projects for free, is in. `release.yml` already carries the
signing step, skipped until `SIGNPATH_API_TOKEN` exists, and it runs before the
checksums and the attestation so those describe the signed exe. Granting needs
`SIGNPATH_API_TOKEN` as a secret and `SIGNPATH_ORGANIZATION_ID` as a variable.

The application is a long shot and should be treated as one. SignPath asks for
an established project, and this one was four days old with no stars and about
130 downloads when it applied. A rejection is the likely answer, and the
fallback is a paid certificate.

Azure Trusted Signing is the paid fallback at roughly nine euros a month, but
it wants three years of verifiable identity history from an individual. Do not
buy a certificate before SignPath answers.

Two things are already done and do not need repeating here. The exe carries a
VERSIONINFO resource and an icon, so it is no longer an anonymous blob to the
heuristics, and every release publishes SHA-256 sums. Neither removes the
warning: only the signature does, plus the download reputation that accrues
behind it.

One thing is worth doing regardless of SignPath. Submit each release to
[Microsoft's false-positive form](https://www.microsoft.com/en-us/wdsi/filesubmission)
as a software developer. It is free and they usually clear a file within days.

Signing `tf2ap.exe` does not cover the files it writes. If reports come in
about Defender eating the extracted SourceMod DLLs instead of the launcher,
the answer is a documented exclusion for the TF2 install path in the Windows
guide, not more signing.
