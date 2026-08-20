# Changes

What each release changes, for somebody who plays the game. The workflow in
`.github/workflows/release.yml` reads the section matching the tag and puts it
in the release notes, so this file is the only place to write it.

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
