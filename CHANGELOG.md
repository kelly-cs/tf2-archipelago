# Changes

What each release changes, for somebody who plays the game. The workflow in
`.github/workflows/release.yml` reads the section matching the tag and puts it
in the release notes, so this file is the only place to write it.

## v1.9.0

Switching mission works, the bots take their seats before the wave starts, and
a debug bundle finally holds what somebody needs to read it. Most of the
release is the bots: they used to play like six of the same mercenary standing
near the bomb.

### The run

- **Switching mission works.** Picking one in the Session tab or with
  `!mission` used to land back on the first mission of the list. The plugin read
  the map's default mission while the switch was still loading, decided the run
  was not on a mission it held, and moved the server off the switch it was
  making. Only the starting mission in the settings appeared to work, because
  that one is written into the server config before any of this runs.
- **A refund gives your cash bundles back.** Receive bundles, spend them, press
  refund, and the refund handed back the standard 400. The refund does not
  return what you spent: it restores the balance the game recorded, and a bundle
  was never in that record. The plugin now keeps its own count of every credit a
  bundle added and puts it back on top of the refund.
- **Losing a wave no longer leaves you in the red.** The same restore runs when
  a wave is lost, and it took the bundles you had spent off a balance that never
  held them, which is where the negative money came from. The bundles go back on
  top, and a balance that would still be negative is put at zero.
- **The bots take their seats before the wave.** They used to arrive at the
  moment it started, having never seen the upgrade station, so wave one was six
  bots with stock weapons, no upgrades and no sentry. They now turn up as soon
  as somebody is on the server, shop, and build.
- **They no longer start the wave without you.** A bot that is ready is made
  unready again while any player on RED is not, and a server with nobody on RED
  never readies at all. It used to play the run by itself, wave after wave, and
  check off locations nobody was there for.
- A player who leaves between waves has their seat filled, and a player who
  arrives gets one made for them.
- Test mode behaves like a real room: `!ap unlock mission` hands over the next
  mission ticket, `!ap missing`, `!ap checked` and the rest answer instead of
  saying nothing, and you start on the class you asked for rather than always a
  Scout.

### The bots

**Who they shoot.** The Medic first, then the Sniper and the Engineer, then
giants, rather than whatever is nearest. The distances behind that order were
measured and corrected: ten runs on Decoy read 54 defender deaths against the
previous 43 for the same waves cleared, and the fix put a matched pair at 56
against 25. A bot with nothing to fight holds the hatch instead of standing
still, and one the nav mesh will not path steps toward its target rather than
giving up.

**What they buy.** Upgrade paths for every class, and the upgrades these bots
cannot use are refused rather than ranked last: a Pyro was buying airblast
pushback with a Phlogistinator. Blast, bullet and fire resistance are bought
when the coming wave carries robots that deal that damage, which matters
because explosions are between forty five and sixty percent of every defender
death on every map measured. Leftover credits stay in the wallet instead of
going on canteens nobody drank.

**Engineers.** They build on the spot they picked and keep rebuilding it, split
across the ground a map names so one holds inside and one holds out, and they
put their teleporter exits on different spots. They stop carrying the sentry
around mid-wave, stop dropping a half-built nest wherever the clock caught
them, and give up on a spot they cannot reach instead of walking into a boulder
for the rest of the wave. A sentry that cannot reach the nest goes down beside
them, the disposable sentry is placed beside the real one on purpose, and an
engineer with no sentry left rides his own teleporter home. A map can now say
which dispenser spot belongs to which nest.

**Medics.** The medigun goes on the Heavy, or the biggest body nearby, rather
than whoever the game stood them next to. They shop before they follow anybody,
hold the wave until the charge is full, stop following a patient into the
respawn room, and stop dumping a whole wave's credits into one upgrade.

