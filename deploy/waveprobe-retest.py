#!/usr/bin/env python3
"""Retest all nonpassing waveprobe cases against a game-time limit."""
import concurrent.futures
import json
import pathlib
import queue
import subprocess
import sys
import threading
import zlib


def rows(path):
    with path.open(encoding="utf-8") as stream:
        for line in stream:
            if line.strip():
                yield json.loads(line)


def main(run_dir, binary, worker_ids, source_shards=None):
    settings = dict(line.split("=", 1) for line in
                    (run_dir / "config.txt").read_text().splitlines() if "=" in line)
    shards = int(settings["shards"])
    if not worker_ids or any(worker < 0 or worker >= shards for worker in worker_ids):
        raise ValueError(f"worker ids must be in [0, {shards})")
    if source_shards is None:
        source_shards = worker_ids
    if any(shard < 0 or shard >= shards for shard in source_shards):
        raise ValueError(f"source shard ids must be in [0, {shards})")
    plan = {(row["mission"], row["mode"], row["wave"]): row
            for row in rows(run_dir / "plan.jsonl")}
    latest = {}
    retested = set()
    for path in sorted(run_dir.glob("shard-*.jsonl")) + sorted(run_dir.glob("retest-*.jsonl")):
        for row in rows(path):
            if row["wave"]:
                key = row["mission"], row["mode"], row["wave"]
                if key not in latest or latest[key]["state"] != "passed":
                    latest[key] = row
                if path.name.startswith("retest-"):
                    retested.add(key)
    pending = queue.Queue()
    for key, planned in plan.items():
        if planned["state"] != "planned":
            continue
        if zlib.crc32(key[0].encode()) % shards not in source_shards:
            continue
        if key in retested and latest[key]["state"] != "inconclusive":
            continue
        if key not in latest or latest[key]["state"] != "passed":
            pending.put((key, planned["map"]))
    total = pending.qsize()
    print(f"Retesting {total} incomplete waves with {len(worker_ids)} isolated servers.", flush=True)
    lock = threading.Lock()
    finished = 0

    def worker(index):
        nonlocal finished
        port = 27035 + index
        path = run_dir / f"retest-{index}.jsonl"
        with path.open("a", encoding="utf-8") as stream:
            while True:
                try:
                    (mission, mode, wave), map_name = pending.get_nowait()
                except queue.Empty:
                    return
                command = [str(binary), "-rcon", f"127.0.0.1:{port}",
                           "-mission", mission, "-mode", mode,
                           "-start-wave", str(wave), "-end-wave", str(wave),
                           "-speed", "20", "-timeout", "5m",
                           "-game-timeout", "15m", "-load-timeout", "90s"]
                try:
                    result = subprocess.run(command, capture_output=True, text=True,
                                            timeout=420, check=False)
                    output = [json.loads(line) for line in result.stdout.splitlines()
                              if line.startswith("{")]
                except (subprocess.TimeoutExpired, json.JSONDecodeError) as error:
                    output = []
                    result = None
                    reason = str(error)
                else:
                    reason = result.stderr.strip()
                wave_rows = [row for row in output if row.get("wave") == wave]
                if wave_rows:
                    row = wave_rows[-1]
                else:
                    if output:
                        reason = output[-1].get("error", reason)
                    row = {"mission": mission, "map": map_name, "mode": mode,
                           "wave": wave, "seed": 1, "state": "inconclusive",
                           "error": f"retest runner produced no wave result: {reason[:300]}"}
                stream.write(json.dumps(row) + "\n")
                stream.flush()
                with lock:
                    finished += 1
                    if finished % 25 == 0 or finished == total:
                        print(f"Retested {finished}/{total} waves.", flush=True)
                pending.task_done()

    with concurrent.futures.ThreadPoolExecutor(max_workers=len(worker_ids)) as pool:
        list(pool.map(worker, worker_ids))


if __name__ == "__main__":
    main(pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2]),
         [int(value) for value in sys.argv[3].split(",")],
         [int(value) for value in sys.argv[4].split(",")]
         if len(sys.argv) > 4 else None)
