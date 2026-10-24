#!/usr/bin/env python3
"""Single-server, counterbalanced timescale experiment; never uses a live volume."""
import collections
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import statistics
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent
REPO = Path('/tmp/tf2ap-stability-schedule')
CFG = json.loads((ROOT / 'run.json').read_text())
ENV = {**os.environ, 'WAVEPROBE_RCONPW': secrets.token_hex(24)}
ENV.update(SRCDS_RCONPW=ENV['WAVEPROBE_RCONPW'], SRCDS_PORT=str(CFG['port']))
COMPOSE = ['docker', 'compose', '-p', CFG['project'], '-f', str(ROOT / 'compose.yml')]
CONTAINER = CFG['project'] + '-srcds-1'
TRIALS = []


def now():
    return dt.datetime.now(dt.timezone.utc).isoformat()


def call(args, timeout=120, check=True):
    return subprocess.run([str(a) for a in args], env=ENV, capture_output=True, text=True,
                          timeout=timeout, check=check)


def rcon(command):
    return call([ROOT / 'rcon', command], 15).stdout


def rows(path):
    return [json.loads(line) for line in path.read_text().splitlines() if line.strip()]


def live_version():
    response = json.loads(call(['curl', '-fsSL', '--retry', '3', '--max-time', '20',
        'https://api.steampowered.com/ISteamApps/UpToDateCheck/v1/?appid=440&version=1']).stdout)['response']
    if not response.get('success'):
        raise RuntimeError('Valve did not provide the live build')
    return str(response['required_version'])


def execute_probe(command, output, errors, limit):
    started = time.monotonic()
    with output.open('w') as stream, errors.open('w') as stderr:
        proc = subprocess.Popen(command, env=ENV, stdout=stream, stderr=stderr, start_new_session=True)
        next_update = started + 30
        try:
            while proc.poll() is None:
                if time.monotonic() - started > limit:
                    raise TimeoutError('trial process exceeded its overall wall limit')
                if time.monotonic() >= next_update:
                    observed = rows(output)
                    print(f"  {output.parent.name}: {len(observed)} cases recorded, {time.monotonic()-started:.0f}s", flush=True)
                    next_update += 30
                time.sleep(2)
        finally:
            if proc.poll() is None:
                os.killpg(proc.pid, 2)
                proc.wait(timeout=20)
    return proc.returncode, time.monotonic() - started


