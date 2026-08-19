#!/bin/sh
# Render a command's terminal output into an SVG for the book and the README.
#
# The Linux launcher has no window: walk is a Win32 binding, so a terminal is
# what there is to show. SVG rather than a screenshot of one, for three
# reasons. It is text, so a diff says what changed in a capture. It is sharp at
# any size, on any display. And it uses the reader's own monospace font, so
# nothing here depends on which fonts this machine happened to have installed,
# which is what made the Wine captures of the Windows window look wrong.
#
# Usage:
#   docs/capture.sh <title> <output.svg> <<'EOF'
#   ... the text to draw ...
#   EOF
#
# Or pipe a real command through it:
#   ./dist/tf2ap-linux-amd64 -status | docs/capture.sh 'tf2ap -status' out.svg
set -eu

title=${1:?usage: capture.sh <title> <output.svg>}
output=${2:?usage: capture.sh <title> <output.svg>}

# The window's own measurements, in pixels. The character box is what a 14px
# monospace glyph occupies; every other number is derived from it, so changing
# the size here moves the whole capture together.
char_width=8.4
line_height=20
font_size=14
padding=18
bar_height=34

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

# Strip the colours and the carriage returns a console leaves behind, then
# escape the three characters XML cannot take raw.
sed -e 's/\x1b\[[0-9;]*[a-zA-Z]//g' -e 's/\r$//' \
	-e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' > "$tmp"

lines=$(wc -l < "$tmp")

# The window is as wide as its widest line, and never narrower than the title
# plus the three dots beside it, which is what would otherwise overlap.
title_columns=$((${#title} + 12))
columns=$(awk -v floor="$title_columns" \
	'{ if (length($0) > n) n = length($0) } END { print (n < floor ? floor : n) }' "$tmp")

width=$(awk "BEGIN { printf \"%.0f\", $columns * $char_width + $padding * 2 }")
height=$(awk "BEGIN { printf \"%.0f\", $lines * $line_height + $bar_height + $padding * 2 }")

{
	printf '<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" font-family="ui-monospace, SFMono-Regular, Menlo, Consolas, monospace" font-size="%s">\n' \
		"$width" "$height" "$width" "$height" "$font_size"
	printf '<rect width="%s" height="%s" rx="10" fill="#1b1f27"/>\n' "$width" "$height"
	printf '<rect width="%s" height="%s" rx="10" fill="#252b36"/>\n' "$width" "$bar_height"
	printf '<rect y="%s" width="%s" height="10" fill="#252b36"/>\n' \
		"$(awk "BEGIN { print $bar_height - 10 }")" "$width"

	# The three dots every terminal window has, so the picture reads as one.
	dot=0
	for colour in '#ff5f57' '#febc2e' '#28c840'; do
		printf '<circle cx="%s" cy="17" r="6" fill="%s"/>\n' \
			"$(awk "BEGIN { print 20 + $dot * 20 }")" "$colour"
		dot=$((dot + 1))
	done
	printf '<text x="%s" y="22" fill="#8b94a7">%s</text>\n' \
		"$(awk "BEGIN { print $width / 2 }")" "$title" | sed 's/<text /<text text-anchor="middle" /'

	y=$(awk "BEGIN { print $bar_height + $padding + $font_size }")
	while IFS= read -r line; do
		printf '<text x="%s" y="%s" fill="#d5dae5" xml:space="preserve">%s</text>\n' \
			"$padding" "$y" "$line"
		y=$(awk "BEGIN { print $y + $line_height }")
	done < "$tmp"
	printf '</svg>\n'
} > "$output"

echo "wrote $output (${width}x${height}, $lines lines)"
