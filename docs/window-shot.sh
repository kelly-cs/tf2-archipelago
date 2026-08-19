#!/bin/sh
# Photograph the launcher's window.
#
# The window is Win32, so a machine that is not running Windows needs Wine and
# a display to put it on. Both are here rather than in a person's hands because
# a screenshot taken by hand is a screenshot nobody retakes: the last three
# were a release behind and showed a tab that had been renamed and a button
# that had been added.
#
# Needs wine, xvfb-run, xdotool and ImageMagick's import. Debian:
#   sudo apt-get install --no-install-recommends wine xvfb xdotool imagemagick
#
# Usage:
#   docs/window-shot.sh <exe> <output.png> [seconds to wait] [main|dialog]
#
# The launcher is run with no room address, which is what a first evening
# starts from: it opens the settings dialog by itself and starts no server.
# "dialog" shoots that; "main" closes it first and shoots the window behind.
set -eu

exe=${1:?usage: window-shot.sh <exe> <output.png> [wait] [main|dialog]}
output=${2:?usage: window-shot.sh <exe> <output.png> [wait] [main|dialog]}
wait=${3:-15}
what=${4:-main}

# The second half of this script runs inside the virtual display, which is why
# it re-runs itself: an inline shell string would need every dollar sign in it
# escaped twice, and the escaping is what breaks.
if [ "${WINDOW_SHOT_DISPLAY:-}" = 1 ]; then
	# The tab strip has a strip beside the last tab that the control never
	# paints, and on a fresh display what shows through is whatever the frame
	# buffer held, which is black. Windows has the button face colour there,
	# so the display does too.
	xsetroot -solid '#f0f0f0'

	wine "$exe" >/dev/null 2>&1 &
	sleep "$wait"

	settings=$(xdotool search --name Settings | tail -n 1)
	main=$(xdotool search --name "Mann vs Archipelago" | tail -n 1)

	if [ "$what" = dialog ]; then
		[ -n "$settings" ] || { echo "window-shot: the settings dialog never opened"; exit 1; }
		import -window "$settings" "$output"
	else
		# The dialog sits on top of the window this shoots, so it goes first.
		[ -n "$settings" ] && xdotool key --window "$settings" Escape
		sleep 2
		[ -n "$main" ] || { echo "window-shot: no launcher window"; exit 1; }
		import -window "$main" "$output"
	fi

	wineserver -k >/dev/null 2>&1 || true
	exit 0
fi

command -v wine >/dev/null || { echo "window-shot: no wine"; exit 1; }
command -v xvfb-run >/dev/null || { echo "window-shot: no xvfb-run"; exit 1; }
command -v xdotool >/dev/null || { echo "window-shot: no xdotool"; exit 1; }
command -v import >/dev/null || { echo "window-shot: no import"; exit 1; }

# A prefix of the build's own, so the shot never carries whatever else a Wine
# prefix on this machine has been used for. It is kept between runs: building
# one takes the best part of a minute, and every shot after the first is then
# a few seconds.
WINEPREFIX=${WINEPREFIX:-$PWD/dist/wineprefix}
export WINEPREFIX WINEDEBUG=-all
mkdir -p "$WINEPREFIX"
# Nothing of whoever ran this belongs in a committed picture. Wine builds the
# Windows user directory from the host account, so the path to the Archipelago
# app is given rather than found.
export TF2AP_INSTALL_ROOT='C:\tf2-archipelago'
export TF2AP_ARCHIPELAGO_DIR='C:\Program Files\Archipelago'
export SRCDS_RCONPW=hidden

WINDOW_SHOT_DISPLAY=1
export WINDOW_SHOT_DISPLAY

xvfb-run -a -s "-screen 0 1600x1000x24" "$0" "$exe" "$output" "$wait" "$what"

echo "wrote $output"
