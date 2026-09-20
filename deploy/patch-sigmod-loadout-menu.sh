#!/bin/sh
# The pinned SigMod release formats an empty phrase name for unselected extra
# loadout items. SourceMod rejects that %t and displays the format literally.
# Change only that final placeholder to %s before the package is staged.
set -eu

root=$1
title_patch=$2
for relative in \
	addons/sourcemod/extensions/sigsegv.ext.2.tf2.so \
	addons/sourcemod/extensions/x64/sigsegv.ext.2.tf2.so
do
	file="$root/$relative"
	[ -f "$file" ] || { echo "missing SigMod extension: $file" >&2; exit 1; }
	count=$(LC_ALL=C grep -aoF '%t: %s %s %t' "$file" | wc -l)
	[ "$count" -eq 1 ] || { echo "unexpected SigMod menu marker count in $file: $count" >&2; exit 1; }
	before=$(wc -c <"$file")
	LC_ALL=C sed -i 's/%t: %s %s %t/%t: %s %s %s/' "$file"
	[ "$(wc -c <"$file")" -eq "$before" ] || { echo "SigMod patch changed the size of $file" >&2; exit 1; }
	[ "$(LC_ALL=C grep -aoF '%t: %s %s %s' "$file" | wc -l)" -eq 1 ] || {
		echo "SigMod menu patch did not land in $file" >&2
		exit 1
	}
	arch=x86
	case "$relative" in */x64/*) arch=x64 ;; esac
	perl "$title_patch" "$file" "$arch"
done
