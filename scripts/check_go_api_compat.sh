#!/bin/sh
set -eu

APIDIFF_VERSION="v0.0.0-20260709172345-9ea1abe57597"

if [ "$#" -ne 2 ]; then
    echo "usage: check_go_api_compat.sh <baseline-module-dir> <current-module-dir>" >&2
    exit 2
fi

case "$1" in
    /*) baseline_dir=$1 ;;
    *) baseline_dir="$(pwd -P)/$1" ;;
esac
case "$2" in
    /*) current_dir=$2 ;;
    *) current_dir="$(pwd -P)/$2" ;;
esac

if [ ! -f "$baseline_dir/go.mod" ]; then
    echo "go API compatibility: baseline has no go.mod: $baseline_dir" >&2
    exit 2
fi
if [ ! -f "$current_dir/go.mod" ]; then
    echo "go API compatibility: current has no go.mod: $current_dir" >&2
    exit 2
fi

baseline_module=$(cd "$baseline_dir" && go list -m -f '{{.Path}}')
current_module=$(cd "$current_dir" && go list -m -f '{{.Path}}')
if [ "$baseline_module" != "$current_module" ]; then
    echo "go API compatibility: module paths differ: $baseline_module != $current_module" >&2
    exit 2
fi

tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/wowapi-apidiff.XXXXXX")
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

(
    cd "$baseline_dir"
    go run "golang.org/x/exp/cmd/apidiff@$APIDIFF_VERSION" -m -w "$tmp_dir/baseline.api" "$baseline_module"
)
(
    cd "$current_dir"
    go run "golang.org/x/exp/cmd/apidiff@$APIDIFF_VERSION" -m -w "$tmp_dir/current.api" "$current_module"
)

incompatible=$(go run "golang.org/x/exp/cmd/apidiff@$APIDIFF_VERSION" -m -incompatible "$tmp_dir/baseline.api" "$tmp_dir/current.api")
if [ -n "$incompatible" ]; then
    printf '%s\n' "$incompatible" >&2
    echo "go API compatibility: breaking public API change detected" >&2
    exit 1
fi

echo "go API compatibility: compatible"