def trial(speed, repeat):
    directory = ROOT / f'{speed:03d}x-repeat-{repeat}'
    directory.mkdir()
    print(f'Starting {speed}x repeat {repeat} at {now()}', flush=True)
    rcon('host_timescale 1')
    rcon('sm_waveprobe_reset')
    rcon('changelevel mvm_decoy')
    time.sleep(3)
    limit = time.monotonic() + 120
    while time.monotonic() < limit:
        try:
            if 'state=idle' in rcon('sm_waveprobe_status'):
                break
        except subprocess.SubprocessError:
            pass
        time.sleep(1)
    else:
        raise RuntimeError('fresh map did not become ready')
    before = rcon('sig_mvm_robot_limit') + rcon('tf_mvm_max_invaders')
    (directory / 'settings-before.txt').write_text(before)
    if not re.search(r'\"sig_mvm_robot_limit\" = \"1\"', before):
        raise RuntimeError('production limiter is disabled or unverified')
    command = [str(ROOT / 'waveprobe'), '-rcon', f"127.0.0.1:{CFG['port']}",
               '-mission', CFG['mission'], '-mode', 'both', '-seed', str(CFG['seed']),
               '-speed', str(speed), '-timeout', '180s', '-load-timeout', '90s']
    if speed > 20:
        command.append('-experimental-speed')
    plan = call([ROOT / 'waveprobe', '-plan', '-mission', CFG['mission'], '-mode', 'both']).stdout
    (directory / 'plan.jsonl').write_text(plan)
    (directory / 'config.txt').write_text(
        f"scope=Decoy Expert speed comparison; repeat {repeat}\nshards=1\nspeed={speed}\nseed={CFG['seed']}\n"
        f"started_utc={now()}\nfirst_pass_timeout=180s\nwave_timeout=900s\nsource_commit={CFG['source_commit']}\n")
    (directory / 'command.json').write_text(json.dumps(command) + '\n')
    trial_start = now()
    status, elapsed = execute_probe(command, directory / 'shard-0.jsonl', directory / 'shard-0.err', 3600)
    original = rows(directory / 'shard-0.jsonl')
    logs = call(['docker','logs','--since',trial_start,CONTAINER], check=False)
    (directory / 'server.log').write_text(logs.stdout+logs.stderr)
    settings = rcon('sig_mvm_robot_limit') + rcon('tf_mvm_max_invaders') + rcon('tf_allow_server_hibernation')
    (directory / 'settings-after.txt').write_text(settings)
    if not re.search(r'\"sig_mvm_robot_limit\" = \"1\"',settings):
        raise RuntimeError('production limiter changed or unverified')
    failed = [row for row in original if row.get('outcome', row['state']) != 'passed']
    for index, row in enumerate(failed[:1]):
        if row['wave'] <= 0:
            continue
        retry = command[:]
        retry[retry.index('-mode')+1] = row['mode']
        retry[retry.index('-timeout')+1] = '900s'
        retry += ['-start-wave', str(row['wave']), '-end-wave', str(row['wave'])]
        print(f"  retrying {speed}x {row['mode']} wave {row['wave']} with 900s limit", flush=True)
        execute_probe(retry, directory / f'retest-{index}.jsonl', directory / f'retest-{index}.err', 1200)
    with (directory / 'REPORT.md').open('w') as report:
        subprocess.run([sys.executable, str(REPO / 'deploy/waveprobe-report.py'), str(directory)], stdout=report, check=True)
    (directory / 'health.txt').write_text(rcon('host_timescale') + rcon('sm_waveprobe_debug'))
    (directory / 'stats.json').write_text(call(['docker', 'stats', '--no-stream', '--format', '{{json .}}', CONTAINER]).stdout)
    result = {'speed': speed, 'repeat': repeat, 'process_exit': status, 'elapsed_seconds': elapsed,
              'summary': json.loads((directory / 'SUMMARY.json').read_text()),
              'seed': CFG['seed'],
              'crash_free': not re.search(r'Segmentation fault|FATAL ERROR|EXCEPTION_ACCESS_VIOLATION', logs.stdout+logs.stderr, re.I),
              'first_pass_cases': len(original),
              'first_pass_outcomes': dict(collections.Counter(row.get('outcome', row['state']) for row in original)),
              'directory': directory.name}
    TRIALS.append(result)
    (ROOT / 'trials.json').write_text(json.dumps(TRIALS, indent=2) + '\n')
    print(f"Finished {speed}x repeat {repeat}: {result['first_pass_outcomes']}; {elapsed:.1f}s", flush=True)
    return result


