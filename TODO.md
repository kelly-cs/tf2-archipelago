# TODO

Bugs first. Every feature below is off by default and none of them ships before
the bugs above it, because a feature built on broken money or a broken lineup
inherits the bug.

## Bugs, most impactful first

### 8. RED gets nine bots when the mission loads. Half fixed

Peppy: "the server is putting 9 bots on red team sometimes, it happens when the
mission is loaded but once the wave is restarted it fixes itself."

One ask adds more than it asked for, and it is the mod that does the adding.
`sm_addbots N` runs `AddBotsBasedOnLineupMode(N)`, which calls
`AddBotsFromTeamComposition(N)` first: that adds the seats the named team still
wants and reports success only when it filled all N. A team that does not name
every seat cannot, so the caller falls through to the lineup mode and adds N
more on top. Three named seats and an ask for six is three plus six.

Which is nine, and it is nine at mission load rather than mid-mission because
that is the one moment RED is empty and the whole six are asked for at once.

Why the wave restart cleared it: the only trim this plugin ran was scheduled
three seconds after `mvm_begin_wave`, so between the map loading and the first
wave nothing counted the team at all.

Fixed here, which is the half this repository owns. `Bots_Fill` runs
`Bots_TrimSurplus` on its own poll now, so RED comes back to its size within
three seconds wherever the surplus came from, and the debug log says who left.

The other half is the mod's and it is the actual bug:
`AddBotsBasedOnLineupMode` has to subtract what the composition already added
before it falls back. It is filed as tf2-mvm-bots item 18, with the test-bed
repro, because a fix there is worth more than a trim here: a bot that is added
and kicked three seconds later still spent a seat and still cost the wave its
upgrade time.

### 9. "Let the mod pick" picks classes that are unticked. Fixed here

Peppy: "the bot seats I set as 'Let the mod pick' still picks from classes that
I have unchecked in the Classes tab."

Two faults, both in the launcher, both about what a draw seat is worth.

A seat is a place in `sm_redbots_manager_team_composition`, counted from one,
and the mod reads a seat it cannot parse as a seat it should leave to the
lineup mode. Both settings pages stored the classes somebody named and dropped
the draws, so naming seat 4 and leaving the first three to the mod stored one
class, and the mod filled it as seat 1. The loadout file numbers seats from the
same list, so the Wrangler went to whichever engineer came first rather than to
the seat that asked for it.

And a team of nothing but draws stored nothing, which is an empty convar, and
an empty convar is what makes the mod fall back to the map's own default
lineup. It then treats that default as a team somebody named, and a named team
beats the blacklist: `IsBotClassBlacklisted` returns false for any class the
composition mentions. So the map's guess played the classes the Classes tab had
unticked, which is exactly the report.

Fixed. A draw seat is an empty entry rather than a missing one, end to end: the
window, the console form, `SRCDS_BOT_TEAM_COMP` and `SRCDS_BOT_SEAT_LOADOUTS`
through `SplitSeats`, and `Composition`, which writes `engineer,,heavyweapons`
and keeps seat 3 on 3. Trailing draws name no seat and are still dropped. When
every seat is a draw and some class is unticked, the seats go out as holes
anyway, so the string is not empty and the map's guess never gets to overrule
the ticks.

Tested where the bug was: `botloadout` for the convar and for the loadout
file's seat numbers agreeing with it, `settings` for the environment
round-trip, `tui` for a seat keeping its number when the seat before it goes
back on the draw, and `gui` for what the Bots tab means, which is now a plain
value the Win32 widgets are read into so it can be tested off Windows.

What is still open is the mod's own rule, and it belongs there: a lineup the
map guessed at should not beat a blacklist somebody set. tf2-mvm-bots item 7.

### 1. A bundle is money the game never recorded. Half fixed

Two reports, one cause. Spend a bundle on upgrades and lose the wave, and the
balance goes negative. Receive bundles, spend them, press refund, and the refund
hands back the standard 400: "on wave start I had like 1200 after receiving
bundles, spent it, and then clicked refund I had the standard 400."

`MvM_GrantCredits` writes `m_nCurrency` straight onto the player. That puts a
number on the screen that the game's own bookkeeping never saw, and both of
those buttons read the bookkeeping rather than the number. The refund does not
give back what was spent: it restores the balance the game recorded at wave
start, which was 400 because the 800 in bundles was never part of it.

Handing the money over the way the game does was the plan, and the pack is out.
Three attempts on a live server:

- `SetEntProp(pack, Prop_Send, "m_nAmount")` throws. A throw aborts the whole
  callback, so the bundle was never paid, never acknowledged, and asked for
  again on the next poll: eighty exceptions and two hundred and forty dead packs
  in ten minutes.
- `Prop_Data` for the same name is not there either. That one said so once a
  bundle and paid nobody, which is the spam a play-test reported.
