#!/usr/bin/env python3
"""Summarize the unattended wave sweep without hiding unrun waves."""
import collections
import json
import pathlib
import re
import sys


def rows(path):
    with path.open(encoding="utf-8") as stream:
        for line in stream:
            if line.strip():
                yield json.loads(line)


def evidence_rank(row):
    if row["state"] == "passed":
        return 4
    if row["state"] == "failed":
        error = row.get("error", "")
        if "remained active for" in error:
            return 3
        if "did not complete within" in error:
            match = re.search(r"Elapsed:([0-9.]+)", error)
            elapsed = row.get("game_seconds", 0) or (float(match.group(1)) if match else 0)
            if elapsed >= 900:
                return 3
        if "game or probe failed wave" in error:
            return 2
    return 1


def main(run_dir):
    plan = {(r["mission"], r["mode"], r["wave"]): r
            for r in rows(run_dir / "plan.jsonl")}
    observed = {}
    load_errors = {}
    shards = sorted(run_dir.glob("shard-*.jsonl"))
    result_files = shards + sorted(run_dir.glob("retest-*.jsonl"))
    for shard in result_files:
        for row in rows(shard):
            key = (row["mission"], row["mode"], row["wave"])
            if row["state"] == "load_failed":
                load_errors[key[:2]] = row
            else:
                if key not in observed or evidence_rank(row) >= evidence_rank(observed[key]):
                    observed[key] = row

    def outcome(key):
        if plan[key]["state"] == "unsupported_reverse":
            return "reverse objective"
        if key in observed:
            row = observed[key]
            if (row["state"] == "inconclusive"
                    and "retest runner produced no wave result" in row.get("error", "")
                    and (key[:2] in load_errors or "did not load" in row["error"]
                         or "population file" in row["error"])):
                return "load blocked"
            if row["state"] == "failed" and "game or probe failed wave" in row.get("error", ""):
                return "wave lost"
            if row["state"] == "failed" and "remained active for" in row.get("error", ""):
                return ("no enemies observed" if not row.get("bots") and not row.get("tanks")
                        else "active at limit")
            if row["state"] == "failed" and "did not complete within" in row.get("error", ""):
                elapsed = row.get("game_seconds", 0)
                if not elapsed:
                    match = re.search(r"Elapsed:([0-9.]+)", row["error"])
                    elapsed = float(match.group(1)) if match else 0
                if elapsed < 900:
                    return "inconclusive"
                return ("no enemies observed" if not row.get("bots") and not row.get("tanks")
                        else "active at limit")
            return row["state"]
        if key[:2] in load_errors:
            return "load blocked"
        return "not run"

    counts = collections.Counter(outcome(key) for key in plan)
    elapsed = sum(row.get("wall_seconds", 0) for row in observed.values())
    missions = {key[0] for key in plan}
    print("# MvM wave progression sweep")
    print()
    config = run_dir / "config.txt"
    if config.exists():
        settings = dict(line.split("=", 1) for line in config.read_text().splitlines()
                        if "=" in line)
        print(f"Started {settings.get('started_utc', 'unknown')} with "
              f"{settings.get('shards', '?')} isolated servers at "
              f"{settings.get('speed', '?')}× game speed, seed "
              f"{settings.get('seed', '1')}; "
              f"{settings.get('wave_timeout', '?')} wall-time limit per wave.")
        print()
    print(f"{len(missions)} missions with usable navigation; {len(plan)} wave/mode cases; "
          f"{len(shards)} server shards. The probe attempts to defeat each "
          "observed BLU bot and tank 15–25 game seconds after first sighting. "
          "The game’s wave-complete event is required for a pass.")
    print()
    print("Catalog missions marked `no_nav` are excluded because the server "
          "cannot run them. Reverse MvM cases are counted separately below.")
    print()
    print("| Result | Waves |")
    print("| --- | ---: |")
    for state in ("passed", "active at limit", "no enemies observed", "failed", "inconclusive", "wave lost", "load blocked", "not run", "reverse objective"):
        print(f"| {state} | {counts[state]} |")
    print()
    print("Active at limit means the wave was still running after 900 game "
          "seconds with at least one recorded defeat; it is a review "
          "candidate, not proof of a deadlock. No enemies observed means "
          "the wave ran that long with no recorded bot or tank defeat. "
          "Inconclusive means the shorter wall-time screening limit expired. "
          "Wave lost means the game ended the test run before completion; "
          "it does not show a progression deadlock.")
    print()
    print("| Mode | Passed | Active at limit | No enemies observed | Failed | Inconclusive | Wave lost | Load blocked | Not run | Reverse objective |")
    print("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
    for mode in sorted({key[1] for key in plan}):
        by_mode = collections.Counter()
        for key in plan:
            if key[1] != mode:
                continue
            by_mode[outcome(key)] += 1
        print(f"| {mode} | {by_mode['passed']} | {by_mode['active at limit']} | "
              f"{by_mode['no enemies observed']} | "
              f"{by_mode['failed']} | {by_mode['inconclusive']} | "
              f"{by_mode['wave lost']} | "
              f"{by_mode['load blocked']} | {by_mode['not run']} | "
              f"{by_mode['reverse objective']} |")
    print(f"| **Total** | **{counts['passed']}** | **{counts['active at limit']}** | "
          f"**{counts['no enemies observed']}** | **{counts['failed']}** | "
          f"**{counts['inconclusive']}** | **{counts['wave lost']}** | "
          f"**{counts['load blocked']}** | **{counts['not run']}** | "
          f"**{counts['reverse objective']}** |")
    print()
    review = collections.defaultdict(collections.Counter)
    for mission, mode, wave in plan:
        state = outcome((mission, mode, wave))
        if state in ("active at limit", "no enemies observed", "inconclusive", "wave lost", "load blocked"):
            review[mission][f"{mode}:{state}"] += 1
    if review:
        print("## All missions with unresolved cases")
        print()
        print("Every mission with an unresolved standard wave case is listed "
              "here. Counts are wave cases; full per-wave details follow below.")
        print()
        print("| Mission | Normal active | Surge active | No enemies | Inconclusive | Wave lost | Load blocked |")
        print("| --- | ---: | ---: | ---: | ---: | ---: | ---: |")
        ranked = sorted(review, key=lambda mission: (
            -review[mission]["normal:active at limit"],
            -review[mission]["surge:active at limit"],
            -sum(count for kind, count in review[mission].items()
                 if kind.endswith(":load blocked")), mission))
        column_totals = collections.Counter()
        for mission in ranked:
            current = review[mission]
            idle = current["normal:no enemies observed"] + current["surge:no enemies observed"]
            inconclusive = current["normal:inconclusive"] + current["surge:inconclusive"]
            lost = current["normal:wave lost"] + current["surge:wave lost"]
            blocked = current["normal:load blocked"] + current["surge:load blocked"]
            print(f"| {mission} | {current['normal:active at limit']} | "
                  f"{current['surge:active at limit']} | {idle} | "
                  f"{inconclusive} | {lost} | {blocked} |")
            column_totals['normal'] += current['normal:active at limit']
            column_totals['surge'] += current['surge:active at limit']
            column_totals['idle'] += idle
            column_totals['inconclusive'] += inconclusive
            column_totals['lost'] += lost
            column_totals['blocked'] += blocked
        print(f"| **Total ({len(ranked)} missions)** | **{column_totals['normal']}** | "
              f"**{column_totals['surge']}** | **{column_totals['idle']}** | "
              f"**{column_totals['inconclusive']}** | **{column_totals['lost']}** | "
              f"**{column_totals['blocked']}** |")
        print()
    clean = sorted(missions - set(review) -
                   {key[0] for key in plan if plan[key]["state"] == "unsupported_reverse"})
    if clean:
        print("## Missions with every tested wave passed")
        print()
        for mission in clean:
            print(f"- {mission}")
        print()
    print(f"Total measured wave time across shards: {elapsed / 3600:.2f} hours.")
    print()
    problems = [row for row in observed.values() if row["state"] != "passed"]
    problems += [row for pair, row in load_errors.items()
                 if any(key[:2] == pair and outcome(key) == "load blocked" for key in plan)]
    if problems:
        print("## Problems")
        print()
        print("| Mission | Mode | Wave | Result | Detail |")
        print("| --- | --- | ---: | --- | --- |")
        for row in sorted(problems, key=lambda r: (r["mission"], r["mode"], r["wave"])):
            detail = row.get("error", "").replace("|", "\\|").replace("\n", " ")
            key = (row["mission"], row["mode"], row["wave"])
            state = outcome(key) if key in plan else row["state"]
            print(f"| {row['mission']} | {row['mode']} | {row['wave']} | "
                  f"{state} | {detail[:400]} |")
        print()
    empty = [row for row in observed.values() if row["state"] == "passed"
             and row.get("bots", 0) == 0 and row.get("tanks", 0) == 0]
    if empty:
        print("## Completed without a recorded kill")
        print()
        print("The game reported completion, but the probe recorded no bot or "
              "tank defeat. These cases do not validate the 15–25 second "
              "kill timing and need a separate inspection if enemies were "
              "expected.")
        print()
        for row in sorted(empty, key=lambda r: (r["mission"], r["mode"], r["wave"]))[:50]:
            print(f"- {row['mission']}, {row['mode']}, wave {row['wave']}")
        if len(empty) > 50:
            print(f"- …and {len(empty) - 50} more")
        print()
    if counts["not run"]:
        missing = sorted(key for key in plan if outcome(key) == "not run")
        print("## Unrun waves")
        print()
        for mission, mode, wave in missing[:100]:
            print(f"- {mission}, {mode}, wave {wave}")
        if len(missing) > 100:
            print(f"- …and {len(missing) - 100} more (see plan.jsonl)")
        print()
    reverse = sorted({key[0] for key in plan if outcome(key) == "reverse objective"})
    if reverse:
        print("## Reverse MvM requires an objective simulator")
        print()
        print("These missions force the player side to BLU. Killing RED robots "
              "does not complete the objective, so their waves were not run "
              "by this kill-only probe.")
        print()
        for mission in reverse:
            print(f"- {mission}")
        print()
    print("## Raw results")
    print()
    for shard in result_files:
        print(f"- [{shard.name}]({shard.name})")
    if (run_dir / "INCIDENTS.md").exists():
        print("- [Isolated server incidents](INCIDENTS.md)")
    print()
    print("A pass shows this timing and configuration completed once. It does "
          "not prove every authored spawn appeared, and it does not model "
          "player combat or every random timing seed. A wall-time expiry before "
          "900 game seconds is inconclusive and needs a slower retest.")


if __name__ == "__main__":
    main(pathlib.Path(sys.argv[1]))