def report():
    baseline = {}
    for item in TRIALS:
        if item['speed'] == 20:
            baseline[item['repeat']] = {(r['mode'], r['wave']): r for r in rows(ROOT / item['directory'] / 'shard-0.jsonl') if r['wave'] > 0}
    lines = ['# Decoy accelerated-wave comparison', '',
        f"Source `{CFG['source_commit']}` plus the recorded experimental opt-in patch. TF2 ServerVersion `{CFG.get('running_version', 'unknown')}`.", '',
        f"Mission `{CFG['mission']}` on `{CFG['map']}`; all 7 waves, normal and Bot Surge modes, seed 1. One isolated Linux Docker server, 4-CPU quota, 4 GiB RAM.", '',
        'Order: ascending speeds, then descending speeds. Each trial starts with a fresh map load. First-pass results and nonpass snapshots remain separate from 900-second diagnostic retries.', '',
        '| Speed | Repeat | First-pass passes | Final passes | Trial wall s | Wave wall s | Wave game s | Effective acceleration | Spawn-count differences vs 20x |',
        '| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |']
    comparison = []
    for item in TRIALS:
        records = [r for r in rows(ROOT / item['directory'] / 'shard-0.jsonl') if r['wave'] > 0]
        wall = sum(r.get('wall_seconds', 0) for r in records)
        game = sum(r.get('game_seconds', 0) for r in records)
        control = baseline.get(item['repeat'], baseline.get(1, {}))
        differences = []
        for row in records:
            key = row['mode'], row['wave']
            base = control.get(key)
            if base and any(row.get(field, 0) != base.get(field, 0) for field in ('bot_spawns', 'tank_spawns')):
                differences.append({'mode': row['mode'], 'wave': row['wave'],
                    'baseline': {f: base.get(f, 0) for f in ('bot_spawns', 'tank_spawns')},
                    'observed': {f: row.get(f, 0) for f in ('bot_spawns', 'tank_spawns')}})
        item['spawn_differences'] = differences
        item['effective_acceleration'] = game / wall if wall else None
        summary = item['summary']
        lines.append(f"| {item['speed']}x | {item['repeat']} | {item['first_pass_outcomes'].get('passed', 0)}/14 | {summary['passed_cases']}/14 | {item['elapsed_seconds']:.1f} | {wall:.1f} | {game:.1f} | {game/wall if wall else 0:.2f}x | {len(differences)} |")
        comparison.append(item)
    lines += ['', 'A completion witness on this one stock mission is not proof of reliability on other maps, authored timers, physics-dependent objectives, community scripts, or Windows. Spawn-count differences may occur even between repeated controls; inspect `COMPARISON.json` before attributing them to speed.', '',
              'Trial time includes probe setup and mission/mode loads, but excludes the initial game update and explicit fresh-map reset. Wave time also includes the runner’s wave-arm/restart waits. Effective acceleration is observed game seconds divided by observed wave wall seconds.', '']
    if CFG.get('error'):
        lines += ['Run error: ' + CFG['error'], '']
    (ROOT / 'REPORT.md').write_text('\n'.join(lines))
    (ROOT / 'COMPARISON.json').write_text(json.dumps(comparison, indent=2) + '\n')


