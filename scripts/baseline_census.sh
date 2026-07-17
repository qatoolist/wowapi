#!/bin/sh
# Emit a deterministic, schema-complete semantic manifest for a freshly
# migrated framework database.  stdout contains manifest lines only; command
# diagnostics are preserved on stderr and retained in a temporary directory on
# failure.
#
# Usage: scripts/baseline_census.sh [ADMIN_DSN]
set -eu

ADMIN="${1:-postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable}"
DIR=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
WORK=$(mktemp -d "${TMPDIR:-/tmp}/wowapi-baseline-census.XXXXXX")
SUFFIX=$(basename "$WORK" | tr -cd 'a-zA-Z0-9' | tail -c 20)
SCRATCH="wowapi_baseline_census_$$_$SUFFIX"
REF=$(printf '%s' "$ADMIN" | sed -E "s#/[^/?]+(\?|$)#/${SCRATCH}\1#")
CREATED=0

cleanup() {
    status=$?
    if [ "$CREATED" -eq 1 ]; then
        psql -X -v ON_ERROR_STOP=1 "$ADMIN" -qAtc "DROP DATABASE IF EXISTS \"$SCRATCH\"" \
            >>"$WORK/cleanup.stdout" 2>>"$WORK/cleanup.stderr" || true
    fi
    if [ "$status" -eq 0 ]; then
        rm -rf "$WORK"
    else
        echo "baseline census failed; diagnostics retained at $WORK" >&2
    fi
}
trap cleanup EXIT
trap 'exit 130' HUP INT TERM

run_step() {
    name=$1
    shift
    if "$@" >"$WORK/$name.stdout" 2>"$WORK/$name.stderr"; then
        if [ -s "$WORK/$name.stderr" ]; then
            sed "s/^/baseline census [$name]: /" "$WORK/$name.stderr" >&2
        fi
        return 0
    fi
    echo "baseline census [$name] failed" >&2
    sed "s/^/  /" "$WORK/$name.stderr" >&2
    return 1
}

run_step create psql -X -v ON_ERROR_STOP=1 "$ADMIN" -qAtc "CREATE DATABASE \"$SCRATCH\""
CREATED=1

if ! DATABASE_URL="$REF" go run ./internal/tools/migrate \
    >"$WORK/migrate.stdout" 2>"$WORK/migrate.stderr"; then
    echo "baseline census [migrate] failed" >&2
    sed 's/^/  /' "$WORK/migrate.stderr" >&2
    exit 1
fi
if [ -s "$WORK/migrate.stderr" ]; then
    sed 's/^/baseline census [migrate]: /' "$WORK/migrate.stderr" >&2
fi

if ! psql -X -v ON_ERROR_STOP=1 -qAt "$REF" -f "$DIR/baseline_census.sql" \
    >"$WORK/census.stdout" 2>"$WORK/census.stderr"; then
    echo "baseline census [catalog] failed" >&2
    sed 's/^/  /' "$WORK/census.stderr" >&2
    exit 1
fi
if [ -s "$WORK/census.stderr" ]; then
    sed 's/^/baseline census [catalog]: /' "$WORK/census.stderr" >&2
fi

LC_ALL=C sort "$WORK/census.stdout"
