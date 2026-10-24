# Experimental Reverse MvM wave probe

The daily and local stability sweeps include the bomb/carrier subset through
`-reverse-bombs`: Breaking Walls, Computer Crash, and Denominator. Their 34
wave/mode cases count in the environment percentage. The other 78 reverse cases
remain explicitly unsupported and excluded. The [initial investigation report](audits/reverse-probe-20261024/REPORT.md)
records the simulator validation and payload/control-point gap.

## Completion evidence

`-reverse-bombs` enables only the validated mission list. `-experimental-reverse`
is a separate opt-in for investigating the rest of the catalog; daily checks
never pass it. Use `-reverse-only` with either flag and `-mission all` to inspect
reverse coverage. There are 112 reverse wave/mode cases across eleven missions
with navigation; the reverse mission tagged `no_nav` remains excluded. A bare
waveprobe invocation without either flag keeps all reverse cases unsupported.

The probe uses a BLU fake player and defeats RED enemies using the existing
wear-down fixture. It preserves allied BLU robots and tanks. In reverse missions
it disables the standard probe's kill-on-bomb-pickup fixture so authored allied
carriers can survive; ordinary missions keep that fixture enabled.

For missions allowing human bomb pickup and deployment, the simulator moves the
fake player to an available bomb, then to an enabled BLU capture zone. It uses
normal collision/touch handling and the mission's own pickup/capture permissions.
If the mission forbids human pickup or deployment, it leaves bomb delivery to
its authored allied carriers. It does not change those permissions, enable
capture zones, fire capture outputs, or call winning relays. The ordinary probe's
capture-output stall workaround is disabled for BLU-player tests.

A reverse bomb pass requires the native `mvm_wave_complete` event, a BLU
`teamplay_round_win`, and a native `teamplay_flag_event` capture by a BLU
participant in that wave. Computer Crash wave 6 is an authored combat finale:
it instead requires a native death of `THE CORRUPTED` and `mvm_wave_complete`.
The target-death witness is serialized as `target_kills`; killing unrelated
support bots cannot satisfy it. A native RED round win before that target
death remains a loss. For bomb waves, a native RED round win is a failure even when
SigMod suppresses `mvm_wave_failed`. A completion
without capture evidence is a probe error, not accepted coverage. JSONL includes
objective visit, pickup, and capture counts; raw logs identify capture actors.
The simulator is enabled and checked before every wave because map/plugin reloads
can reset experimental state. Missing simulator support refuses the experiment.
After a completed reverse mission, the controller waits for the game's natural
mission/map transition before loading another mission. Otherwise its queued
changelevel can overwrite a new mission after readiness appears to pass.
Transition timeouts remain explicit load failures; automatic transitions stay
enabled. Reverse mission readiness also waits for SourceMod's
`OnConfigsExecuted` callback. Seeing a map name before server/plugin configs
finish is not sufficient readiness.

## Running it

Build the controller and probe plugin using the same pinned toolchain as the
native stability bed (`make stability-bed`; see [stability setup](stability.md)).
Use an isolated disposable server with both community packs, never a live player
server. The server must use 32 player slots and keep the production robot limiter
enabled. Use the same headless wake fixture as the standard wave probe.

The production Docker entrypoint periodically syncs staged addons into the game.
Bind the candidate probe under
`/opt/tf2-archipelago/addons/sourcemod/plugins/tf2_waveprobe.smx`; updating only
`tf/addons/sourcemod/plugins/tf2_waveprobe.smx` is temporary and can be overwritten.
Verify the staged and installed SHA-256 hashes before and after every trial,
and check the running status includes `configsready`. A plugin reversion
invalidates candidate validation. Use a fixed compiled file for the container's
bind mount; an atomic replacement of a host file can leave a bind mount on its
old inode.

For a single mission:

```sh
dist/linuxbed/waveprobe -rcon 127.0.0.1:27035 \
  -mission mvm_rottenburg_rev_int_breaking_walls \
  -reverse-bombs -mode both -speed 50 -timeout 180s
```

For the daily reverse-only plan (34 planned and 78 unsupported cases):

```sh
dist/linuxbed/waveprobe -plan -mission all -mode both \
  -reverse-bombs -reverse-only > reverse-plan.jsonl
```

Keep the first run's JSONL and debug snapshots before retesting a specific
nonpass with the same profile and a longer wall budget. Hosted and local runners
record `reverse_bombs=1` in their config, and retries preserve that flag. The
manual `reverse-smoke` workflow scope tests Breaking Walls on all three
environments without publishing full-catalog badges. Broad experimental runs
remain separate from the daily profile.

## What this measures

Teleporting the fake player into objective volumes exercises real pickup,
capture permissions, authored hatch locks and engine/SigMod completion logic.
It does not validate walking routes, player survival, gate reachability, or
whether a human can physically navigate to that volume. Allied carriers use
their own AI and navigation. Timers, payloads, multi-stage gates, and objectives
without a flag capture may require additional simulators; they must not be
marked supported solely because RED robots can be cleared.

The local investigation's initial kill-only baseline, candidate snapshots,
per-wave native capture evidence and reports are saved separately under
`docs/audits/reverse-probe-20261024/` in the original workspace. This source branch
does not include the raw game logs or generated binaries.
