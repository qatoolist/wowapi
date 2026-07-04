#!/usr/bin/env bash
# check_unwired.sh — heuristic for the "built-but-not-wired" pattern (the #1
# recurring review miss: a primitive exists but nothing on the real path calls it).
#
# For a given kernel package it lists exported constructors/types (New*, *Registry,
# *Engine, *Service, *Writer, *Store, *Pipeline, *Tracer) that are NOT referenced
# anywhere outside that package's own non-test files. Such a symbol is a candidate
# for "never wired into kernel.New / module.Context / app boot".
#
# NOTE: the kernel is a library, so some exports are legitimately the product-facing
# API and will show up here — this is an ATTENTION LIST, not a verdict. Confirm each
# hit is either wired internally or intentionally product-facing.
#
# Usage:   miscellaneous/check_unwired.sh kernel/retention
#          miscellaneous/check_unwired.sh              # scans all kernel/* packages
# Read-only. Exit 0 always (advisory).
set -euo pipefail
cd "$(dirname "$0")/.."

scan_pkg() {
    local pkg="$1"
    [ -d "$pkg" ] || return 0
    local symbols
    symbols="$(grep -rhoE --include='*.go' --exclude='*_test.go' \
        '^(func (New[A-Za-z0-9_]+)|type ([A-Za-z0-9_]*(Registry|Engine|Service|Writer|Store|Pipeline|Tracer|Authenticator|Scheduler)))' \
        "$pkg" 2>/dev/null | sed -E 's/^func (New[A-Za-z0-9_]+).*/\1/; s/^type ([A-Za-z0-9_]+).*/\1/' | sort -u || true)"
    [ -z "$symbols" ] && return 0

    local flagged=""
    for sym in $symbols; do
        # references outside this package's non-test .go files
        local refs
        refs="$(grep -rl --include='*.go' --exclude='*_test.go' "\b${sym}\b" . 2>/dev/null \
                | grep -v "^\./${pkg}/" || true)"
        if [ -z "$refs" ]; then
            flagged="$flagged $sym"
        fi
    done
    if [ -n "$flagged" ]; then
        echo "  $pkg:$flagged"
    fi
}

echo "== built-but-not-wired candidates (verify each is wired or product-facing) =="
if [ "${1:-}" != "" ]; then
    scan_pkg "$1"
else
    for d in kernel/*/; do
        scan_pkg "${d%/}"
    done
fi
echo "(done — hits are candidates, not confirmed defects)"
