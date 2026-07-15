#!/bin/sh
set -eu

if [ "$#" -ne 2 ]; then
    echo "usage: smoke_candidate_arch.sh <candidate-image@sha256:digest> <amd64|arm64>" >&2
    exit 2
fi

image=$1
architecture=$2
case "$architecture" in
    amd64|arm64) ;;
    *)
        echo "candidate architecture smoke: unsupported architecture: $architecture" >&2
        exit 2
        ;;
esac
case "$image" in
    *@sha256:*) ;;
    *)
        echo "candidate architecture smoke: candidate image must use an immutable digest" >&2
        exit 2
        ;;
esac
digest=${image##*@sha256:}
case "$digest" in
    *[!0-9a-f]*|"")
        echo "candidate architecture smoke: malformed sha256 digest" >&2
        exit 2
        ;;
esac
if [ "${#digest}" -ne 64 ]; then
    echo "candidate architecture smoke: malformed sha256 digest" >&2
    exit 2
fi

output=$(docker run --rm --platform "linux/$architecture" "$image" version 2>&1) || {
    status=$?
    printf '%s\n' "$output" >&2
    echo "candidate architecture smoke: $architecture failed to boot" >&2
    exit "$status"
}
printf '%s\n' "$output"
case "$output" in
    *"wowapi "*) ;;
    *)
        echo "candidate architecture smoke: $architecture returned an unexpected version response" >&2
        exit 1
        ;;
esac
echo "candidate architecture smoke: $architecture passed"
