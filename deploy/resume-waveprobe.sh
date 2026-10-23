#!/usr/bin/env bash
# Resume an interrupted sweep without losing witnessed passes.
set -euo pipefail

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
run_dir=${1:?usage: bash deploy/resume-waveprobe.sh <run-directory>}
run_dir=$(cd -- "$run_dir" && pwd)
[[ -f $run_dir/plan.jsonl && -x $run_dir/waveprobe && -f $run_dir/projects.txt ]] || {
    echo 'The run directory needs plan.jsonl, waveprobe, and projects.txt.' >&2
    exit 2
}
mapfile -t projects < "$run_dir/projects.txt"
(( ${#projects[@]} > 0 )) || { echo 'No disposable server projects were recorded.' >&2; exit 2; }

finish() {
    result=$?
    trap - EXIT INT TERM
    python3 "$root/deploy/waveprobe-report.py" "$run_dir" > "$run_dir/REPORT.md" || result=1
    for project in "${projects[@]}"; do
        docker stop "${project}-srcds-1" > /dev/null 2>&1 || true
    done
    echo "Report: $run_dir/REPORT.md" >&2
    exit "$result"
}
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

if [[ -z ${WAVEPROBE_RCONPW:-} ]]; then
    WAVEPROBE_RCONPW=$(docker inspect "${projects[0]}-srcds-1" \
        --format '{{range .Config.Env}}{{println .}}{{end}}' | sed -n 's/^SRCDS_RCONPW=//p')
    export WAVEPROBE_RCONPW
fi
: "${WAVEPROBE_RCONPW:?the disposable server RCON password is unavailable}"
for project in "${projects[@]}"; do
    docker start "${project}-srcds-1" > /dev/null
done
workers=$(seq -s, 0 "$((${#projects[@]} - 1))")
WAVEPROBE_PHASE=screen python3 "$root/deploy/waveprobe-retest.py" \
    "$run_dir" "$run_dir/waveprobe" "$workers" > "$run_dir/screen.log" 2>&1
python3 "$root/deploy/waveprobe-retest.py" \
    "$run_dir" "$run_dir/waveprobe" "$workers" > "$run_dir/retest.log" 2>&1
