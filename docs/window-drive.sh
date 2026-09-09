#!/bin/sh
# Drive the settings window, so a button that does nothing is caught here.
#
# window-shot.sh photographs the window; this one uses it. The Load button
# shipped doing nothing at all, twice, because a screenshot of a window nobody
# clicked looks exactly like a working one.
#
# What it does: name a team, save it, move a seat away from what was saved,
# load it back, save a second team and remove one, photographing each step.
# Seat 1 should read Pyro in step 4 and Scout in step 3, and config.json should
# hold one team at the end and two before the Remove.
#
# The coordinates are measured off a capture rather than guessed, and they move
# when a button is added to a row: a click that lands on nothing looks exactly
# like a button that does nothing, which is the failure this script is for.
#
# Two things about driving Win32 menus with xdotool, both learned the hard way:
# a click opens the drop-down list, which then swallows the next click, so the
# list is closed with Escape and the selection moved with the arrow keys; and
# xdotool type goes to whatever has focus, so the box is clicked first and the
# result photographed rather than assumed.
#
# Needs the same tools as window-shot.sh. Usage:
#   docs/window-drive.sh dist/tf2ap.exe /tmp/preset.png
set -eu
exe=${1:?usage: window-drive.sh <exe> <output.png>}
out=${2:?usage: window-drive.sh <exe> <output.png>}
if [ "${WINDOW_SHOT_DISPLAY:-}" = 1 ]; then
	xsetroot -solid '#f0f0f0'
	wine "$exe" >/dev/null 2>&1 &
	sleep 28
	w=$(xdotool search --name Settings | tail -n 1)
	eval "$(xdotool getwindowgeometry --shell "$w")"
	xdotool mousemove $((X + 345)) $((Y + 19)) click 1; sleep 3

	# Seat 1: click to focus, Escape to shut the list, then the keyboard moves
	# the selection without a popup in the way.
	xdotool mousemove $((X + 455)) $((Y + 204)) click 1; sleep 1
	xdotool key Escape; sleep 1
	xdotool key Down Down Down; sleep 1
	import -window "$w" "${out%.png}-1-seat.png"

	# Name it, then Save the team.
	xdotool mousemove $((X + 480)) $((Y + 178)) click 1; sleep 1
	xdotool type --delay 80 "three down"; sleep 1
	xdotool mousemove $((X + 731)) $((Y + 178)) click 1; sleep 2
	import -window "$w" "${out%.png}-2-saved.png"

	# Move seat 1 away from what was saved.
	xdotool mousemove $((X + 455)) $((Y + 204)) click 1; sleep 1
	xdotool key Escape; sleep 1
	xdotool key Up Up; sleep 1
	import -window "$w" "${out%.png}-3-changed.png"

	# Load it back, then close the dialog with its own Save so the seats reach
	# the config file, which is what the assertion reads.
	xdotool mousemove $((X + 650)) $((Y + 178)) click 1; sleep 2
	import -window "$w" "${out%.png}-4-loaded.png"
	# Save a second team, then remove whichever the menu is showing: the one
	# that is left says Remove took the right one and kept the rest.
	xdotool mousemove $((X + 455)) $((Y + 204)) click 1; sleep 1
	xdotool key Escape; sleep 1; xdotool key Up; sleep 1
	xdotool mousemove $((X + 480)) $((Y + 178)) click 1; sleep 1
	xdotool type --delay 60 "one up"; sleep 1
	xdotool mousemove $((X + 731)) $((Y + 178)) click 1; sleep 2
	import -window "$w" "${out%.png}-5-two-teams.png"

	xdotool mousemove $((X + 812)) $((Y + 178)) click 1; sleep 2
	import -window "$w" "${out%.png}-6-removed.png"

	# The dialog's own Save. It refuses to close while the room address is
	# empty and test mode is off, which is why the seats reach config.json
	# only on a prefix that has a room set.
	xdotool mousemove $((X + 755)) $((Y + 500)) click 1; sleep 3
	exit 0
fi
export TF2AP_INSTALL_ROOT='C:\tf2-archipelago'
export SRCDS_RCONPW=hidden
WINDOW_SHOT_DISPLAY=1 exec xvfb-run -a -s "-screen 0 1920x1400x24" "$0" "$exe" "$out"
