# Initial hosted fixture observation

Run 40046227486, commit 21ec890, started 50× with twelve balanced shards per
environment. Docker readiness succeeded on TF2 11138802, but the limiter
readback was an unknown command: SigMod was not loaded. The new guard refused
to run or score that unmodded sweep. The production image defaults to loading
selected mods only when its configured mission pool requires them; a headless
probe starting stock Decoy has no seeded mission pool. The fixture now sets
SRCDS_MOD_LOADING=always, matching the native bed's existing setting. No
limiter or production feature was disabled. Smoke validation is separate from
full-catalog scoring and cannot publish the README badges.
