#!/usr/bin/env python3
"""Combine independent environment evidence into a dated repository dashboard."""
import argparse
import datetime as dt
import importlib.util
import json
from pathlib import Path
import shutil
import subprocess
import sys
import zipfile
import zlib

ENVIRONMENTS = {"ubuntu24": "Ubuntu 24.04", "docker": "Docker", "windows": "Windows"}


def read_json(path):
    return json.loads(path.read_text(encoding="utf-8")) if path.exists() else {}


def case_key(row):
    return row["mission"], row["mode"], row["wave"]


def records(path):
    return [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines() if line.strip()]


def combine_shards(source, plan_path, shards, commit):
    """Rebuild each environment from the full canonical plan, never just returned shards."""
    for environment in ENVIRONMENTS:
        root = source / environment
        root.mkdir(parents=True, exist_ok=True)
        meta = {"environment": environment, "commit": commit, "startup": True,
                "fresh": True, "completed": True, "shards": []}
        errors = []
        if not plan_path.exists():
            errors.append("build did not produce the canonical wave plan")
            meta.update(startup=False, fresh=False, completed=False)
            plan = []
        else:
            shutil.copy2(plan_path, root / "plan.jsonl")
            plan = records(plan_path)
        for shard in range(shards):
            path = root / f"part-{shard}"
            evidence = read_json(path / "environment.json")
            meta["shards"].append(evidence or {"shard": shard, "error": "no result"})
            valid = evidence.get("commit") == commit and evidence.get("shard") == shard and evidence.get("shards") == shards
            if not valid:
                errors.append(f"shard {shard}: missing result or incorrect source/partition")
            if evidence.get("error"):
                errors.append(f"shard {shard}: {evidence['error']}")
            meta["startup"] &= valid and evidence.get("startup") is True
            meta["fresh"] &= valid and evidence.get("fresh") is True
            meta["completed"] &= valid and evidence.get("completed") is True
            if not valid:
                continue
            expected = {case_key(row) for row in plan if zlib.crc32(row["mission"].encode()) % shards == shard}
            actual = {case_key(row) for row in records(path / "plan.jsonl")} if (path / "plan.jsonl").exists() else set()
            if actual != expected:
                errors.append(f"shard {shard}: wave plan differs from its canonical partition")
                meta.update(fresh=False, completed=False)
                continue
            # Each hosted shard uses one local retry worker. Prefix its output
            # when merging to preserve phase ordering in the existing report.
            for phase in ("shard", "screen", "retest"):
                destination = root / f"{phase}-{shard}.jsonl"
                with destination.open("w", encoding="utf-8") as stream:
                    for file in sorted(path.glob(f"{phase}-*.jsonl")):
                        stream.write(file.read_text(encoding="utf-8"))
        for field in ("installed_version", "running_version", "host"):
            values = {part.get(field) for part in meta["shards"] if part.get(field)}
            meta[field] = next(iter(values)) if len(values) == 1 else "mixed or unknown"
            if field.endswith("version") and len(values) != 1:
                meta["fresh"] = False
        starts = [part["started_utc"] for part in meta["shards"] if part.get("started_utc")]
        finishes = [part["finished_utc"] for part in meta["shards"] if part.get("finished_utc")]
        meta["started_utc"] = min(starts, default="unknown")
        meta["finished_utc"] = max(finishes, default="unknown")
        if errors:
            meta["error"] = "; ".join(errors)
        (root / "environment.json").write_text(json.dumps(meta, indent=2) + "\n", encoding="utf-8")
        if plan:
            with (root / "REPORT.md").open("w", encoding="utf-8") as stream:
                subprocess.run([sys.executable, str(Path(__file__).with_name("waveprobe-report.py")), str(root)],
                               stdout=stream, check=True)


def evaluate(meta, summary, commit, latest=None):
    eligible = summary.get("eligible_cases", 0)
    passed = summary.get("passed_cases", 0)
    tested = summary.get("tested_cases", 0)
    valid = (meta.get("commit") == commit and meta.get("startup") is True
             and meta.get("fresh") is True and eligible > 0
             and 0 <= passed <= tested <= eligible
             and (latest is None or meta.get("installed_version") == latest))
    # Missing waves remain in the denominator, including when the runner crashed.
    score = round(100 * passed / eligible, 2) if valid else None
    if not meta:
        status = "no result"
    elif meta.get("commit") != commit:
        status = "wrong commit"
    elif not meta.get("startup"):
        status = "startup failed"
    elif not meta.get("fresh") or (latest is not None and meta.get("installed_version") != latest):
        status = "latest build unverified"
    elif not valid:
        status = "invalid or empty wave plan"
    elif not meta.get("completed") or tested < eligible:
        status = "incomplete"
    elif passed < eligible:
        status = "wave failures"
    else:
        status = "passed"
    return {"status": status, "score": score, "startup": bool(meta.get("startup")),
            "passed": passed, "tested": tested, "eligible": eligible,
            "unsupported_reverse": summary.get("outcomes", {}).get("reverse objective", 0),
            "by_mode": summary.get("by_mode", {}), "evidence": meta}


