#!/bin/sh
# Fail when a pinned Go image disagrees with the go directive in go.mod.
#
# go.mod owns the version, because that directive has to be a literal and every
# other pin can be derived from it. Three places cannot derive it: the bridge
# image's FROM line, the srcds image's bots stage, which needs a toolchain to
# stage the defender mod out of the module cache, and the container image ci.yml
# runs its go job in. None of them reads a file, so all three are edited by hand
# and all three are forgettable.
#
# ci.yml's plugin job is not in the list on purpose: it takes its toolchain from
# actions/setup-go with go-version-file, so it derives the version rather than
# repeating it and there is nothing here to check.
#
# Forgetting is not harmless. golangci-lint v2.12.2 carried a staticcheck that
# panicked on the Go 1.27 standard library, so a tree on 1.27 with a builder
# still on 1.26 could not lint at all. The failure looked like a linter bug.
set -eu

want=$(sed -n 's/^go //p' go.mod)
if [ -z "$want" ]; then
	echo "go.mod has no go directive" >&2
	exit 1
fi

status=0
for file in deploy/Dockerfile.bridge deploy/Dockerfile.srcds .github/workflows/ci.yml; do
	pinned=$(grep -oE 'golang:[0-9]+\.[0-9]+(\.[0-9]+)?' "$file" | head -1 | cut -d: -f2)
	if [ -z "$pinned" ]; then
		echo "$file pins no golang image, so this check cannot see it" >&2
		status=1
		continue
	fi
	# A pin may be shorter than the go directive: golang:1.27 serves 1.27.0.
	case "$want" in
	"$pinned" | "$pinned".*) ;;
	*)
		echo "$file pins golang:$pinned, go.mod says $want" >&2
		status=1
		;;
	esac
done

if [ "$status" -ne 0 ]; then
	echo "the pinned Go images disagree with go.mod" >&2
	exit 1
fi
echo "go $want, and every pinned image agrees"
