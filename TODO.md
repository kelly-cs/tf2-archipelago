# TODO

## 1. A bundle spent on upgrades leaves the money negative. Open

From the second play-test, and the half of "keep the cash after a failure" that
was left undone. A bundle pays at the upgrade station now, which is where the
money is meant to be spent. Spend it, then lose the wave, and the count goes
negative. The upgrades bought with it stay.

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
