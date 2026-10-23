# Headless wave probe wake-up comparison — 2026-10-23

The probe now sets and verifies `tf_allow_server_hibernation 0` before setting
up waves, matching the pinned SigMod Linux and Windows beds. Fake clients do
not wake MvM's robot allocator as a human connection does. The robot limiter
remained **enabled**, the invader cap remained **22**, and speed remained **20×**.
No production mod, dependency pin, mission file or population rule changed.

Two complete Decoy Expert sweeps each passed all seven waves with Bot Surge
off and on: **28/28 first-pass cases**, without diagnostic retries. Wall times
were **92.0 s** and **90.1 s**. TF2 ServerVersion **11138802** matched Valve's
live version before and after. The fixed session's captured logs contained no
segmentation fault, fatal error or access-violation markers. The server ran in
an isolated Linux Docker worker with four CPUs and 4 GiB RAM.

Earlier limiter-enabled diagnostics stalled. Two fresh attempts in this session
crashed before the planned mid-wave intervention could be applied, including
one after fresh pinned addon staging. Those failures remain in the local raw
audits; they do not establish a crash root cause or prove that wake-up fixes
every crash. Only the fixed session is claimed as passing here.

Raw `repeat-*.jsonl`, before/after cvar readbacks, the runner diff and
`evidence.json` accompany this report. Full local baseline and fixed logs live
under `docs/audits/limiter-wake*-20261023-*` in the original working checkout;
large game copies and executable artifacts are not committed.

This validates one Linux Docker mission, not all maps or Windows/Ubuntu native
startup. The daily run's separate Ubuntu engine startup failure and Windows
RCON discovery failure remain unresolved by this change. Headless wake-up does
not validate real human joins or idle hibernation/wake behavior.

The dashboard additionally requires the robot limiter to read enabled before
and after the sweep. Recorded fatal engine errors invalidate a score even if
srcds_run automatically restarts and later probes pass. First-pass results,
timeouts, missing coverage and failures remain visible.
