# Daily stability

The [daily report](https://github.com/kelly-cs/tf2-archipelago/releases/tag/stability)
measures the fork’s `main` branch on native **Ubuntu 24.04**, **Windows**, and
the **production Docker image** built from that commit. It runs daily at 05:17
UTC. Start it manually through Actions → Daily stability → Run workflow,
selecting `main`.
The README widgets link to the report. They show the measurement date; results
older than 36 hours are stale, including when scheduling stops. A green badge
from a previous day is not evidence about a subsequent TF2 update.

Each of six shards per environment gets a fresh disposable install of TF2 and both community
packs, using the launcher's installer and pinned mod dependencies. Startup
requires RCON, SourceMod, the Archipelago plugin, the wave probe, and no failed
plugin or extension loads. The Docker check runs the production entrypoint,
including its own SteamCMD update; the launcher-prepared tree supplies the
community packs. It never mounts a live server's files.

Startup is checked before the wave runner applies its headless wake-up fixture
(`tf_allow_server_hibernation 0`). The sweep keeps `sig_mvm_robot_limit` enabled
and records its readback before and after testing; a disabled or unreadable
limiter prevents a numeric score. The invader cap, population files, mod code
and 20× speed are unchanged. This fixture models an awake server with a human
connected; it does not validate idle hibernation or real-client wake-up.
Fatal errors and segmentation faults in the captured engine logs invalidate the
score, including when `srcds_run` restarts the engine and later checks succeed.

SteamCMD installs the current public TF2 server, without a game version pin.
The measurement compares `tf/steam.inf`'s `ServerVersion` with Valve's live
`ISteamApps/UpToDateCheck` response before setup, at readiness, after the sweep,
and again when publishing the combined report. A failed update cannot silently
test an old server. If Valve updates during a run, the report marks affected
environments unverified. `TF2_SERVER_VERSION` is recorded by the normal build,
but does not select which TF2 build this job installs.
The engine's own startup version must also match at readiness; the report
includes the dependency pins and both the running and installed version.

## What the score means

The [existing wave probe](waveprobe.md) supplies the plan and the detailed
reports: every eligible mission on every catalog map, every wave, Bot Surge
off and on, at requested 20× speed and seed 1. It screens at 180 real seconds
per wave, then retries every nonpass at 900 real seconds. The game's completion
event is the pass criterion. First-pass records survive retries. Nonpasses
carry spawn/kill/alive counts, progress and timeline/debug evidence.

**Score = passed cases / all eligible planned cases × 100.** Missing results
count against it. Startup failure, an empty plan, a different source commit or
an unverified current TF2 build, an unverified robot limiter, or a recorded
engine crash produces an unverified score. A lost, timed-out
or interrupted run fails the workflow. The report also shows raw coverage and
each Bot Surge mode separately.

The existing detailed wave report calls `passed / tested` its stability score.
The repository dashboard deliberately uses `passed / eligible` so an interrupted
run cannot inflate its badge. Both counts appear in the evidence.

Reverse MvM requires an objective simulator the existing probe does not have;
those cases are counted as unsupported and excluded, never reported as passing.
Missions tagged `no_nav` cannot run and are excluded by the catalog. The score
measures wave completion under this probe's strategy. It does not measure
human client joins, multiworld connectivity, performance, or every bot shopping
and combat path. The separate defender-bot release gate remains useful.

The [local speed comparison](audits/speedprobe-wake-20261023/REPORT.md) tested
Decoy Expert with the wake-up fix and production robot limiter enabled. 50×
passed five full runs across four seeds, averaging 48.4 seconds versus 89.4 at
20×. Higher tested settings intermittently lost the last wave's completion
during automatic map changes. This is one-map evidence; the daily sweep stays
at 20× pending validation across the full catalog and environments.

## Hosted runner setup

The fork has no self-hosted runner requirement. Six mission shards run for each
of Ubuntu 24.04, Docker on Ubuntu 24.04, and native Windows Server 2022, using
standard GitHub-hosted runners. At most twelve measurement jobs run together.
Each shard is deterministic (`CRC32(mission) % 6`) and runs both Bot Surge modes.
The build emits the complete canonical plan; aggregation checks every returned
shard against its assigned partition and reconstructs coverage from that full
plan. A missing or failed shard cannot shrink the denominator or become green.

Before installing TF2, jobs remove optional SDKs from the ephemeral runner and
install Linux’s 32-bit runtime libraries. Cleanup refuses to run on a personal
or self-hosted machine. Disk checks require 35 GiB free for native installs and
45 GiB for Docker. A runner that cannot meet that check produces an explicit
host-setup error; it never silently drops community maps to fit.
Windows uses the roomiest available volume, so a small workspace drive does
not waste space reclaimed on the system drive.

GitHub-hosted jobs have a [six-hour execution limit](https://docs.github.com/en/actions/reference/limits).
Each measurement job is capped at 350 minutes, with a 4.5-hour test/install
budget. Timeouts preserve partial results and remain failures. Six shards are
the initial capacity choice; adjust the matrix, shard arguments and report
shard count together after inspecting a complete run. Standard hosted runner
availability and the fork’s Actions settings still apply.

The fork’s main now includes the parent’s latest main and the stability
implementation. Every report identifies the exact merged source commit.
The report job publishes the full report to a dedicated prerelease, leaving
`releases/latest` alone. Its three badge JSON endpoints are also published in
one commit to `stability-badges`, served through `raw.githubusercontent.com`.
Shields blocks `github.com` release URLs as custom endpoints; raw-content URLs
are supported. The badge branch contains generated JSON only.

Actions defaults remain read-only; only the report job receives `contents: write`
to update the report assets and badge branch. No extra secrets or Steam account
are needed. The prerelease’s assets must remain mutable.

## Evidence and local use

The prerelease contains `REPORT.md`, `stability.json`, three Shields endpoint
JSON files, and `stability-evidence.zip`. The ZIP contains each environment's
full report, raw plan and screening/retest JSONL, startup/final plugin and
extension lists, server logs, SourceMod logs and TF2 manifest/version files.
The Actions run has the same evidence retained for 30 days, so previous daily
measurements remain available even when the latest assets are replaced.

Check out the fork’s `main` first. Build the beds with `BOTS_EXTENSIONS_DIR=/path/to/extensions make stability-bed`.
Then, from the repository root, on the environment being measured:

```sh
export WAVEPROBE_RCONPW="$(openssl rand -hex 24)"
python3 deploy/stability-run.py --environment ubuntu24 --bed dist/linuxbed \
  --root /tmp/tf2ap-stability-fresh --output /tmp/stability-result
```

Use `windows` and `dist/winbed` on Windows. For Docker, build the production
image with the pins as the workflow does, then use `--environment docker`
with `--image tf2ap-stability:current`. The root must not already exist. The
output directory must be empty, to keep previous passes out of a new run. The
script stops its own server after reporting; local runs leave their disposable
files for inspection. The workflow removes only its recorded temporary root
after uploading results. External cancellation or loss of the runner may
require removing that run's temporary files manually.

To combine downloaded environment artifacts placed in `evidence/<environment>/part-<shard>`:

```sh
python3 deploy/stability-report.py evidence dashboard \
  --commit <tested-commit> --run-url <actions-run-url> --check-live \
  --shards 6 --plan <canonical-plan.jsonl>
```

Missing environment artifacts still produce an explicit failing report and
badges. Build/setup failures therefore replace the previous dashboard instead
of leaving yesterday's green result looking like today's measurement.
