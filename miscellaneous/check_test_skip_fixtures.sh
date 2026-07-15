#!/usr/bin/env bash
# Negative/positive fixtures for the machine-checked t.Skip approval manifest.
set -euo pipefail
cd "$(dirname "$0")/.."

log="$(mktemp)"
trap 'rm -f "$log"' EXIT
set +e
go run ./internal/tools/testskipmanifest \
    -root internal/tools/testskipmanifest/testdata/unapproved \
    -manifest internal/tools/testskipmanifest/testdata/unapproved/manifest.json >"$log" 2>&1
status=$?
set -e
cat "$log"
if [ "$status" -eq 0 ]; then
    echo "ERROR: unapproved t.Skip fixture passed validation" >&2
    exit 1
fi
if ! grep -Fq "unapproved t.Skip" "$log"; then
    echo "ERROR: unapproved fixture failed without the expected diagnosis" >&2
    exit 1
fi
echo "PASS: unapproved t.Skip fixture was rejected (expected exit $status)"

go run ./internal/tools/testskipmanifest \
    -root internal/tools/testskipmanifest/testdata/approved \
    -manifest internal/tools/testskipmanifest/testdata/approved/manifest.json

echo "PASS: approved t.Skip fixture with owner and rationale was accepted"
