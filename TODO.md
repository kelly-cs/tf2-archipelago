# TODO

## 1. A bundle is money the game never recorded. Fixed, needs a play-test

Two reports, one cause. Spend a bundle on upgrades and lose the wave, and the
balance goes negative. Receive bundles, spend them, press refund, and the
refund hands back the standard 400: "on wave start I had like 1200 after
receiving bundles, spent it, and then clicked refund I had the standard 400."

`MvM_GrantCredits` wrote `m_nCurrency` straight onto the player. That puts a
number on the screen that the game's own bookkeeping never saw, and both of
those buttons read the bookkeeping rather than the number. The refund does not
give back what was spent: it restores the balance the game recorded at wave
start, which was 400 because the 800 in bundles was never part of it. Losing a
wave restores the same record, which is where the negative balance came from.

The fix is to hand the money over the way the game does. `MvM_DropCredits`
spawns an `item_currencypack_custom` worth the bundle at the player's feet, and
picking it up goes through the game's own currency path, so the record includes
it. `m_bDistributed` stays off: that flag means the team has already been paid,
and a pack carrying it pays nobody. That is what `collectmoney.sp` in the bots
mod reads it as, which is where the meaning was checked.

Not yet played. What to watch:

- the refund after a bundle should hand back what the run actually held
- a lost wave should not go negative
- the pack is collected by walking over it. It is dropped at the player's own
  feet, between waves, where they are standing at the upgrade station, so the
  touch should be immediate. If a bundle ever goes uncollected the money is
  gone and the grant was acknowledged anyway, which is the one thing this
  design can get wrong that writing the property could not
- the A+ rating counts cash dropped against cash collected. A pack that is
  dropped and collected moves both halves and leaves the rating alone; one that
  is missed lowers it, which is honest but worth seeing once

The other half is `CTFGameRules::DistributeCurrencyAmount`, which is what the
game calls itself. It wants a signature in our own gamedata for two platforms
and upkeep across every Team Fortress 2 update. Worth it only if the pack turns
out to be wrong.

## 2. Sign the Windows exe. Open

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

## 3. Four optional unlocks from Peppy's post. Open

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

## 4. The bots only turn up when the first wave starts. Fixed here, not in the mod

`deploy/srcds-entrypoint.sh` runs the mod in AUTO_BOTS, where
`ManageDefenderBots` is called from the mod's `mvm_begin_wave` handler and from
a one-second monitor that returns early unless the round is running. So between
the map loading and the first wave there was nobody on RED but the players, and
the bots appeared at the moment the wave did.

What that cost is the upgrade station. A bot that spawns at wave start has not
shopped, so wave one was played by six bots with stock weapons and no upgrades,
and the engineer had nothing built because he had not existed long enough to
build it. Reported from a play-test as the bots only joining on F4.

`Bots_Fill` in `plugin/scripting/tf2_archipelago/bots.inc` now tops RED up
every three seconds between waves, once a player is on the server. It is the
same `sm_addbots` the disconnect backfill already used.

Two things make it safe, and both were already here:

- a bot that turns up before the players is held unready until every player on
  RED is ready, and never readies at all with nobody on RED. Otherwise filling
  the team early means the bots starting wave one by themselves while the first
  player is still choosing a class.
- a seat is held for every player who is connected and not on RED yet, or this
  fights `Bots_MakeRoom`: that frees a seat the moment a player connects, and a
  fill three seconds later would put a bot back in it before they had finished
  joining.

Still worth doing in the mod one day, and that is where it belongs: AUTO_BOTS
fills on `mvm_begin_wave` because that is when the mode was written to fill, and
filling between rounds as well is a change to `Timer_ManageDefenderBots`. What
is here is a plugin asking for something the mod is meant to manage.

## 5. De-upgrading appears not to take. Open, and it may be item 1

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

## 6. The bot upgrade chat named the wrong upgrade. Done

From the same play-test: the chat showed defender bots buying upgrades their
class cannot have, while inspecting those bots showed the right upgrades on
them. So the purchases were fine and the line describing them was not.

`sm_dump_upgrades` in `tf2-mvm-bots` walks `CMannVsMachineUpgradeManager` and
prints the index the game holds each upgrade at. It holds 63.
`scripts/items/mvm_upgrades.txt` has 64 `attribute` lines, because entry 14,
`heal rate bonus`, is commented out line by line and `Bots_LoadUpgradeNames`
counted it. Indices 0 to 18 were right; from 19 on every purchase was named
after the upgrade before it, which is 44 of the 63.

The fix is a comment strip before the split: a `//` outside quotes ends the
line, which is what KeyValues does and what the file expects, since it also
puts a note after a value. The same parse over the file the server ships now
gives the game's 63 names, in the game's order.
