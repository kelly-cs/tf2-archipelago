#!/usr/bin/env python3
"""Summarize a waveprobe sweep with a comparable score and per-case evidence."""

import collections
import json
import pathlib
import sys

RESULTS = ("passed", "wave failed", "wave 0", "no enemies spawned",
           "wave timed out", "probe error", "load blocked", "inconclusive",
           "not run", "reverse objective")
MODES = {"normal": "Bot Surge off", "surge": "Bot Surge on"}


def rows(path):
    with path.open(encoding="utf-8") as stream:
        for line in stream:
            if line.strip():
                yield json.loads(line)


def classify(row, prior_load_failure=None):
    if row.get("retest_no_wave_result") and prior_load_failure:
        return classify(prior_load_failure)
    value = row.get("outcome") or row["state"]
    # Old runs recorded kill attempts, not spawns. Do not promote their
    # zero-kill results to the new "No enemies spawned" finding.
    return {"wave lost": "wave failed", "active at limit": "inconclusive",
            "no enemies observed": "inconclusive", "load_failed": "load blocked",
            "failed": "probe error"}.get(value, value)


def pct(n, d):
    return f"{100 * n / d:.1f}%" if d else "n/a"


def progress(row):
    value = row.get("progress_percent", -1)
    return f"{value:.0f}%" if value >= 0 else "unknown"


def safe(value):
    return str(value).replace("|", "\\|").replace("\n", " ")


def samples_to_show(samples, limit=12):
    if len(samples) <= limit:
        return samples
    indices = {round(i * (len(samples) - 1) / (limit - 1)) for i in range(limit)}
    return [samples[i] for i in sorted(indices)]


def load_results(run_dir):
    plan = {(r["mission"], r["mode"], r["wave"]): r for r in rows(run_dir / "plan.jsonl")}
    observed, load_errors = {}, {}
    files = sorted(run_dir.glob("shard-*.jsonl")) + sorted(run_dir.glob("retest-*.jsonl"))
    for path in files:
        for row in rows(path):
            key = row["mission"], row["mode"], row["wave"]
            if row["state"] == "load_failed":
                load_errors[key[:2]] = row
            elif key not in observed or classify(row) == "passed" or classify(observed[key]) != "passed":
                observed[key] = row

    def outcome(key):
        if plan[key]["state"] == "unsupported_reverse":
            return "reverse objective"
        if key in observed:
            return classify(observed[key], load_errors.get(key[:2]))
        if key[:2] in load_errors:
            return classify(load_errors[key[:2]])
        return "not run"

    return plan, observed, load_errors, files, outcome


