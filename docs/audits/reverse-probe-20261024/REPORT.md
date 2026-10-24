# Reverse MvM probe investigation — 2026-10-24

An opt-in bomb/objective prototype stacked on stability PR #243. The daily workflow and published scores still exclude reverse cases. This is partial support, not complete reverse-catalog validation.

## Final candidate validation

One isolated production Docker worker, 2 CPUs and 3 GiB RAM, TF2 ServerVersion 11138802. Both Bot Surge modes are included in every row. The production robot limiter remains enabled; robot caps, population files and mod code were not changed.

| Trial | Speed | Seed | Passed / planned | Elapsed | Native BLU flag captures |
| --- | ---: | ---: | ---: | ---: | ---: |
| breaking-walls-full-20x-seed10 | 20× | 10 | 12 / 12 | 173.6s | 12 |
| breaking-walls-full-50x-seed11 | 50× | 11 | 12 / 12 | 148.8s | 12 |
| computer-crash-20x-seed4 | 20× | 4 | 2 / 2 | 20.1s | 2 |
| denominator-20x-seed4 | 20× | 4 | 2 / 2 | 35.2s | 2 |
| shogun-negative-20x-seed4 | 20× | 4 | 0 / 2 | 70.3s | 0 |
| decoy-control-50x-seed3 | 50× | 3 | 2 / 2 | 15.1s | 0 |

Timings cover each controller trial and evidence collection, excluding initial server preparation. Different seeds were used; these timings are not a controlled speed benchmark.

Every accepted reverse wave requires a native flag capture by a BLU participant, a native BLU round win, and `mvm_wave_complete`. The native RED round win is a loss, even if reverse bookkeeping suppresses `mvm_wave_failed`. Positive JSONL rows contain capture/visit/pickup evidence; server logs record actors and native winning teams.

For each final trial, staged and installed probe SHA-256 values are verified before and after execution. The running status must contain the new `configsready` field. These checks detect the old-probe restoration encountered during development.

Probe SHA-256: `b55f2eca1a46083fcc32b4ffbcb1263776a0635626b5570bf980013dbef6058d`.

Controller SHA-256: `eed300453426840f541d852ec58dff88b81878d8a80378dff828e66bdbcd1905`.

## Findings and scope

- **Human bomb delivery:** the fake BLU player can use real pickup/deployment logic on Breaking Walls and Computer Crash. It only enters enabled objective volumes; it never fires winning relays or capture outputs.
- **Allied carriers:** Denominator disables human deployment. The simulator leaves its bomb to authored allied BLU carriers and preserves allied robots/tanks rather than changing that permission.
- **Payload/gate gap:** Shogun Slaughter has a payload train and four control points. The bomb-only simulator cannot progress those objectives. Its native losses are retained as unsupported simulator coverage, not evidence that the authored mission is broken.
- **Readiness and transitions:** completed reverse missions schedule native mission/map transitions. The controller waits for the actual transition and map configs before loading another mission, preserving the transition rather than disabling it.
- **Normal regression:** the final Decoy control uses ordinary defaults. Reverse remains opt-in; a default full plan is still 2,342 planned cases plus 112 unsupported reverse cases.

The reverse-only experimental plan contains 112 cases across eleven missions with navigation. This investigation does not validate every one of those cases, native Ubuntu, or Windows. Teleporting a fake player to an objective does not validate walking routes, survival, gate reachability, or human movement.

Valve's live version was checked at `2026-10-24T19:12:06.714402+00:00` and matched the tested `11138802` build.

## Preserved controls

The initial kill-only Breaking Walls wave-1 baseline lost when its authored timer expired. Early positive 20× results and a fresh-server 50× control motivated the prototype, but are diagnostic evidence rather than final-candidate validation.

The production entrypoint periodically restores staged addons. Hot-loading only the installed probe allowed an older staged binary to replace it in early trials. Those trials, their apparent passes, failed transitions and debug snapshots remain saved; none replace or inflate the final table. The final run pins the same candidate in both locations throughout.

One initial startup fixture omitted the required 32-player setting. That startup log is retained separately; the fixture was corrected before the baseline.

## Evidence

Raw local evidence: `/mnt/b/tf2-archipelago/docs/audits/reverse-probe-20261024/`. The `staged-confirmation/` directory contains final plans, first-pass JSONL, reports, before/after settings, objective snapshots, probe hashes and complete server logs. Earlier phase directories retain the negative and inconclusive controls. No experimental results were published to the daily badges.

`objective-inventory.json` records declarations and entity mentions from the twelve reverse population files, including the excluded no-navigation mission. Declarations may be inherited; runtime readbacks and native events are authoritative.

Go waveprobe tests/vet, 27 stability harness tests, 20 probe report/retry tests and the pinned SourcePawn compiler pass. Payload/control-point simulation and broader catalog/platform validation remain outside this prototype.
