#!/usr/bin/env python3
"""Confirm the fastest failure-free setting across additional full-mission seeds."""
import importlib.util, json, re, time, subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('experiment',ROOT/'run.py')
e=importlib.util.module_from_spec(spec);spec.loader.exec_module(e)
e.TRIALS=json.loads((ROOT/'trials.json').read_text())
values=json.loads(subprocess.check_output(['docker','inspect','--format','{{json .Config.Env}}',e.CONTAINER],text=True))
password=next(v.split('=',1)[1] for v in values if v.startswith('SRCDS_RCONPW='))
e.ENV.update(SRCDS_RCONPW=password,WAVEPROBE_RCONPW=password)
e.CFG['confirmation_started_utc']=e.now()
e.CFG.pop('error',None)
def passed(t):return t['first_pass_outcomes']=={'passed':14} and t['first_pass_cases']==14 and t['crash_free']
e.call(['docker','start',e.CONTAINER])
try:
    deadline=time.monotonic()+600
    while time.monotonic()<deadline:
        try:
            evidence=''.join(e.rcon(c) for c in ['version','sm plugins list','sm exts list','sm_waveprobe_status'])
            if ('TF2 MvM Archipelago' in evidence and 'TF2 MvM wave probe' in evidence and
                not re.search(r'<Failed>|<Error>|Unknown command',evidence,re.I)):
                (ROOT/'confirmation-startup.txt').write_text(evidence)
                break
        except (subprocess.SubprocessError,OSError):pass
        time.sleep(3)
    else:raise RuntimeError('confirmation server did not become ready')
    speeds=sorted({t['speed'] for t in e.TRIALS if t['speed']<65},reverse=True)
    e.CFG['fastest_confirmed']=None
    for speed in speeds:
        if any(not passed(t) for t in e.TRIALS if t['speed']==speed):continue
        print('Confirming '+str(speed)+'x across seeds 2, 3, 4.',flush=True)
        clean=True
        for seed in [2,3,4]:
            e.CFG['seed']=seed
            repeat=max(t['repeat'] for t in e.TRIALS if t['speed']==speed)+1
            result=e.trial(speed,repeat);e.report()
            if not passed(result):
                clean=False
                break
        if clean:
            e.CFG['fastest_confirmed']=speed
            break
    e.CFG['latest_after']=e.live_version()
    e.CFG['fresh']=e.CFG['latest_after']==e.CFG['running_version']==e.CFG['latest_before']
except Exception as error:
    e.CFG['confirmation_error']=str(error)
    print('Confirmation error: '+str(error),flush=True)
finally:
    e.CFG['finished_utc']=e.now()
    (ROOT/'run.json').write_text(json.dumps(e.CFG,indent=2)+'\n')
    logs=e.call(['docker','logs',e.CONTAINER],check=False)
    (ROOT/'server.log').write_text(logs.stdout+logs.stderr)
    e.report()
    e.call(['docker','stop',e.CONTAINER],check=False)
    print('Confirmation ended; isolated worker stopped.',flush=True)
