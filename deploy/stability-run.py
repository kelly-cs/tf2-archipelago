#!/usr/bin/env python3
"""Run the same full wave plan on a disposable native or production Docker bed."""
import argparse
import datetime as dt
import json
import os
from pathlib import Path
import platform
import re
import shutil
import signal
import socket
import subprocess
import sys
import time
import urllib.request
import uuid


def live_version():
    url = "https://api.steampowered.com/ISteamApps/UpToDateCheck/v1/?appid=440&version=1"
    for attempt in range(3):
        try:
            with urllib.request.urlopen(url, timeout=20) as response:
                value = json.load(response)["response"]
            version = str(value["required_version"])
            if value.get("success") and version.isdigit() and int(version) > 0:
                return version
            raise ValueError("Valve returned no live ServerVersion")
        except (OSError, ValueError, KeyError, TypeError):
            if attempt == 2:
                raise
            time.sleep(5)


def installed_version(game):
    match = re.search(r"^ServerVersion=(\d+)\s*$",
                      (game / "tf/steam.inf").read_text(), re.MULTILINE)
    if not match:
        raise ValueError("steam.inf has no ServerVersion")
    return match[1]


def running_version(game, evidence=""):
    # Prefer the live engine's RCON reply. On Linux the Breakpad startup line
    # is in stdout and can be absent from -condebug's console.log.
    current = re.findall(r"^\s*Server version:\s*(\d+)\s*$", evidence, re.MULTILINE | re.IGNORECASE)
    if current:
        return current[-1]
    # Older engine output: the startup line used by the debug bundle.
    console = (game / "tf/console.log").read_text(encoding="utf-8", errors="replace")
    versions = re.findall(r"Using Breakpad minidump system\. Version: (\d+)", console)
    if not versions:
        raise ValueError("engine did not report its running ServerVersion")
    return versions[-1]


def stop_process(process):
    if process.poll() is not None:
        return
    if os.name == "nt":
        subprocess.run(["taskkill", "/PID", str(process.pid), "/T", "/F"],
                       capture_output=True, check=False, timeout=30)
    else:
        # winbed handles Ctrl-C by cancelling RunServer, which also stops the
        # separate process group srcds_run owns. SIGTERM would bypass that
        # handler and leave its game server running after the bed exited.
        try:
            os.killpg(process.pid, signal.SIGINT)
        except ProcessLookupError:
            pass
        try:
            process.wait(timeout=15)
        except subprocess.TimeoutExpired:
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
    process.wait(timeout=30)


def startup_ready(evidence):
    return ("TF2 MvM Archipelago" in evidence and "TF2 MvM wave probe" in evidence
            and not re.search(r"<Failed>|<Error>|<Bad Load>|failed to load|Unknown command",
                              evidence, re.IGNORECASE))


def robot_limiter_enabled(reply):
    # Fail closed on an unknown cvar, changed output or a disabled production mod.
    return bool(re.search(r'^\s*"sig_mvm_robot_limit"\s*=\s*"1(?:\.0+)?"',
                          reply, re.MULTILINE))


def check_robot_limiter(rcon, output):
    reply = rcon("sig_mvm_robot_limit")
    output.write_text(reply, encoding="utf-8")
    if not robot_limiter_enabled(reply):
        raise RuntimeError("production robot limiter is disabled or unverified")


def crash_evidence(output):
    # srcds_run restarts a crashed engine. A later healthy RCON reply must not
    # erase a fatal error from startup or an earlier wave in this measurement.
    failures = []
    for name in ("server.log", "docker.log", "console.log"):
        path = output / name
        if path.exists():
            for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
                if re.search(r"Segmentation fault|FATAL ERROR|EXCEPTION_ACCESS_VIOLATION", line, re.IGNORECASE):
                    failures.append(f"{name}: {line}")
    return failures


def rcon_hosts(environment):
    hosts = ["127.0.0.1"]
    if environment != "docker":
        hosts += sorted({row[4][0] for row in socket.getaddrinfo(socket.gethostname(), None, socket.AF_INET)})
    return list(dict.fromkeys(hosts))


