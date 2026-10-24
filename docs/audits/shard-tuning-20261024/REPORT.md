# Local shard tuning — 2026-10-24

The balanced default is **12 workers at 50×**, each limited to **2 CPUs and
3 GiB RAM**, on the measured Ryzen 9 9950X3D (16 physical/32 logical CPUs),
61 GiB WSL machine. Sixteen workers are available explicitly for more throughput.

All **798/798 Decoy Expert wave/mode cases passed**. Every worker played all
seven waves with Bot Surge off and on. Robot limiter readbacks stayed enabled,
the invader cap stayed 22, and measured workers recorded no fatal engine errors.
TF2 11138802 still matched Valve's live build after the test.

| Workers | Seed | Passed | Longest probe s | Cases/minute | Gain vs one worker |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1 | 14/14 | 47.4 | 17.7 | 1.00× |
| 4 | 1 | 56/56 | 51.5 | 65.2 | 3.68× |
| 8 | 1 | 112/112 | 57.0 | 117.9 | 6.66× |
| 12 | 1 | 168/168 | 64.2 | 157.1 | 8.87× |
| 16 | 1 | 224/224 | 76.5 | 175.6 | 9.92× |
| 16 | 2 | 224/224 | 76.0 | 176.8 | 9.98× |

Twelve workers reached 157 cases/minute; sixteen reached 176–177. That is about
12% more throughput for 33% more workers. Twelve leaves host headroom. The
sixteen-worker option passed 448/448 cases over two seeds.

Timing includes each Go probe's setup and mission/mode run, and excludes image
preparation, worker boot and the explicit fresh-map reset before the probe.
Cases/minute uses the longest worker probe duration. `phase_seconds` also
retains resets, metric sampling and log collection time. This one-mission
benchmark does not certify full-catalog stability.

The shared benchmark image was 13.55 GiB; all sixteen private write layers
added 3.21 GiB total. No live container or game volume was modified. Both full
community archives were subsequently validated and installed once into an
additional shared layer by the launcher's normal installer for catalog runs.
The source archive files were not modified.

Two fixture defects were caught before timing: Docker recorded binary tar
stdout as expanded JSON logs, and image commit retained a temporary sleep
entrypoint. The helper now disables logging only for its binary transfer,
cleans it up explicitly, and restores the production launch command. Server
logs remain enabled. Setup observations and full logs remain in the original
local audit directory. Raw case records, cvar readbacks, timing/CPU and storage
measurements are committed here. The runner source matches 21ec890; binary
metadata records the earlier working-tree build.
