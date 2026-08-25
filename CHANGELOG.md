# Changes

What each release changes, for somebody who plays the game. The workflow in
`.github/workflows/release.yml` reads the section matching the tag and puts it
in the release notes, so this file is the only place to write it.

## Unreleased

### The bots

- A bot seat left on "Let the mod pick" stays where you put it. The launcher
  stored only the classes you named and left the draws out. A class in seat 4
  then played as seat 1, and its weapons went to another bot.
- Unticking a class in the Classes tab now holds even when every seat is left to
  the mod. A team that named nobody let the map's own default lineup through, and
  the mod plays a named lineup whatever the ticks say.
- RED comes back to its size within three seconds instead of at the next wave.
  One request for six bots added nine at mission load, and they stayed until
  somebody restarted the wave.
- The bots mod is on v2.13.2. Most of the play-test reports about the bots are
  fixed there: a partial bot team no longer puts nine bots on RED, the engineers
  and medics no longer stand where the wave left them for the whole break, and a
  class you untick is never played, whatever the map would have picked. The
  sniper standing in spawn was the same freeze as the engineers, and so was the
  engineer freezing after a lost wave or a mission restart.
- The bots also stop crashing the server when nest relocation is switched on,
  which was a nav mesh search with no limit on it.

## v1.9.0

Most of this release is the bots.

### The run

- Mission switch works from the Session tab and from `!mission`.
- A refund returns your cash bundles. The plugin counts every credit a bundle
  adds, and puts that count back on top of the refund.
- A lost wave returns the bundles you spent. A balance below zero goes to zero.
- The win comes off the locations this server played. Another player's
  `!collect` checks a mission clear without ending your run.
- Bots take their seats as soon as somebody joins the server. They shop at the
  upgrade station and build before the wave starts.
- Bots ready up only after every player on RED is ready. A server with nobody
  on RED never readies at all.
- Bots fill the seat of a player who leaves between waves, and free a seat for
  a player who arrives.
- Test mode answers as a real room does. `!ap unlock mission` hands over the
  next ticket, `!ap missing` and `!ap checked` reply, and you start on the class
  you asked for.

### The bots

- Bots shoot the Medic first, then the Sniper and the Engineer, then giants.
- Ten runs on Decoy measure 25 defender deaths, against 56 for the same waves
  without the target order.
- A bot with nothing to fight holds the hatch. A bot the nav mesh refuses a path
  steps toward its target.
- Every class has upgrade paths, and a bot refuses an upgrade it cannot use. A
  Phlogistinator Pyro has no use for airblast pushback.
- Bots buy blast, bullet and fire resistance against the wave to come.
  Explosions cause 45 to 60 percent of defender deaths on every map measured.
- Leftover credits stay in the wallet.
- Engineers build on the spot they picked and rebuild it there, one nest inside
  and one out, with the teleporter exits apart.
- Engineers keep the sentry on its spot through the wave, finish the nest they
  start, and give up on a spot they cannot reach.
- A sentry that cannot reach the nest goes down beside the engineer. The
  disposable sentry goes beside the real one on purpose.
- An engineer with no sentry left rides his own teleporter home.
- A map can name which dispenser spot belongs to which nest.
- The game walks the Medic, and the mod points the game's own heal action at the
  biggest body nearby.
- On Coal Town the beam connects in 75 percent of samples, and 72 percent of
  those samples show a Heavy on the end of it.
- A Medic deploys six ubers in a mission, and moves 337 units between samples.
- Medics shop before they follow anybody, hold the wave until the charge is
  full, and spread a wave's credits over more than one upgrade.
- Soldiers and Demomen keep their distance from a tank hull.
- The Demoman rates the stickybomb launcher as no tank weapon, and holds the
  detonator while his own pipes sit on a hull.
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
- Bots wear no war paint. It painted the weapons the upgrade station replaces,
  and killed the server when two engineers finished shopping.
- Buildings go on ground that exists, and the wearable sweep ends.
- Nav mesh searches for health, ammo and revive markers run less than once a
  frame.
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
- Save leaves a stopped server stopped. Start is the button that starts the
  server.
- Numbers read from the left, columns line up, and no row runs under the
  scrollbar.
- Over Steam the join line reads `Steam public IP:`, and the Join button goes to
  that address.
- The page that issues a login token is a link on the tab that needs one, with
  app id 440 and a memo.
- The terminal launcher follows the window: the run first, the same seats and
  saved teams, the same Steam address.

### When something goes wrong

- Debug logs hold the game server's own console log, the last launcher log and
  the crash dumps.
- They also hold what the bridge says about the run, and which defender bots
  played.
- The plugin writes what it does to the console and the SourceMod log by
  default.
- The plugin writes down every purchase and sale at the upgrade station, players
  and bots, with the credits held after.
- The bots name the upgrade they bought. `mvm_upgrades.txt` has 64 entries and
  the game loads 63, because one entry carries a comment marker.
- A defender bot version bump builds that version, on any machine and in CI.

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