def main():
    volume = CFG['project'] + '_game'
    CFG.update(started_utc=now(), source_commit=(ROOT / 'source-commit.txt').read_text().strip(),
               host=call(['uname', '-a']).stdout.strip(), image_id=call(['docker', 'image', 'inspect', '--format', '{{.Id}}', CFG['image']]).stdout.strip(),
               binary_sha256=hashlib.sha256((ROOT / 'waveprobe').read_bytes()).hexdigest())
    compose = f'''services:
  srcds:
    image: {CFG['image']}
    restart: "no"
    cpus: 4
    mem_limit: 4g
    network_mode: bridge
    ports:
      - "127.0.0.1:{CFG['port']}:27015/tcp"
    environment:
      SRCDS_HOSTNAME: TF2 isolated speed comparison
      SRCDS_RCONPW: ${{WAVEPROBE_RCONPW}}
      SRCDS_PORT: 27015
      SRCDS_MAXPLAYERS: 32
      SRCDS_STARTMAP: mvm_decoy
      SRCDS_LAN: 1
      SRCDS_REACH: lan
      SRCDS_BOTS: 0
      SRCDS_MODS: sigsegv-mvm
      TF2AP_TEST_MODE: 0
    volumes:
      - game:/home/steam/tf-dedicated
      - {ROOT}/tf2_waveprobe.smx:/opt/tf2-archipelago/addons/sourcemod/plugins/tf2_waveprobe.smx:ro
      - {ROOT}/tf2_archipelago.smx:/opt/tf2-archipelago/addons/sourcemod/plugins/tf2_archipelago.smx:ro
volumes:
  game:
    external: true
    name: {volume}
'''
    (ROOT / 'compose.yml').write_text(compose)
    if call(['docker', 'volume', 'inspect', volume], check=False).returncode == 0:
        raise RuntimeError('refusing to reuse an existing experiment volume')
    call(['docker', 'volume', 'create', volume])
    try:
        print('Copying the pristine test volume; source is mounted read-only.', flush=True)
        copied = call(['docker', 'run', '--rm', '-v', CFG['source_volume']+':/src:ro', '-v', volume+':/dst',
                       'debian:bookworm-slim', 'sh', '-c', 'cp -a /src/. /dst/'], 1200)
        (ROOT / 'copy.log').write_text(copied.stdout + copied.stderr)
        call(['docker','run','--rm','-v',volume+':/g','debian:bookworm-slim','sh','-c','mv /g/tf/addons /g/tf/addons.original; mv /g/tf/cfg /g/tf/cfg.original; mkdir /g/tf/cfg; cp /g/tf/cfg.original/*_default.txt /g/tf/cfg/; chown -R 1000:1000 /g/tf/cfg'],120)
        call(COMPOSE + ['up', '-d'])
        print('Waiting for SteamCMD update and isolated server readiness.', flush=True)
        deadline = time.monotonic() + 1200
        while time.monotonic() < deadline:
            try:
                evidence = ''.join(rcon(c) for c in ['version', 'sm plugins list', 'sm exts list', 'sm_waveprobe_status'])
                if ('TF2 MvM Archipelago' in evidence and 'TF2 MvM wave probe' in evidence and
                    not re.search(r'<Failed>|<Error>|Unknown command', evidence, re.I)):
                    (ROOT / 'startup.txt').write_text(evidence)
                    version = re.search(r'^Server version:\s*(\d+)', evidence, re.M)
                    CFG['running_version'] = version[1] if version else None
                    CFG['latest_before'] = live_version()
                    if CFG['running_version'] != CFG['latest_before']:
                        raise RuntimeError('running engine is not the latest public TF2 build')
                    break
            except (subprocess.SubprocessError, OSError):
                pass
            time.sleep(5)
        else:
            raise RuntimeError('server did not become ready in 20 minutes')
        def passed(item):
            return item['first_pass_outcomes'] == {'passed': 14} and item['first_pass_cases'] == 14 and item['crash_free']
        first_failure = None
        for speed in CFG['speeds']:
            result = trial(speed, 1)
            report()
            if not passed(result):
                first_failure = speed
                CFG['first_failure'] = speed
                break
        # Repeat in reverse to detect drift and confirm a failed speed.
        for speed in reversed([t['speed'] for t in TRIALS if t['repeat'] == 1]):
            trial(speed, 2)
            report()
        good = [t['speed'] for t in TRIALS if t['repeat'] == 1 and passed(t) and
                any(u['speed']==t['speed'] and u['repeat']==2 and passed(u) for u in TRIALS)]
        if not good: raise RuntimeError('no speed passed both full repeats')
        best = max(good)
        if first_failure and first_failure-best > 5:
            for speed in range(best+5, first_failure, 5):
                a = trial(speed, 1); report()
                b = trial(speed, 2); report()
                if passed(a) and passed(b): best = speed
                else: break
        # Require an additional complete mission at another kill-delay seed.
        CFG['seed'] = 2
        confirmation = trial(best, 3); report()
        CFG['fastest_confirmed'] = best if passed(confirmation) else None
        CFG['seed'] = 1
        trial(20, 3); report()
        CFG['latest_after'] = live_version()
        CFG['fresh'] = CFG['latest_before'] == CFG['latest_after'] == CFG['running_version']
    except Exception as error:
        CFG['error'] = str(error)
        print('Experiment error: '+str(error), flush=True)
    finally:
        CFG['finished_utc'] = now()
        (ROOT / 'run.json').write_text(json.dumps(CFG, indent=2)+'\n')
        logs = call(COMPOSE + ['logs', '--no-color'], check=False)
        (ROOT / 'server.log').write_text(logs.stdout + logs.stderr)
        call(['docker', 'cp', CONTAINER+':/home/steam/tf-dedicated/tf/console.log', str(ROOT / 'console.log')], check=False)
        report()
        call(COMPOSE + ['stop'], check=False)
        print('Isolated server stopped. Report: '+str(ROOT / 'REPORT.md'), flush=True)


if __name__ == '__main__':
    main()
