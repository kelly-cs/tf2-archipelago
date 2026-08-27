#!/bin/sh
# Push a Docker image, and try again when the registry answers with a layer it
# does not hold whole.
#
# The nightly cancels its own runs on purpose: two of them racing would both
# delete and recreate the `nightly` tag, and the loser would leave the release
# describing a commit the tag no longer points at. A run cancelled part-way
# through a push leaves the blob it was uploading incomplete, and the next run
# that refers to that layer gets "unknown blob" and fails.
#
# Re-pushing uploads the layer again, which is the whole fix. It covers the
# registry having a bad minute as well, which is the other way this line fails.
#
# Bounded: three attempts, then the failure is real and the step should show it.
set -eu

ATTEMPTS=3
DELAY=5

attempt=1
while :; do
	if docker push "$@"; then
		exit 0
	fi
	if [ "$attempt" -ge "$ATTEMPTS" ]; then
		echo "push failed $ATTEMPTS times: $*" >&2
		exit 1
	fi
	echo "push failed, trying again in ${DELAY}s (attempt $attempt of $ATTEMPTS)" >&2
	sleep "$DELAY"
	attempt=$((attempt + 1))
	DELAY=$((DELAY * 2))
done
