#!/usr/bin/env bash
# Runs an opt-in seeded race and succeeds only when Go's race detector reports it.
set -euo pipefail
cd "$(dirname "$0")/.."

log="$(mktemp)"
trap 'rm -f "$log"' EXIT
set +e
go test -race -tags=wowapi_race_fixture ./internal/verificationfixtures/racefixture -run '^TestSeededDataRace$' -count=1 >"$log" 2>&1
status=$?
set -e
cat "$log"

if [ "$status" -eq 0 ]; then
    echo "ERROR: seeded data-race fixture passed under -race" >&2
    exit 1
fi
if ! grep -Fq "DATA RACE" "$log"; then
    echo "ERROR: seeded fixture failed, but the race detector did not report DATA RACE" >&2
    exit 1
fi

echo "PASS: seeded data-race fixture was detected (expected go test exit $status)"