- A probe over `m_nAmount`, `m_nCurrencyAmount` and `m_iAmount`, both tables,
  answered on 2026-08-22: *"A currency pack here has none of the amount
  properties."* `item_currencypack_custom` cannot be told what it is worth
  through a property, and that route is closed.

So the money keeps going on with `m_nCurrency`, and the bookkeeping is kept
here instead. `g_BundleCredits` is a per-client ledger of everything a bundle
ever added, cleared on a map change. The station's refund command is caught and
the ledger is put back on top of the balance the game restored, one frame
later, because the game writes that balance during the command.

What is left:

1. The refund command's name is a guess: `MVM_Respec`, which is what the
   community documents the button as sending. Any other command the station
   sends is now written to the log once, so the next debug bundle either shows
   nothing (the guess was right) or names the real one.
2. The lost wave is covered by the same ledger, on `mvm_wave_failed`, and a
   balance that would still come back negative is put at zero. Watch a debug
   bundle for the line saying what each player came back to: if the game
   restores after the event rather than during it, the frame's delay is not
   enough and it wants a short timer instead.
3. `CTFGameRules::DistributeCurrencyAmount` is what the game calls itself and
   would make the ledger unnecessary. It wants a signature in our own gamedata
   for two platforms and upkeep across every Team Fortress 2 update, which is
   why it stays last.

### 5. De-upgrading appears not to take. Open, and it may be item 1

From the play-test: buy an upgrade, sell it again before leaving the upgrade
station, and it stays bought. Reported as "not sure if it is the AP or MvM in
general", and that is still the open question.

Nothing in this plugin touches a purchase. `Unlocks_EnforceSlots` removes
weapons in slots the run has not opened and leaves every other slot alone, and
the upgrade command is only read, never blocked.

The hypothesis worth testing first is item 1 above. `MvM_GrantCredits` writes
`m_nCurrency` straight onto the player, so after a Cash Bundle the player's
credits and the number the game recorded for the wave disagree. A refund pays
back through the game's own accounting, and accounting that has been written
around from outside is exactly the kind that refuses a transaction while
showing it as accepted.

That gives a test that costs one evening: sell an upgrade on a run where no
Cash Bundle has landed yet, and sell one after a bundle has. If only the second
fails, this is item 1 and not a Valve bug, and fixing item 1 fixes both.

The plugin now writes every purchase and every sale to the SourceMod log,
players included, with the credits held afterwards. So a bundle from the
evening says whether the sale reached the game at all, which is the half nobody
could see: a sale the station accepted and the game ignored looks the same as
one that worked, from the chair.

### 7. The goal read a check somebody else made. Fixed in the bridge, open in the apworld

A play-tester was told their run was complete having beaten three of the five
missions they drew. Another player in the room finished, ran `!collect`, and
Archipelago checked every location holding that player's items, mission clears
among them. The bridge adopts the room's checked list on connect, which is what
makes a lost state file survivable, and it read the win off that same list.

Fixed by keeping the two apart. `Played` is the locations this server checked
itself, only the plugin writes it, and the goal is read off it. `Checks` still
holds everything the room says, because that is what `!checked` and a recovered
state file are for. State format 3; a file written before it takes its checks as
played, since the run in it was played by somebody.

What is still open is the report's own suggestion, and it is the better shape:

- Lock a trophy item onto every mission clear, an **Australium Medal**, so no
  other player's item is ever under one and `!collect` cannot touch it. The goal
  becomes `state.has("Australium Medal", player, target)` rather than
  `can_reach_location`, which is what generation logic wants anyway.
- It costs the multiworld eight locations' worth of other people's items, which
  is a real loss: a mission clear is one of the better checks this world offers.
- It needs an item id, so it is a gamedata change on the Go side, a regenerated
  `data.py`, and a data format bump. That is why it waits for the next time the
  format moves rather than going out on its own.

## Distribution

### 2. Sign the Windows exe. Open

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

## Features. All requested, all off by default

Every item here is a feature and not a fix. Each one gets its own YAML option in
`apworld/tf2_mvm/options.py`, off by default, the same shape as `death_link`, so
a seed that does not ask for it gets none of it in the pool and an existing seed
keeps meaning what it meant. The progressive ones are a count rather than a
boolean: zero copies is off, N copies is how many steps the run walks.

None of them starts before the bugs above. Two of them are blocked on item 1 by
construction, and the rest would be built on a lineup that item 9 says does not
do what the settings page asked for.

### 3. Feature: four optional unlocks from Peppy's post. Off by default, open

From the second post in
[`docs/en/discord-mvm-thread.md`](docs/en/discord-mvm-thread.md). Not started,
and none of them is in `gamedata/items.go` yet.

