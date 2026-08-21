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