**Soldiers and Demomen.** They stop blowing themselves up: no aiming at the
feet of a robot standing on them, and the Demoman throws a pipe as far as it
actually flies. A pipe may leave while the aim is still moving, a rocket may
not, which is the difference between the two arcs. Demomen hold the stickybomb
launcher, close to a range their pipes arrive at, and put an empty launcher
down in a fight instead of reloading it. The Soldier carries the stock rocket
launcher.

**Scouts and Pyros.** The Pyro walks in instead of parking at shotgun range.
Money on the ground gets picked up.

**How they look.** Each class carries a named loadout and a seat can name its
own, so two engineers can hold different weapons. A random cosmetic item, and
an unusual effect on it, in two ticks; a bot keeps what it drew for the whole
mission, which is how you tell one Heavy from another. War paints were tried
and removed: they painted the weapons the upgrade station replaces, and the
server died the moment two engineers finished shopping.

**Under the hood.** Several server crashes fixed, buildings go on ground that
exists, and a wearable sweep that could not terminate now does. Several frames'
worth of work came out of every tick: nav mesh searches for health, ammo and
revive markers no longer run per frame, the shopping list is built once a
session, and hats are handed out one bot at a time. Any of this can be switched
off one feature at a time with `sm_redbots_feature_<name>`, which is how two
ways of playing get compared, and a feature that lost its measured A/B was
deleted rather than left in.

### The window

- **The run is the first tab and the log is the second.** A mission button names
  the mission rather than its pop file, a locked mission says so on the button
  instead of refusing in chat afterwards, and pressing it says what is loading
  until the server confirms it.
- **The Bots tab is three pages**: Team, Classes and Looks. Each seat is a line
  of what it plays and what it holds, teams can be saved by name, and the class
  pool is two ticks to a line with the weapons beside each.
- The title bar says which build it is, which matters when several carry the
  same version.
- Saving settings no longer starts a stopped server. Start is the button that
  starts the server.
- Numbers read from the left, columns line up, and no row runs under the
  scrollbar any more.
- Over Steam the join line reads `Steam public IP:` and the Join button goes to
  that address rather than to a local one that means nothing to the friend it
  was sent to. The page that issues a login token is a link on the tab that
  needs one, with the two things it asks for: app id 440, and a memo.
- The terminal launcher follows the window: the run first, the same seats and
  saved teams, the same Steam address.

### When something goes wrong

- **Debug logs hold what somebody needs to read them**: the game server's own
  console log, which never existed before, the launcher log from the run before
  this one, the crash dumps, what the bridge says about the run, and which
  defender bots were playing.
- The plugin writes what it does to the console and the SourceMod log by
  default. Every purchase and sale at the upgrade station is written down,
  players and bots, with the credits held afterwards.
- The bots' purchases in chat named the wrong upgrade. `mvm_upgrades.txt` has 64
  entries and the game loads 63, because one is commented out, so everything
  past it was named after the upgrade before it.
- A defender bot version bump used to build the previous version and say
  nothing, on any machine that had built once, including CI.

## v1.8.2

- `tf2ap.exe` carries the header checksum Windows expects. The Go linker leaves
  it at zero, and zero is one of the things a scanner counts against a file.
- The exe says out loud that it never wants the administrator prompt, instead of
  leaving it to the default.
- Every release now attaches a signed record of which commit and which workflow
  built each binary. `gh attestation verify tf2ap.exe --repo m-this/tf2-archipelago`
  checks a file you already downloaded.

## v1.8.1

- These notes now link the VirusTotal report for `tf2ap.exe` and for the Linux
  binary. The scan already ran on the last two releases, but the links never
  reached the notes.
- The warning about `tf2ap.exe` on the front page is two sentences instead of
  nine, and sits under the download buttons where you meet it.

## v1.8.0

- Medics keep the medigun out. Every robot they could see used to pull them
  onto the syringe gun, which drops the heal and stops the charge building.
- Scouts double jump most of the time, and the second jump goes the other way,
  which is harder to shoot than one long arc.
- Bots that are hurt or low on ammo hold the bomb from a friendly dispenser
  instead of walking off to find a health pack.
