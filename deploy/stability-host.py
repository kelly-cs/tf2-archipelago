#!/usr/bin/env python3
"""Reclaim optional SDKs on an ephemeral GitHub-hosted measurement runner."""
import os
import json
from pathlib import Path
import shutil
import subprocess
import sys


def main():
    if os.environ.get("STABILITY_RUNNER_ENVIRONMENT") != "github-hosted":
        raise RuntimeError("SDK cleanup is allowed only on an ephemeral GitHub-hosted runner")
    if os.name == "nt":
        # Keep Python, Node, MSVC and the VC runtimes the test actually uses.
        paths = [Path("C:/Android"), Path("C:/Program Files/Android"),
                 Path("C:/Program Files/dotnet"), Path("C:/Program Files/Java")]
        tools = Path(os.environ["RUNNER_TOOL_CACHE"])
        paths += [tools / name for name in ("CodeQL", "Go", "Java_Temurin-Hotspot_jdk", "Ruby", "PyPy")]
        for path in paths:
            if path.exists():
                shutil.rmtree(path, ignore_errors=True)
    else:
        subprocess.run(["sudo", "rm", "-rf", "/usr/share/dotnet", "/usr/local/lib/android",
                        "/opt/ghc", "/usr/share/swift", "/opt/hostedtoolcache/CodeQL"], check=True)
        subprocess.run(["sudo", "apt-get", "update", "-qq"], check=True)
        subprocess.run(["sudo", "apt-get", "install", "-y", "lib32gcc-s1", "lib32stdc++6"], check=True)
    root = Path(os.environ["RUNNER_TEMP"])
    free = shutil.disk_usage(root).free / 1024**3
    required = 45 if os.environ["STABILITY_ENV"] == "docker" else 35
    print(f"Game volume: {root.anchor}; free {free:.1f} GiB; required {required} GiB", flush=True)
    if free < required:
        raise RuntimeError(f"insufficient disk: {free:.1f} GiB available, {required} GiB required")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        output = Path("stability-result")
        output.mkdir(exist_ok=True)
        (output / "environment.json").write_text(json.dumps({
            "environment": os.environ.get("STABILITY_ENV"),
            "shard": int(os.environ.get("STABILITY_SHARD", "0")), "shards": 6,
            "commit": os.environ.get("GITHUB_SHA"), "startup": False,
            "fresh": False, "completed": False,
            "error": "host setup failed: " + str(error)}, indent=2) + "\n", encoding="utf-8")
        print(f"stability host setup: {error}", file=sys.stderr)
        raise SystemExit(1)