Four items, one shape: each one is a plugin-side effect on a server-wide
value, and each one gets its own YAML option, off by default, the same shape
as `death_link` in `apworld/tf2_mvm/options.py`. A seed that leaves them off
gets none of them in its pool, so an existing seed keeps meaning what it meant.
Three of the four are progressive, so the option is a count and not a boolean:
zero copies is off, N copies is the number of steps the run walks.

Order to build them in is the order below, easiest first. The first two are one
value each and are worth doing before the pool grows further. The last one is
not designed yet.

**Progressive Respawn Time.** RED starts on a longer respawn wave time and each
copy shortens it back toward the stock value. The candidate is the team respawn
wave time that MvM already sets per team, not `mp_respawnwavetime` alone;
confirm which one MvM actually reads before wiring it. Felt every wave, one
value, no interaction with anything else in the item table.

**Progressive Random Crits.** The team starts with random crits off and each
copy raises the chance. `tf_weapon_criticals` and `tf_weapon_criticals_melee`
are the candidates, but they are booleans, so a chance ramp needs either a
per-shot hook or a coarse ladder of off, melee only, on. Decide which before
promising a percentage in the option's docstring. It hits classes unevenly:
Heavy and Demo notice, Sniper barely does. That is a difficulty knob, not a
fair one, which is fine for a randomizer as long as the docstring says so.

**Grappling Hook.** One convar, `tf_grapplinghook_enable`, granted once, so a
state grant and not a progressive. `useful`, not progression: no access rule
may ever need it, because a wave has to stay winnable without it. Cheapest of
the four and the largest change to how a map plays.

**Progressive Returning Bomb.** The bomb walks itself back toward the robot
spawn while no robot carries it, and each copy speeds that walk up. The most
original of the four and the only one that changes the shape of a wave rather
than a number. It is also the least specified. Before designing it, find out
what MvM's existing idle-bomb behaviour is and whether a plugin can reach the
value behind it; if the answer is that the bomb only resets on a timer the map
owns, this becomes a different feature. Note also that it deflates any
bomb-reset location group, so the two should not ship into one seed without
someone thinking about it.

Not in this batch, and why. Progressive credit amount is blocked on item 1
above: money granted by writing `m_nCurrency` does not survive the wave-loss
restore, so a credit-scaling unlock inherits the negative balance. Random-effect
canteens belong in the trap pool that `spec.md` already describes, not here.
Randomized weapons with tiers need the tier to be the item id, because ADR 0001
forbids an id whose meaning is rolled per seed, and that is a feature of its
own size.

### 10. Feature: per-class weapon slot unlocks. Off by default, open

CaptPurpleHeart: "I personally would love per-class weapon slot unlocks."

Today a slot unlock is one item for every class: `Unlocks_EnforceSlots` strips
a weapon from a slot the run has not opened, without asking who is holding it.
Per class it is nine items a slot instead of one, which is a much larger pool
and a much longer run, and both of those are the point of asking for it.

It is one option, a boolean, off by default: with it off the run keeps today's
one-item-per-slot unlocks and nothing about an existing seed changes.

It is a gamedata change on the Go side, so it wants an item id per class per
slot, a regenerated `data.py` and a data format bump, and it should ride the
same bump as the Australium Medal in item 7 rather than moving the format
twice.

The open design question is what a run starts with. Nine classes times three
slots with nothing unlocked is a team that cannot shoot; the honest shape is
probably that the primary of the class you spawn as is free and everything else
is earned, but that is a generation-logic argument and it needs writing down
before any id is minted.

### 11. Feature: Progressive Starting Money. Off by default, open

CaptPurpleHeart: "an option to have 'Progressive Starting Money' items, where
at the start of every Wave 1, you get X amount more starting money per every
progressive starting money upgrade found."

Same shape as the three progressives in item 3: a YAML option that is a count,
off at zero, so a seed that leaves it out keeps meaning what it meant.

It is blocked on item 1 for the same reason progressive credit amount is.
Money handed over by writing `m_nCurrency` is money the game's bookkeeping
never saw, so a starting grant inherits the negative balance on a lost wave and
the refund that hands back 400. Item 1's ledger covers a bundle; a starting
grant is the same money by a different door, and it wants the same ledger,
which is one more argument for `DistributeCurrencyAmount` over writing the
property.

Note also that this one interacts with the wave-one restart: "at the start of
every Wave 1" is a mission start, and a mission restarted from the vote menu is
a wave one too. Decide whether restarting pays again before implementing, or
the answer will be decided by whichever event handler happened to be wired.
## Done, kept for the record

- The bots only turn up when the first wave starts. `Bots_Fill` tops RED up
  every three seconds between waves so the bots shop and the engineer builds.
- The bot upgrade chat named the wrong upgrade. `Bots_LoadUpgradeNames` counted
  a commented-out `attribute` line, so 44 of the 63 names were off by one.