- Engineers rate dispenser range far higher and buy it early, now that the
  whole team stands in it.

## v1.7.0

- The server no longer freezes at the end of a wave. Every engineer worked out
  where to move its nest in the same frame; they take turns now.
- Engineers stop holding wave one's nest for the whole mission, and upgrade it
  instead.
- Engineers no longer move their nest between waves. It crashed the server at
  every wave transition, so it is off until it works.
  `sm_redbots_manager_engineer_nest_relocate 1` turns it back on.
- A bot that is hurt or low on ammo guards the bomb from a dispenser beside it,
  instead of walking off to find a health pack.
- `tf2ap.exe` carries an icon and says what it is in its file properties.
- Windows still warns about `tf2ap.exe`. The install guide says why, and every
  release now publishes `SHA256SUMS` and links a VirusTotal report so you can
  check what you downloaded.

## v1.6.0

- The defender bots know six maps by hand: Bigrock, Coal Town, Decoy,
  Mannworks, Mannhattan and Rottenburg. Somebody flew each one and stood on
  every spot, so engineers build where a player would build.
- Engineers move their nest between waves when a better spot opens up, instead
  of holding one place all mission.
- Engineers put the dispenser where it was placed by hand, and keep out of each
  other's way.
- On Rottenburg engineers stay off the tank's path on a tank wave, and use the
  platform spot that only works when a tank is rolling.
- Engineers no longer rebuild a level 3 from nothing every wave when they were
  not going to move anyway.

## v1.5.0

- New defender bots. They evade sentry busters, spy check, deploy the medigun
  by what it is instead of by panic, lay and detonate sticky traps, jump as
  Scout, and aim rockets at the ground when the splash pays.
- Engineers buy for their primary and secondary, and pull the wrangler for the
  shield rather than only for the reach.
- Bots stop walking backwards to ride a teleporter that was not worth it.
- A server with no login token now runs on LAN, which is what it is.
- The launcher says where the server is when Join does nothing.
- The Bots tab scrolls on a short window.

## v1.4.0

- The launcher runs in a terminal with the same tabs as the window, for a
  machine with no desktop.

## v1.3.6

- Stop now waits for the half of the server that holds the ports, so starting
  again straight after works.

## v1.3.5

- Screenshots in the setup guides. Nothing in the game changed.

## v1.3.4

- New defender bots: engineers pick a nest near the hatch.
- Screenshots of the launcher in the guides.

## v1.3.3

- A server set to LAN refused the whole network, including the machines meant
  to reach it. Fixed.
- Giving a login token now means the server is meant to be joined from
  somewhere else, and it is set up that way.

## v1.3.2

- New defender bots: the team you name is the team you get, no phantom
  canteens, no afterburn.
- The plugin says which slot a bot bought an upgrade in.

## v1.3.1

- A native Linux launcher.
- The server runs with a console, binds every interface, and takes its children
  down with it.

## v1.3.0

- The first giant and the first tank of a mission are checks.
- Each class's own first weapon slot opens.
- Choose the mission and the class you start on, and exclude missions you do
  not want in the run.
- Reach the internet over Steam's relay, with no port to open.
- A Join button that puts you on the right server.
- Bots play the loadout they were given, buy damage instead of buying at
  random, and keep your seat on RED.
- Name the classes the bots fill RED with.
- A Cash Bundle now pays where the money survives.
- The run counts the waves the team lost.

## v1.2.0

- Death Link. Dying takes the rest of the multiworld with you, and theirs takes
  you.

## v1.1.0

- A Windows launcher: one file, no Docker, it installs the rest.
- Generate the seed from the launcher.
- Pick the map and the shape of the run from lists.
- Defender bots ship with the server and stay between waves.
- Play the whole stack without an Archipelago room, to try it.

## v1.0.0

First release.

- Mann vs Machine as an Archipelago randomiser: missions and waves are checks,
  and your weapons start locked and arrive as items.
- Chat in game talks to the multiworld.
- A Docker stack that runs the server, the bridge and the game.
