#!/bin/sh
# Two modes. `generate` writes one seed into the output directory and stops,
# which is what `make seed` runs before the file goes to archipelago.gg. No
# argument generates a seed if the directory holds none, then hosts it.
#
# Regenerating on every start would hand the players a different multiworld each
# evening and invalidate everything the bridge has recorded.
set -eu

mode="${1:-host}"

AP_PORT="${AP_PORT:-38281}"
AP_SLOT_NAME="${AP_SLOT_NAME:-tf2}"
AP_PASSWORD="${AP_PASSWORD:-}"
AP_VERSION="${ARCHIPELAGO_VERSION:?the image must set ARCHIPELAGO_VERSION}"

output=/ap/output
players=/ap/Players

# The exported tables. player-yaml.py reads the game name out of them, and the
# popfile to mission name mapping that the apworld's options need.
data=/ap/custom_worlds

# Generation runs in a directory of its own, so the archive that comes back is
# the one this run made rather than the oldest one left in the output.
generate() {
	mkdir -p "$players"
	python /usr/local/bin/player-yaml.py "$data" "$AP_VERSION" > "$players/tf2.yaml"

	fresh=$(mktemp -d)
	python Generate.py --player_files_path "$players" --outputpath "$fresh" < /dev/null

	made=$(find "$fresh" -maxdepth 1 -name 'AP_*.zip' | head -n 1)
	if [ -z "$made" ]; then
		echo "generation produced no archive" >&2
		exit 1
	fi

	mv "$made" "$output/"
	rm -rf "$fresh"
	archive="$output/$(basename "$made")"
}

case "$mode" in
generate)
	generate
	echo "generated $archive"
	echo "upload it at https://archipelago.gg/uploads, then create a room"
	exit 0
	;;
host) ;;
*)
	echo "unknown mode $mode, expected generate or host" >&2
	exit 1
	;;
esac

archive=$(find "$output" -maxdepth 1 -name 'AP_*.zip' | sort | head -n 1)

if [ -z "$archive" ]; then
	echo "no seed in $output, generating one"
	generate
fi

echo "hosting $archive on port $AP_PORT"
set -- python MultiServer.py --port "$AP_PORT" --host 0.0.0.0
if [ -n "$AP_PASSWORD" ]; then
	set -- "$@" --password "$AP_PASSWORD"
fi
exec "$@" "$archive" < /dev/null
