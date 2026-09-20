#!/usr/bin/env bash
# Run every catalog wave, with and without Bot Surge, on disposable Docker servers.
set -euo pipefail

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$root"
shards=${WAVEPROBE_SHARDS:-6}
speed=${WAVEPROBE_SPEED:-10}
seed=${WAVEPROBE_SEED:-1}
timeout=${WAVEPROBE_TIMEOUT:-90s}
base_port=${WAVEPROBE_BASE_PORT:-27035}
source_volume=${WAVEPROBE_SOURCE_VOLUME:-tf2-archipelago-waveprobe_tf2game_waveprobe}
run_dir=${WAVEPROBE_RUN_DIR:-$root/docs/audits/waveprobe-$(date -u +%Y%m%d-%H%M%S)}
: "${WAVEPROBE_RCONPW:?set a unique password for the disposable servers}"
[[ $shards =~ ^[1-9][0-9]*$ ]] || { echo 'WAVEPROBE_SHARDS must be positive' >&2; exit 2; }
[[ $base_port =~ ^[1-9][0-9]*$ ]] || { echo 'WAVEPROBE_BASE_PORT must be positive' >&2; exit 2; }
mkdir -p "$run_dir"
docker volume inspect "$source_volume" >/dev/null
printf 'shards=%s\nspeed=%s\nseed=%s\nwave_timeout=%s\nstarted_utc=%s\n' \
    "$shards" "$speed" "$seed" "$timeout" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$run_dir/config.txt"

# Build the two plugins from the current checkout before starting any shard.
"$root/plugin/build.sh" > "$run_dir/plugin-build.log" 2>&1
# shellcheck source=/dev/null
source "$root/deploy/env/versions.env"
compiler="$root/plugin/build/sourcemod-$SOURCEMOD_VERSION/addons/sourcemod/scripting/spcomp64"
[[ -x $compiler ]] || compiler="${compiler%64}"
"$compiler" -E -o"$root/plugin/build/tf2_waveprobe.smx" \
    "$root/plugin/scripting/tf2_waveprobe.sp" > "$run_dir/probe-build.log" 2>&1
(cd "$root" && go build -o "$run_dir/waveprobe" ./launcher/cmd/waveprobe)
"$run_dir/waveprobe" -plan -mode both > "$run_dir/plan.jsonl"

declare -a pids=()
for ((i=0; i<shards; i++)); do
    project=tf2-archipelago-waveprobe
    ((i == 0)) || project+=-$i
    volume="${project}_tf2game_waveprobe"
    if ! docker volume inspect "$volume" >/dev/null 2>&1; then
        docker volume create "$volume" >/dev/null
        echo "copying isolated game files to $volume" >&2
        docker run --rm -v "$source_volume:/src:ro" -v "$volume:/dst" \
            debian:bookworm-slim sh -c 'cp -a /src/. /dst/'
    fi
    port=$((base_port + i))
    WAVEPROBE_RCON_PORT=$port docker compose -p "$project" \
        -f "$root/deploy/compose.waveprobe.yml" up -d --force-recreate > "$run_dir/docker-$i.log" 2>&1
    echo "shard $i/$shards on 127.0.0.1:$port" >&2
    "$run_dir/waveprobe" -rcon "127.0.0.1:$port" -shards "$shards" -shard "$i" \
        -mission all -mode both -speed "$speed" -seed "$seed" -timeout "$timeout" \
        > "$run_dir/shard-$i.jsonl" 2> "$run_dir/shard-$i.err" &
    pids+=("$!")
done

failed=0
for ((i=0; i<shards; i++)); do
    wait "${pids[i]}" || failed=1
done
python3 "$root/deploy/waveprobe-report.py" "$run_dir" > "$run_dir/REPORT.md"
echo "Report: $run_dir/REPORT.md" >&2
for ((i=0; i<shards; i++)); do
    project=tf2-archipelago-waveprobe
    ((i == 0)) || project+=-$i
    WAVEPROBE_RCON_PORT=$((base_port + i)) docker compose -p "$project" \
        -f "$root/deploy/compose.waveprobe.yml" stop > /dev/null
done
exit "$failed"