def main(source, output, commit, run_url, check_live=False):
    output.mkdir(parents=True, exist_ok=True)
    now = dt.datetime.now(dt.timezone.utc)
    latest = None
    if check_live:
        spec = importlib.util.spec_from_file_location("stability_run", Path(__file__).with_name("stability-run.py"))
        runner = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(runner)
        try:
            latest = runner.live_version()
        except (OSError, ValueError, KeyError, TypeError):
            latest = "unverified"
    results = {key: evaluate(read_json(source / key / "environment.json"),
                             read_json(source / key / "SUMMARY.json"), commit, latest)
               for key in ENVIRONMENTS}
    dashboard = {"measured_utc": now.isoformat(), "commit": commit, "run_url": run_url,
                 "stale_after_hours": 36, "latest_at_report": latest, "environments": results}
    (output / "stability.json").write_text(json.dumps(dashboard, indent=2) + "\n")
    lines = ["# TF2-AP daily stability", "",
             f"Updated **{now:%Y-%m-%d %H:%M UTC}** · commit `{commit}` · [Actions run]({run_url})", "",
             "Results more than 36 hours old are stale. Check the date before relying on a badge.", "",
             "| Environment | Startup | TF2 ServerVersion | Stability | Passed / planned | Tested | Result |",
             "| --- | --- | --- | ---: | ---: | ---: | --- |"]
    for key, label in ENVIRONMENTS.items():
        result = results[key]
        score = f"{result['score']:.2f}%" if result["score"] is not None else "unverified"
        version = result["evidence"].get("installed_version", "unknown")
        lines.append(f"| {label} | {'pass' if result['startup'] else 'unverified'} | {version} | "
                     f"{score} | {result['passed']} / {result['eligible']} | {result['tested']} | {result['status']} |")
        badge = {"schemaVersion": 1, "label": label,
                 "message": f"{score} · {result['status']} · {now:%Y-%m-%d}",
                 "color": "brightgreen" if result["status"] == "passed" else "red",
                 "cacheSeconds": 300}
        (output / f"{key}.json").write_text(json.dumps(badge) + "\n")
    lines += ["", "Score = passed / every eligible planned wave/mode case, with startup and live-build verification required. "
              "Unrun waves count against the score; failed startup or unverified freshness has no numeric score.", "",
              "Each mission runs every wave at 20× with Bot Surge off and on. Nonpasses are retried at a 900-second "
              "wall limit; the original 180-second results remain in the evidence. A pass is a completion witness "
              "from this automated probe, not a guarantee for every player strategy.", "",
              "Reverse MvM objectives are unsupported by this probe and excluded from the denominator. "
              "Missions without navigation are excluded by the catalog. Neither exclusion is a pass.", "",
              "Download **stability-evidence.zip** for all environment reports, raw results, startup evidence, "
              "TF2 manifests, and server/SourceMod logs. Previous daily runs retain their evidence as Actions artifacts.", ""]
    for key, label in ENVIRONMENTS.items():
        result = results[key]
        lines += [f"## {label}", "", f"Unsupported Reverse MvM cases: {result['unsupported_reverse']}.", ""]
        meta = result["evidence"]
        lines += [f"Started: {meta.get('started_utc', 'unknown')}. Finished: {meta.get('finished_utc', 'unknown')}.",
                  f"Host: `{meta.get('host', 'unknown')}`. Running engine: `{meta.get('running_version', 'unknown')}`.", ""]
        if result["evidence"].get("error"):
            # Escape report content so runner error text cannot inject Markdown.
            error = str(result["evidence"]["error"]).replace("`", "'").replace("\n", " ")
            lines += [f"Runner error: `{error}`", ""]
        lines += ["| Mode | Passed / planned | Tested |", "| --- | ---: | ---: |"]
        for mode, part in result["by_mode"].items():
            lines.append(f"| {mode} | {part['passed']} / {part['eligible']} | {part['tested']} |")
        lines.append("")
    report = "\n".join(lines)
    (output / "REPORT.md").write_text(report, encoding="utf-8")
    with zipfile.ZipFile(output / "stability-evidence.zip", "w", zipfile.ZIP_DEFLATED) as archive:
        archive.write(output / "REPORT.md", "REPORT.md")
        archive.write(output / "stability.json", "stability.json")
        for path in sorted(source.rglob("*")):
            if path.is_file():
                archive.write(path, path.relative_to(source))
    return all(value["status"] == "passed" for value in results.values())


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    parser.add_argument("output", type=Path)
    parser.add_argument("--commit", required=True)
    parser.add_argument("--run-url", required=True)
    parser.add_argument("--check-live", action="store_true", help="verify Valve's live build again at publication")
    parser.add_argument("--shards", type=int, default=1)
    parser.add_argument("--plan", type=Path, help="canonical full plan emitted by the build job")
    args = parser.parse_args()
    if args.shards > 1:
        if args.plan is None:
            parser.error("--plan is required with --shards")
        combine_shards(args.source, args.plan, args.shards, args.commit)
    raise SystemExit(0 if main(args.source, args.output, args.commit, args.run_url, args.check_live) else 1)