def main(run_dir):
    plan, observed, load_errors, files, outcome = load_results(run_dir)
    outcomes = {key: outcome(key) for key in plan}
    counts = collections.Counter(outcomes.values())
    eligible = sum(row["state"] != "unsupported_reverse" for row in plan.values())
    tested = eligible - counts["not run"]
    passed = counts["passed"]
    config = run_dir / "config.txt"
    settings = dict(line.split("=", 1) for line in config.read_text().splitlines()
                    if "=" in line) if config.exists() else {}
    modes = [mode for mode in MODES if any(key[1] == mode for key in plan)]
    per_mode = {}
    for mode in modes:
        keys = [key for key in plan if key[1] == mode and plan[key]["state"] != "unsupported_reverse"]
        mode_tested = sum(outcomes[key] != "not run" for key in keys)
        mode_passed = sum(outcomes[key] == "passed" for key in keys)
        per_mode[mode] = {"eligible": len(keys), "tested": mode_tested,
                          "passed": mode_passed,
                          "stability_percent": round(100 * mode_passed / mode_tested, 2) if mode_tested else None}
    summary = {"planned_cases": len(plan), "eligible_cases": eligible,
               "tested_cases": tested, "passed_cases": passed,
               "stability_percent": round(100 * passed / tested, 2) if tested else None,
               "verified_coverage_percent": round(100 * passed / eligible, 2) if eligible else None,
               "outcomes": {name: counts[name] for name in RESULTS}, "by_mode": per_mode}
    (run_dir / "SUMMARY.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")

    print("# MvM wave smoke probe\n")
    print(f"## Stability score: **{pct(passed, tested)}** ({passed}/{tested} tested wave/mode cases passed)\n")
    print(f"**Verified coverage:** {pct(passed, eligible)} ({passed}/{eligible} eligible cases); "
          f"**tested:** {tested}/{eligible}; **missions:** {len({key[0] for key in plan})}. "
          "A pass requires the game's wave-complete event and an observed enemy spawn.\n")
    print("The stability score uses tested cases; verified coverage uses the full eligible plan, "
          "so unrun waves cannot silently improve it. Reverse objectives are excluded.\n")
    print(f"Started {settings.get('started_utc', 'unknown')} · {settings.get('shards', '?')} isolated servers · "
          f"requested speed {settings.get('speed', '?')}× · {settings.get('wave_timeout', '?')} "
          f"**real-time** limit after wave start · seed {settings.get('seed', '1')}.\n")
    if settings.get("source_commit") or settings.get("main_commit"):
        print(f"Source commit `{settings.get('source_commit', 'unknown')}` · "
              f"upstream main `{settings.get('main_commit', 'unknown')}`.\n")
    print("| Result | Bot Surge off | Bot Surge on | Total |\n| --- | ---: | ---: | ---: |")
    for name in RESULTS:
        normal = sum(value == name for key, value in outcomes.items() if key[1] == "normal")
        surge = sum(value == name for key, value in outcomes.items() if key[1] == "surge")
        print(f"| {name.title()} | {normal} | {surge} | **{counts[name]}** |")
    print()
    print("**Wave failed** is the game's loss event; **Wave 0** means the population manager did "
          "not initialize a wave; **No enemies spawned** means no BLU bot or tank was observed "
          "in a standard wave. **Wave timed out** means no completion after the real-time limit "
          "despite observed enemies. Probe errors and load blocks are separate.\n")
    print("| Mode | Stability | Passed | Tested | Eligible |\n| --- | ---: | ---: | ---: | ---: |")
    for mode in modes:
        part = per_mode[mode]
        print(f"| {MODES[mode]} | {pct(part['passed'], part['tested'])} | "
              f"{part['passed']} | {part['tested']} | {part['eligible']} |")
    print()

    issue_names = set(RESULTS[1:8])
    issues = sorted((key for key, value in outcomes.items() if value in issue_names),
                    key=lambda key: (plan[key]["map"], key[0], key[2], key[1]))
    if issues:
        print("## Waves to inspect\n")
        print("Each row is one wave in one mode. Progress uses the game's non-support enemy "
              "counter relative to its highest observed value; `unknown` means no baseline.\n")
        print("| Map | Mission | Mode | Wave | Result | Spawned | Auto killed | Alive | Progress | Real / game s |")
        print("| --- | --- | --- | ---: | --- | ---: | ---: | ---: | ---: | ---: |")
        for key in issues:
            row = observed.get(key) or load_errors.get(key[:2], {})
            spawns = row.get("bot_spawns", 0) + row.get("tank_spawns", 0)
            kills = row.get("bots", 0) + row.get("tanks", 0)
            print(f"| {safe(plan[key]['map'])} | {safe(key[0])} | {MODES.get(key[1], key[1])} | "
                  f"{key[2]} | {outcomes[key].title()} | {spawns} | {kills} | "
                  f"{row.get('alive_at_end', 0)} | {progress(row)} | "
                  f"{row.get('wall_seconds', 0):.0f} / {row.get('game_seconds', 0):.0f} |")
        print("\n### Timelines\n")
        for key in issues:
            row = observed.get(key) or load_errors.get(key[:2], {})
            samples = row.get("timeline", [])
            print(f"<details><summary>{safe(key[0])} · {MODES.get(key[1], key[1])} · "
                  f"wave {key[2]} · {outcomes[key].title()}</summary>\n")
            if samples:
                print("| Real s | Game s | Spawned | Auto killed | Alive | Game remaining | Progress |")
                print("| ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
                for point in samples_to_show(samples):
                    print(f"| {point['wall_seconds']:.0f} | {point['game_seconds']:.0f} | "
                          f"{point['spawned']} | {point['killed']} | {point['alive']} | "
                          f"{point['remaining']} | {progress(point)} |")
            else:
                print("No wave timeline: the mission did not load or the probe did not start.")
            if row.get("error"):
                print(f"\nReason: {safe(row['error'])[:500]}")
            print("\n</details>\n")

    if counts["not run"]:
        print(f"## Not run\n\n{counts['not run']} eligible cases have no result; see `plan.jsonl`.\n")
    reverse = {key[0] for key, value in outcomes.items() if value == "reverse objective"}
    if reverse:
        print(f"## Reverse objectives\n\n{len(reverse)} missions need a separate BLU objective simulator "
              "and are excluded from the score.\n")
    print("## Files\n\n- [Machine-readable summary](SUMMARY.json)")
    for path in files:
        print(f"- [{path.name}]({path.name})")
    print("\nA pass is one completion witness at this seed and timing; it does not prove every "
          "authored spawn appeared or every timing will complete.")


if __name__ == "__main__":
    main(pathlib.Path(sys.argv[1]))
