# Changes

What each release changes, for somebody who plays the game. The workflow in
`.github/workflows/release.yml` reads the section matching the tag and puts it
in the release notes, so this file is the only place to write it.

## v1.9.0

Mission switch works, the bots take their seats before the wave starts, and a
debug bundle holds what somebody needs to read it. Most of the release is the
bots.

### The run

- Mission switch works. The plugin read the map's default mission mid-switch and
  moved the server back to the first mission of the list.
- A refund gives your cash bundles back. The plugin counts every credit a bundle
  added and puts it back on top of the refund.
- A lost wave no longer leaves you in the red. Spent bundles go back on top, and
  a balance still below zero goes to zero.
- Your run no longer finishes itself. Another player's `!collect` checks a
  mission clear, and the bridge read that as your win.
- The win now comes off what this server played.
- The bots take their seats before the wave. They arrived as it started, with
  stock weapons, no upgrades and no sentry.
- The bots no longer start the wave without you. A server with nobody on RED
  never readies at all.
- The bots fill the seat of a player who leaves between waves, and make a seat
  for a player who arrives.
- Test mode behaves like a real room: `!ap unlock mission` hands over the next
  ticket, `!ap missing` and `!ap checked` answer, and you start on the class you
  asked for.

### The bots

- They shoot the Medic first, then the Sniper and the Engineer, then giants,
  rather than whatever stands nearest.
- Ten runs on Decoy read 54 defender deaths against 43 before, and a matched
  pair after the fix read 56 against 25.
- A bot with nothing to fight holds the hatch, and a bot the nav mesh refuses a
  path steps toward its target.
- Every class has upgrade paths, and a bot refuses an upgrade it cannot use: a
  Pyro bought airblast pushback with a Phlogistinator.
- Bots buy blast, bullet and fire resistance against the wave to come.
  Explosions cause 45 to 60 percent of defender deaths on every map measured.
- Leftover credits stay in the wallet instead of buying canteens nobody drank.
- Engineers build on the spot they picked and rebuild it there, one nest inside
  and one out, with the teleporter exits apart.
- Engineers stop carrying the sentry mid-wave, stop dropping a half-built nest,
  and give up on a spot they cannot reach.
- A sentry that cannot reach the nest goes down beside the engineer, and the
  disposable sentry goes beside the real one on purpose.
- An engineer with no sentry left rides his own teleporter home.
- A map can name which dispenser spot belongs to which nest.
- Medics walk again. The mod took the walk, the aim and the trigger off the
  game's own medic to pick the patient itself.
- A refused path leaves a bot 120 units forward at a time: on Decoy a medic sat
  10400 units from a patient 400 units away and never closed.
- The game picks the patient again. On Coal Town the beam connects in 61 percent
  of samples, up from 5 to 17 percent.
- Movement between samples went from 0 to 70 units up to 337, and ubers in a
  mission from one to six.
- The mod points the game's own heal action at the biggest body: 75 percent
  connected, and 72 percent of that on a Heavy.
- Medics shop before they follow anybody, hold the wave until the charge is
  full, and stop spending a whole wave's credits on one upgrade.
- Soldiers and Demomen stop blowing themselves up. The Demoman did 2571 points
  of self-harm and four suicides in six waves, against 187 for the next worst.
- The Demoman keeps his distance from a tank hull, and drops the stickybomb
  launcher as a tank weapon.
- He no longer detonates a sticky trap under his own pipes.
- Neither aims at the feet of a robot that stands on them, and the Demoman
  throws a pipe as far as it flies.
- A pipe can leave while the aim still moves, a rocket cannot, which is the
  difference between the two arcs.
- Demomen hold the stickybomb launcher, close to the range their pipes arrive
  at, and put an empty launcher down in a fight.
- The Soldier carries the stock rocket launcher.
- The Pyro walks in instead of parking at shotgun range, and bots pick money off
  the ground.
- Each class carries a named loadout and a seat can name its own, so two
  engineers can hold different weapons.
- A bot draws a random cosmetic, and an unusual effect on it, in two ticks, and
  keeps it for the whole mission.
- War paints are gone. They painted the weapons the upgrade station replaces,
  and the server died when two engineers finished shopping.
- This release fixes several server crashes. Buildings go on ground that exists,
  and a wearable sweep that never ended now ends.
- Nav mesh searches for health, ammo and revive markers no longer run per frame.
- The plugin builds the shopping list once a session, and hands out hats one bot
  at a time.
- `sm_redbots_feature_<name>` switches any of this off one feature at a time,
  which is how two ways of playing get compared.

### The window

- The run is the first tab and the log is the second.
- A mission button names the mission rather than its pop file, and a locked
  mission says so on the button.
- A press says what loads until the server confirms it.
- The Bots tab is three pages: Team, Classes and Looks.
- Each seat is one line of what it plays and what it holds, and teams save by
  name. The class pool is two ticks to a line.
- The title bar says which build it is, which matters when several carry the
  same version.
- Save no longer starts a stopped server. Start is the button that starts the
  server.
- Numbers read from the left, columns line up, and no row runs under the
  scrollbar.
- Over Steam the join line reads `Steam public IP:` and the Join button goes to
  that address rather than a local one.
- The page that issues a login token is a link on the tab that needs one, with
  app id 440 and a memo.
- The terminal launcher follows the window: the run first, the same seats and
  saved teams, the same Steam address.

### When something goes wrong

- Debug logs hold the game server's own console log, which never existed
  before, plus the last launcher log and the crash dumps.
- They also hold what the bridge says about the run, and which defender bots
  played.
- The plugin writes what it does to the console and the SourceMod log by
  default.
- The plugin writes down every purchase and sale at the upgrade station, players
  and bots, with the credits held after.
- The bots named the wrong upgrade in chat. `mvm_upgrades.txt` has 64 entries
  and the game loads 63, so every name past the commented-out one shifted.
- A defender bot version bump built the previous version and said nothing, on
  any machine that built once before, including CI.

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
