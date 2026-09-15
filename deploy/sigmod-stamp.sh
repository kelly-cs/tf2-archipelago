#!/bin/sh
# Write the same pinned-install receipt the native launcher verifies. The
# srcds image runs this against its staged SigMod tree; sync_tree then carries
# the receipt into the persistent game volume with the files it describes.
set -eu

root=$1
version=$2
archive_sha256=$3
stamp="$root/addons/.tf2ap-sigsegv-mvm.stamp"

mkdir -p "$(dirname "$stamp")"
{
	printf '%s\n%s\n' "$version" "$archive_sha256"
	for relative in \
		addons/sourcemod/extensions/sigsegv.ext.2.tf2.so \
		addons/sourcemod/extensions/x64/sigsegv.ext.2.tf2.so \
		addons/sourcemod/extensions/sigsegv.autoload \
		addons/sourcemod/gamedata/sigsegv/population.txt \
		cfg/sigsegv_convars.cfg
	do
		printf '%s ' "$relative"
		sha256sum "$root/$relative" | cut -d ' ' -f 1
	done
} >"$stamp"
