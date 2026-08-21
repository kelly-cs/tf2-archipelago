# Changes

What each release changes, for somebody who plays the game. The workflow in
`.github/workflows/release.yml` reads the section matching the tag and puts it
in the release notes, so this file is the only place to write it.

## v1.9.0

The mission switcher works, the bots are a different team to play with, and a
debug bundle is worth sending.

### The run

- Switching mission works. Picking one in the Session tab or with `!mission`
  used to land back on the first mission of the list: the plugin read the map's
  default mission while the switch was still loading, decided the run was not on
  a mission it held, and moved the server off the switch it was making. Only the
  starting mission in the settings appeared to work, because that one is written
  into the server config before any of this runs.
- The bots no longer start the wave without you. A defender bot that is ready is
  made unready again while any player on RED is not, so the wave begins when the
  team is in front of it rather than when the last bot finished shopping.
- A server with nobody on RED never readies at all. It used to play the run by
  itself, wave after wave, and check off locations nobody was there for.
- A player who leaves between waves has their seat filled by a bot.

### The bots

Most of this release is here. The defender bots played like six of the same
mercenary standing near the bomb; they now pick targets, hold ground and spend
credits the way the guides say to.

- Targets in an order: the Medic first, then the Sniper and the Engineer, then
  giants, rather than whatever is nearest. The distances that go with that order
  were measured and corrected afterwards: ten runs on Decoy read 54 defender
  deaths against the previous 43 for the same waves cleared, and the fix put a
  matched pair at 56 against 25.
- Each class carries a named loadout, read out of the game's own item schema:
  Phlogistinator on the Pyro, Beggar's Bazooka on the Soldier, Wrangler and Jag
  on the Engineer, Ubersaw on the Medic.
- Pyros aim above a tank rather than at it, because flames rise and half of
  every puff was going into the ground. Soldiers shoot the ground under a robot
  Pyro instead of the Pyro, because a reflected rocket is their own damage
  pointed at their team.
- Engineers ready at a level three sentry and a level three dispenser, medics at
  a full charge. They used to press ready the moment a sentry existed, and the
  wave began in front of a level one still being hammered together.
- The medigun goes on the biggest body in range rather than whoever the game
  stood next to, which names the Heavy without a list to maintain, and follows
  the health upgrades the team buys.
- Engineers build on the spot they picked, keep rebuilding, split across the
  ground a map names so one holds inside and one holds out, and give up on a
  spot they cannot reach after twelve seconds instead of walking into a boulder
  for the rest of the wave.
- Bots look behind themselves while the team knows a Spy is about.
- Soldiers and demomen buy reload speed first, because their damage is a burst
  and the wait is what the credits buy; it used to rank seventh. Rocket
  Specialist stops after the first tier, which is the one that does anything.
- Blast, bullet and fire resistance are bought when the coming wave carries
  robots that deal that damage. They were ranked flat and near the bottom, and
  explosions are between forty five and sixty percent of every defender death on
  every map measured.
- Sticky traps stack on one spot for a giant instead of carpeting ground for a
  crowd.
- Bots stop buying upgrades for weapons they are not holding. A negative
  priority is a refusal now rather than a low bid, so a Pyro no longer works
  down the list and buys airblast pushback, and Explode on Ignite needs a Gas
  Passer in hand.
- Hats, and unusual effects on them, in two ticks at the foot of the Bots tab.
  A bot keeps its hat for the mission, which is how you tell one Heavy from
  another. Nothing about it changes a wave.
- Any of this can be switched off one at a time with
  `sm_redbots_feature_<name>`, which is how two ways of playing get compared.
- Four server crashes fixed: a table that wrote past its own fields during
  `exec server.cfg`, a full path computation running once per spot per tick
  under the watchdog, a write into the game's own medic action, and the Gas
  Passer.

### The window

- The Session tab is the first one and the log is the second.
- A mission button names the mission rather than its pop file, a locked mission
  says so on the button instead of refusing in chat afterwards, and pressing it
  says what is loading until the server confirms it.
- Saving settings no longer starts a stopped server. Start is the button that
  starts the server.
- Numbers read from the left, and the Bots tab no longer runs its last few
  characters under its own scrollbar.
- Over Steam the join line reads `Steam public IP:` and the Join button goes to
  that address rather than to a local one that means nothing to the friend it
  was sent to.
- The Steam page that issues a login token is a link on the tab that needs one,
  with the two things it asks for: app id 440, and a memo of your choosing.

### When something goes wrong

- The bots' purchases in chat named the wrong upgrade. `mvm_upgrades.txt` has 64
  entries and the game loads 63, because one of them is commented out, so
  everything past it was named after the upgrade before it. The line also says
  how many tiers were bought, which is why credits appeared to go nowhere.
- Debug logs now hold what somebody would need to read them: the game server's
  own console log, which never existed before, the launcher log from the run
  before this one, and what the bridge says about the run.
- The plugin writes what it does to the server console and the SourceMod log by
  default. `tf2ap_debug 2` puts it in the chat as well, which is what `1` used
  to mean.
- Every purchase and sale at the upgrade station is written down, players
  included, with the credits held afterwards.
- Test mode starts on the class the settings ask for. It handed out a Scout
  whatever you picked.
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
