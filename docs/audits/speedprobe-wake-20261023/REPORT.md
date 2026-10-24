# Decoy speed comparison with the production robot limiter

Source `c462fda6cfaed9074b0ee27ea8f69005f96fa883` with the archived experimental speed-only overlay. Live TF2 `11138802`.

Single isolated Linux Docker worker, four CPUs and 4 GiB RAM. Every run covers seven Decoy Expert waves, normal and Bot Surge. Wake-up fix enabled; robot limiter remains enabled; maximum invaders remains 22. No wave, mission-cycle, kill timing or polling changes.

| Speed | Clean full runs | First-pass passes / planned | Mean wall s | Wall range s | Gain vs 20× | Seeds |
| ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 20× | 3/3 | 42/42 | 89.4 | 88.0–90.0 | 1.00× | [1] |
| 30× | 2/2 | 28/28 | 65.0 | 64.0–66.0 | 1.37× | [1] |
| 40× | 2/2 | 28/28 | 53.0 | 52.0–54.0 | 1.69× | [1] |
| 50× | 5/5 | 70/70 | 48.4 | 48.0–50.0 | 1.85× | [1, 2, 3, 4] |
| 60× | 3/4 | 55/56 | 47.0 | 44.0–50.0 | 1.90× | [1, 2, 3] |
| 65× | 2/3 | 41/42 | 46.7 | 44.0–50.0 | 1.91× | [1, 2] |
| 70× | 1/2 | 27/28 | 48.0 | 46.0–50.0 | 1.86× | [1] |

**Result:** 50× was the fastest tested setting with zero observed first-pass
failures: 70/70 cases over five full runs and four seeds. It averaged 48.4
seconds against 89.4 seconds at 20×: about 46% less wall time, or 1.85× the
throughput. Settings above 50× saved little wall time and were intermittent.
55× was not tested; escalation stopped at the first observed failure, so 80×
and 100× were not run. These results establish a tested candidate, not an
absolute speed limit or a guarantee across the mission catalog.

All three first-pass failures were the final Bot Surge wave. The raw server
logs show `WAVEPROBE passed wave=7`, followed by an automatic change to Coaltown;
the controller's next status read saw the new map and reported a probe error.
Each targeted diagnostic retry passed. Those retries remain separate and do
not turn the failed first-pass runs into clean runs. The passing runs all
recorded the same eight tank spawns; robot counts varied between controls.
There were no recorded segmentation faults, fatal engine errors or access
violations during the measured sessions. Robot limiter and invader-cap
readbacks stayed at 1 and 22 respectively before and after every trial.

First-pass failures and missing cases remain failures in this comparison. A representative nonpass receives a separate 900-second diagnostic retry; that retry never changes the first-pass result. Raw JSONL retains completion events, enemy counts, timelines and failure snapshots.

Wall times include probe mission/mode setup and polling, and exclude the initial installation and explicit fresh-map reload before each trial. Spawn totals vary even between controls, so successful completion does not prove identical population behavior.

This bounds speeds for this one map and this local worker. It does not certify all maps, Windows, Ubuntu native or real human join/wake behavior. Daily speed remains 20× until wider coverage is measured.

Setup caveat: the first cold start lacked custom plugins until an explicit restart after staging; see SETUP-NOTES.md and cold-start.log. This is not a clean cold-start certification. No bridge or multiworld is connected, so bridge errors and later objective-queue overflow messages remain in the raw logs; this measures server wave execution only.

Finished: 2026-10-23T23:58:06.916796+00:00. Latest-build verification: True.
