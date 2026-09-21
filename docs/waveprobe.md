# Unattended MvM wave sweep

`deploy/run-waveprobe.sh` starts disposable Docker servers and runs every
catalog mission with and without Bot Surge. It skips missions marked `no_nav`,
which cannot run in the game. The live Compose project is never restarted.

The report marks Reverse MvM missions separately. They force players to BLU
and finish through BLU's objective; defeating RED robots alone cannot prove
their waves can complete. A separate objective simulator is needed for those
missions.

Each standard test starts the real wave, attempts to defeat each observed BLU
bot or tank 15–25 game seconds after it first appears, and requires the game's
`mvm_wave_complete` event to pass.
The fake RED player uses Medic for missions tagged `medic_only`, including
Remedic, and Scout otherwise. A case is one mission wave with Bot Surge either
off or on; the two modes have separate results and scores.
The disposable probe prevents damage to RED participants. This keeps
survival objectives and friendly mission NPCs alive while it measures BLU
spawn clearance and wave completion.
It uses `tf_bot_flag_kill_on_touch` so an unguarded hatch does not turn a
population test into an automatic loss. The spawn delays use a fixed seed and
can be changed for a repeat run.

Run this from a checkout with both community packs extracted under
`community-content/tf/` (see the [community content guide](../community-content/README.md)).
Point the probe at the image used by the running Compose server:

```sh
export WAVEPROBE_SRCDS_IMAGE="$(docker inspect --format '{{.Config.Image}}' tf2-archipelago-srcds-1)"
```

Prepare a disposable base volume once. This copies the installed game files
from the live volume without writing to it:

```sh
docker volume create tf2-archipelago-waveprobe_tf2game_waveprobe
docker run --rm \
  -v tf2-archipelago_tf2game:/src:ro \
  -v tf2-archipelago-waveprobe_tf2game_waveprobe:/dst \
  debian:bookworm-slim sh -c 'cp -a /src/. /dst/'
```

Then run the sweep on a machine with enough RAM and disk for the requested
number of game copies:

```sh
export WAVEPROBE_RCONPW="$(openssl rand -hex 24)"
WAVEPROBE_SHARDS=6 WAVEPROBE_SPEED=20 bash deploy/run-waveprobe.sh
```

`WAVEPROBE_SHARDS`, `WAVEPROBE_SPEED` (1–20), and `WAVEPROBE_TIMEOUT`
(real time per running wave, default `900s`) control runtime. At 20×, 900
real seconds can represent up to five game hours; the deadline never shrinks
to 45 real seconds. Map and wave load have a separate 90-second deadline.
`WAVEPROBE_RUN_DIR`
chooses where the JSONL files and `REPORT.md` go; otherwise a timestamped
directory is made in `docs/audits/`. `SUMMARY.json` carries comparable counts
and the stability score (`passed / tested`). `REPORT.md` puts that score and
verified coverage (`passed / eligible`) first, then separates Bot Surge modes
and lists every issue with a spawn/kill/alive/progress timeline. “Wave failed”
means the game emitted a loss, “Wave 0” means the population manager never
initialized a wave, and “No enemies spawned” means the probe saw no BLU bot or
tank in a standard wave. A wave still active at the wall limit is “Wave timed
out”; that result includes enemy counts and final HUD progress.
HUD progress uses the game's non-support enemy count relative to its highest
observed value. Some custom missions complete with that counter below 100%,
so the game's completion event remains the pass criterion.

A pass is one completion witness at the chosen timing and seed. It does not
prove every authored bot appeared, or that the wave will complete under every
player strategy. Re-run timeouts at a lower speed or another seed before
attributing them to the mission or Bot Surge.

For a second pass, start the stopped disposable Compose projects again and run
`deploy/waveprobe-retest.py` with the run directory, the freshly built runner,
and the comma-separated shard IDs whose servers are available. It reruns only
cases without a pass, using the same real-time wave limit and writing
`retest-<id>.jsonl` alongside the first pass. Then regenerate `REPORT.md` with
`deploy/waveprobe-report.py`. A wave that remains active at that limit needs
inspection; the time limit alone does not prove a deadlock.

An optional fourth argument is a comma-separated list of source shard IDs.
Use it to move unfinished cases onto any free worker servers after their
original shard runners finish.

Each sweep writes its unique Compose project names to `projects.txt` in the run
folder. The runner stops those containers on completion or interruption, but
keeps their copied game volumes for retests. After retesting, remove every
shard container and volume from that run with:

```sh
while IFS= read -r project; do
  WAVEPROBE_RCONPW=cleanup docker compose -p "$project" -f deploy/compose.waveprobe.yml down -v
done < docs/audits/waveprobe-YYYYMMDD-HHMMSS/projects.txt
```

Keep the pristine `tf2-archipelago-waveprobe_tf2game_waveprobe` source volume
until you no longer need new sweeps. It is never used as a shard volume.