def main(opt):
    out = opt.output.resolve()
    out.mkdir(parents=True, exist_ok=True)
    if any(out.iterdir()):
        raise ValueError("use an empty --output; previous passes must not leak into this measurement")
    bed = opt.bed.resolve()
    root = opt.root.resolve()
    if root.exists():
        raise ValueError("use a fresh disposable --root; existing roots are refused")
    suffix = ".exe" if os.name == "nt" else ""
    binary = lambda name: str(bed / (name + suffix))
    game = root / "tf-dedicated"
    port = "27035"
    os.environ["SRCDS_RCONPW"] = os.environ["WAVEPROBE_RCONPW"]
    os.environ["SRCDS_PORT"] = port
    shard = getattr(opt, "shard", 0)
    shards = getattr(opt, "shards", 1)
    if shards < 1 or not 0 <= shard < shards:
        raise ValueError("shard must be in [0, shards)")
    speed = getattr(opt, "speed", 50)
    if not 1 <= speed <= 50:
        raise ValueError("speed must be between 1 and 50")
    partition = ["-shards", str(shards), "-shard", str(shard)]
    manifest = getattr(opt, "partition", None)
    if manifest:
        shutil.copy2(manifest, out / "global-partition.json")
        partition += ["-partition", str((out / "global-partition.json").resolve())]
    meta = {"environment": opt.environment, "shard": shard, "shards": shards, "speed": speed,
            "started_utc": dt.datetime.now(dt.timezone.utc).isoformat(),
            "commit": os.environ.get("GITHUB_SHA", "local"), "host": platform.platform(),
            "run_url": os.environ.get("STABILITY_RUN_URL", ""), "startup": False,
            "fresh": False, "completed": False, "robot_limiter": False, "crash_free": False}
    process = None
    container = "tf2ap-stability-" + uuid.uuid4().hex[:12]
    deadline = time.monotonic() + opt.hours * 3600

    def run(args, filename, limit=None, check=True):
        remaining = max(1, deadline - time.monotonic())
        with (out / filename).open("w", encoding="utf-8") as stream:
            with (out / (filename + ".err")).open("w", encoding="utf-8") as errors:
                command = subprocess.Popen(args, stdout=stream, stderr=errors,
                                           start_new_session=os.name != "nt")
                try:
                    status = command.wait(timeout=min(remaining, limit or remaining))
                except BaseException:
                    stop_process(command)
                    raise
                if check and status:
                    raise subprocess.CalledProcessError(status, args)
                return status

    def rcon(command):
        result = subprocess.run([binary("rcon"), command], capture_output=True,
                                text=True, timeout=15, check=True)
        return result.stdout

    try:
        shutil.copy2("deploy/env/versions.env", out / "pins.env")
        if opt.environment == "ubuntu24":
            os_release = Path("/etc/os-release").read_text()
            if not re.search(r'^ID=ubuntu$', os_release, re.MULTILINE) or 'VERSION_ID="24.04"' not in os_release:
                raise ValueError("ubuntu24 requires native Ubuntu 24.04")
        if opt.environment == "windows" and os.name != "nt":
            raise ValueError("Windows must run natively")
        run([binary("waveprobe"), "-plan", "-mode", "both", *partition], "plan.jsonl", 60)
        (out / "config.txt").write_text(
            f"shards=1\nspeed={speed}\nseed=1\nfirst_pass_timeout=180s\nwave_timeout=900s\n"
            f"base_port={port}\nstarted_utc={meta['started_utc']}\nsource_commit={meta['commit']}\n")
        meta["latest_before"] = live_version()
        run([binary("winbed"), "-root", str(root), "-probe", str(bed / "tf2_waveprobe.smx"),
             "-install"], "install.log", 7200)
        # The player launcher tolerates a failed update; the measurement cannot.
        meta["installed_version"] = installed_version(game)
        if meta["installed_version"] != meta["latest_before"]:
            raise ValueError("installed server is not Valve's current build")
        if opt.environment == "docker":
            # A fresh launcher-prepared tree supplies the same two community packs.
            # Production entrypoint, mods, SteamCMD update and server flags still run.
            run(["docker", "run", "--rm", "--user", "root", "--entrypoint", "chown",
                 "-v", f"{game}:/game", opt.image, "-R", "steam:steam", "/game"], "ownership.log", 300)
            args = ["docker", "run", "--rm", "--name", container,
                    "-p", f"127.0.0.1:{port}:27015/tcp",
                    "-v", f"{game}:/home/steam/tf-dedicated",
                    "-v", f"{bed / 'tf2_waveprobe.smx'}:/opt/tf2-archipelago/addons/sourcemod/plugins/tf2_waveprobe.smx:ro"]
            for value in ["SRCDS_REACH=lan", "SRCDS_LAN=1", "SRCDS_STARTMAP=mvm_decoy",
                          "SRCDS_MAXPLAYERS=32", "SRCDS_PORT=27015", "SRCDS_BOTS=0",
                          "SRCDS_MODS=sigsegv-mvm", "SRCDS_RCONPW", "TF2AP_TEST_MODE=0"]:
                args += ["-e", value]
            args += [opt.image]
        else:
            args = [binary("winbed"), "-root", str(root), "-port", port, "-serve"]
        with (out / "server.log").open("w", encoding="utf-8") as log:
            process = subprocess.Popen(args, stdout=log, stderr=log, start_new_session=os.name != "nt",
                                       creationflags=subprocess.CREATE_NEW_CONSOLE if os.name == "nt" else 0)
        hosts = rcon_hosts(opt.environment)
        host_index = 0
        ready_limit = min(deadline, time.monotonic() + 900)
        while time.monotonic() < ready_limit:
            if process.poll() is not None:
                raise RuntimeError("server exited before readiness")
            os.environ["SRCDS_RCON_HOST"] = hosts[host_index % len(hosts)]
            host_index += 1
            try:
                evidence = "\n".join(rcon(cmd) for cmd in
                                     ["version", "sm plugins list", "sm exts list", "sm_waveprobe_status"])
                (out / "startup.txt").write_text(evidence, encoding="utf-8")
                if startup_ready(evidence):
                    break
            except (subprocess.SubprocessError, OSError):
                pass
            time.sleep(5)
        else:
            raise RuntimeError("startup did not load the mod, extensions and probe within 15 minutes")
        address = os.environ["SRCDS_RCON_HOST"]
        meta["rcon_host"] = address
        with (out / "config.txt").open("a") as config:
            config.write(f"rcon_host={address}\n")
        meta["startup"] = True
        meta["installed_version"] = installed_version(game)
        meta["running_version"] = running_version(game, evidence)
        meta["latest_at_start"] = live_version()
        if not meta["installed_version"] == meta["running_version"] == meta["latest_at_start"]:
            raise ValueError("TF2 changed during setup; rerun against the new build")
        meta["fresh"] = True
        check_robot_limiter(rcon, out / "robot-limiter-before.txt")
        run([binary("waveprobe"), "-rcon", f"{address}:{port}", "-mission", "all",
                      "-mode", "both", "-speed", str(speed), "-seed", "1", "-timeout", "180s", *partition],
                     "shard-0.jsonl", check=False)
        run([sys.executable, "deploy/waveprobe-retest.py", str(out), binary("waveprobe"), "0"],
            "retest.log")
        final_health = "\n".join(rcon(cmd) for cmd in
                                  ["version", "sm plugins list", "sm exts list", "sm_waveprobe_status"])
        (out / "final-health.txt").write_text(final_health, encoding="utf-8")
        if not startup_ready(final_health):
            raise RuntimeError("server mods did not remain healthy through the sweep")
        check_robot_limiter(rcon, out / "robot-limiter-after.txt")
        meta["robot_limiter"] = True
        meta["latest_after"] = live_version()
        meta["fresh"] = meta["latest_after"] == meta["installed_version"]
        meta["completed"] = True
    except (Exception, KeyboardInterrupt) as error:
        meta["error"] = str(error)
        print(f"stability: {error}", file=sys.stderr)
    finally:
        if opt.environment == "docker":
            try:
                with (out / "docker.log").open("w") as logs:
                    subprocess.run(["docker", "logs", container], stdout=logs,
                                   stderr=subprocess.STDOUT, timeout=30, check=False)
                subprocess.run(["docker", "rm", "-f", container], capture_output=True, timeout=30, check=False)
            except (OSError, subprocess.SubprocessError) as error:
                meta["cleanup_error"] = str(error)
        if process:
            stop_process(process)
        for relative in ["tf/console.log", "tf/steam.inf", "steamapps/appmanifest_232250.acf"]:
            source = game / relative
            if source.exists():
                shutil.copy2(source, out / source.name)
        if (game / "tf/addons/sourcemod/logs").exists():
            shutil.copytree(game / "tf/addons/sourcemod/logs", out / "sourcemod-logs", dirs_exist_ok=True)
        crashes = crash_evidence(out)
        meta["crash_free"] = not crashes
        if crashes:
            meta["crashes"] = crashes
            meta["error"] = "; ".join(filter(None, [meta.get("error"), "engine crash recorded; automatic restart does not count as recovery"]))
        meta["finished_utc"] = dt.datetime.now(dt.timezone.utc).isoformat()
        (out / "environment.json").write_text(json.dumps(meta, indent=2) + "\n")
        if (out / "plan.jsonl").exists():
            with (out / "REPORT.md").open("w", encoding="utf-8") as report:
                subprocess.run([sys.executable, "deploy/waveprobe-report.py", str(out)],
                               stdout=report, check=True)
    summary = json.loads((out / "SUMMARY.json").read_text()) if (out / "SUMMARY.json").exists() else {}
    return 0 if (meta["startup"] and meta["fresh"] and meta["completed"] and meta["crash_free"]
                 and summary.get("eligible_cases", 0) > 0
                 and summary.get("passed_cases") == summary.get("eligible_cases")) else 1


if __name__ == "__main__":
    def interrupted(signum, frame):
        raise KeyboardInterrupt(f"interrupted by signal {signum}")

    signal.signal(signal.SIGTERM, interrupted)
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--environment", choices=["ubuntu24", "docker", "windows"], required=True)
    parser.add_argument("--bed", type=Path, required=True)
    parser.add_argument("--root", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--image", default="tf2ap-stability:current")
    parser.add_argument("--hours", type=float, default=20)
    parser.add_argument("--shards", type=int, default=1)
    parser.add_argument("--shard", type=int, default=0)
    parser.add_argument("--speed", type=int, default=50)
    parser.add_argument("--partition", type=Path)
    sys.exit(main(parser.parse_args()))
