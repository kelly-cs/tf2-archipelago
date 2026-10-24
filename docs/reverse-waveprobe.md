# Experimental Reverse MvM wave probe

This investigation is stacked on the daily stability feature and stays opt-in.
The [measured investigation report](audits/reverse-probe-20261024/REPORT.md)
records the supported cases and failures. The daily workflow, README badges,
and default wave plan still exclude Reverse
MvM. This prototype is not a claim of full reverse-catalog coverage.

## Completion evidence

Use `-experimental-reverse` to enable the candidate simulator and
`-reverse-only` with `-mission all` to select reverse missions. The catalog
currently contributes 112 wave/mode cases across eleven missions with navigation.
The reverse mission tagged `no_nav` remains excluded by the normal selection.
Without the opt-in, reverse cases remain `unsupported_reverse`.

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

A reverse pass requires the native `mvm_wave_complete` event, a BLU
`teamplay_round_win`, and a native `teamplay_flag_event` capture by a BLU
participant in that wave. A native RED round win is a wave failure even when
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
  -experimental-reverse -mode both -speed 20 -timeout 180s
```

For an experimental reverse-only plan:

```sh
dist/linuxbed/waveprobe -plan -mission all -mode both \
  -experimental-reverse -reverse-only > reverse-plan.jsonl
```

Keep the first run's JSONL and debug snapshots before retesting a specific
nonpass with the same opt-in and a longer wall budget. The existing daily runner
and automatic retry wrapper do not yet propagate this opt-in. Experimental
results must stay separate from the published daily badges.

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
