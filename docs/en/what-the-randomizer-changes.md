# What the randomizer changes

This page assumes the vocabulary of
[Archipelago for MvM players](archipelago-for-mvm-players.md).

Everything below happens on the server. The players install nothing.

## The classes start locked

The nine mercenaries are nine separate items. A run starts with one to four of
them, depending on how hard the easiest mission of the run is.

| Tier of the starting mission | Classes at the start | Weapon slots at the start |
| --- | --- | --- |
| Normal | 1 | 1 |
| Intermediate | 2 | 1 |
| Advanced | 3 | 2 |
| Expert | 4 | 3 |

A player who picks a locked class in the class menu gets a line in the chat and
stays where they are. A player who is already on a class that the run has not
unlocked keeps playing the wave and changes class at the next spawn. The plugin
never forces a respawn: in MvM a free respawn hands back the life that dying
cost you.

## The weapon slots start locked

There are three weapon slots: primary, secondary, melee. One item,
`Progressive Weapon Slot`, opens them in that order. There are three copies of
it in the pool.

A locked weapon slot is empty. The server removes the weapon each time the game
gives it to you: at spawn, at the resupply locker, and at the upgrade station.
If the weapon that goes away is the one in your hands, the server puts another
one there, so you are never left holding nothing.

The individual weapons are not randomized. The Scattergun and the
Force-A-Nature are the same item to this run: a primary weapon slot.

## The missions sit behind tickets

Each of the 29 Valve missions has its own `Mission Ticket` item. The tickets
are what the generator uses to decide the order of the run.

The plugin does not refuse a map. If the server runs a mission whose ticket the
run has not found, the chat says so and the waves still count as checks. The
map rotation belongs to you, not to the randomizer.

## A cleared wave is a check

Each wave that the team clears reports one check. Each mission that the team
clears reports one more.

A lost wave reports nothing, and the randomizer adds no penalty. The team
replays the wave, the same as in normal MvM.

The game decides when a wave is lost, and this project does not change that
rule. The robots have to carry the bomb into the hatch. A team wipe on its own
does not lose the wave: the game respawns the team, and the wave goes on.

## DeathLink kills the team, and nothing more

DeathLink is off unless the seed asks for it. With it on, a death crosses the
multiworld both ways.

Outbound: the team loses a wave, and the bridge sends that loss as a death. The
plugin hooks the game's own `mvm_wave_failed`, so what other players receive is
what ended the wave on your screen.

Inbound: a death arrives, and the plugin kills everybody on RED, bots included.
It fires no wave-failed event. Nobody holds the hatch until the team respawns,
so the wave is usually lost, but the game decides that as always.

A wave lost to an arriving death is not sent back out. So one death cannot
travel back and forth between two linked players.

## Credits can arrive as items

`Cash Bundle` pays 200 credits. Each player on RED gets the full 200, so a
bundle that arrives with six players on the server is 1200 credits of upgrades.

A bundle that arrives when nobody is on the server pays nobody. Credits are the
one item in this project that happens once and is over. Classes, weapon slots
and tickets are facts that stay true, so the server can apply them again after
any restart.

## Everybody shares the unlocks

There is one set of unlocks for the whole server. A weapon slot that the run
finds opens for all six players at the same moment. A class that the run has
not found is locked for all six.

This also means that a player who joins in the middle of the evening gets the
current set at once. There is nothing per player to carry over.

## Nobody installs anything

Your friends connect with a standard Team Fortress 2 client. No mod, no
launcher, no account link. Everything they see arrives as chat text from the
server.

The run belongs to the server. It is not stored on anybody's Steam account.

## What does not change

The upgrade station, the credits that robots drop, the canteens, the wave
layouts and the robots themselves are stock MvM. This version randomizes the
classes, the weapon slots, the missions and the money. The individual weapons,
the upgrade lines, the canteens and the traps are not in it.
