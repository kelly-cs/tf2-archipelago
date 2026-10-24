#!/usr/bin/env python3
"""Summarize first-pass speed evidence, never letting retries erase failures."""
from pathlib import Path
import collections, json, statistics, re, hashlib
ROOT = Path(__file__).resolve().parent
cfg=json.loads((ROOT/'run.json').read_text())
trials=json.loads((ROOT/'trials.json').read_text()) if (ROOT/'trials.json').exists() else []
rows=[]
for speed in sorted({t['speed'] for t in trials}):
    runs=[t for t in trials if t['speed']==speed]
    raw=[json.loads(line) for t in runs for line in (ROOT/t['directory']/'shard-0.jsonl').read_text().splitlines() if line.strip()]
    outcomes=collections.Counter(r.get('outcome',r['state']) for r in raw)
    successful=sum(t['first_pass_outcomes']=={'passed':14} and t['first_pass_cases']==14 and t['crash_free'] for t in runs)
    mean=statistics.mean(t['elapsed_seconds'] for t in runs)
    rows.append(dict(speed=speed,runs=len(runs),complete_clean_runs=successful,passed=outcomes['passed'],expected=14*len(runs),outcomes=dict(outcomes),mean_seconds=mean,min_seconds=min(t['elapsed_seconds'] for t in runs),max_seconds=max(t['elapsed_seconds'] for t in runs),seeds=sorted({t['seed'] for t in runs}),by_mode={m:dict(collections.Counter(r.get('outcome',r['state']) for r in raw if r['mode']==m)) for m in ['normal','surge']}))
base=next((r['mean_seconds'] for r in rows if r['speed']==20),None)
lines=['# Decoy speed comparison with the production robot limiter', '',f"Source `{cfg.get('source_commit','pending')}` with the archived experimental speed-only overlay. Live TF2 `{cfg.get('running_version','pending')}`.", '', 'Single isolated Linux Docker worker, four CPUs and 4 GiB RAM. Every run covers seven Decoy Expert waves, normal and Bot Surge. Wake-up fix enabled; robot limiter remains enabled; maximum invaders remains 22. No wave, mission-cycle, kill timing or polling changes.', '', '| Speed | Clean full runs | First-pass passes / planned | Mean wall s | Wall range s | Gain vs 20× | Seeds |','| ---: | ---: | ---: | ---: | ---: | ---: | --- |']
for r in rows:
    gain=f"{base/r['mean_seconds']:.2f}×" if base else 'pending'
    lines.append(f"| {r['speed']}× | {r['complete_clean_runs']}/{r['runs']} | {r['passed']}/{r['expected']} | {r['mean_seconds']:.1f} | {r['min_seconds']:.1f}–{r['max_seconds']:.1f} | {gain} | {r['seeds']} |")
lines+=['', 'First-pass failures and missing cases remain failures in this comparison. A representative nonpass receives a separate 900-second diagnostic retry; that retry never changes the first-pass result. Raw JSONL retains completion events, enemy counts, timelines and failure snapshots.', '', 'Wall times include probe mission/mode setup and polling, and exclude the initial installation and explicit fresh-map reload before each trial. Spawn totals vary even between controls, so successful completion does not prove identical population behavior.', '', 'This bounds speeds for this one map and this local worker. It does not certify all maps, Windows, Ubuntu native or real human join/wake behavior. Daily speed remains 20× until wider coverage is measured.', '']
lines+=['Setup caveat: the first cold start lacked custom plugins until an explicit restart after staging; see SETUP-NOTES.md and cold-start.log. This is not a clean cold-start certification. No bridge or multiworld is connected, so bridge errors and later objective-queue overflow messages remain in the raw logs; this measures server wave execution only.', '']
if cfg.get('finished_utc'):lines+=['Finished: '+cfg['finished_utc']+'. Latest-build verification: '+str(cfg.get('fresh'))+'.', '']
if cfg.get('error'):lines+=['Run error: '+cfg['error'], '']
(ROOT/'REPORT.md').write_text('\n'.join(lines))
(ROOT/'RESULTS.json').write_text(json.dumps({'configuration':cfg,'speeds':rows},indent=2)+'\n')
print('\n'.join(lines))
