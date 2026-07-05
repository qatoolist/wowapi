#!/usr/bin/env bash
# Migration reversibility drill WITH SCHEMA-SNAPSHOT DIFFING (roadmap O2, backlog B-4).
#
# The Go drill (migrations/TestIntegrationMigrationsReversible) proves the
# goose-version round-trip: up -> down-to-0 -> up returns to the same head
# *version* and one sentinel table reappears. That alone misses migrations that
# are asymmetric at the SCHEMA level -- a Down that drops a table but not its
# index/policy/function, or leaves a default/constraint behind. goose is happy
# (versions match) but the physical schema has drifted.
#
# This drill closes that gap. On a throwaway database it:
#   1. migrates straight up to head        -> captures the CLEAN schema snapshot
#   2. rolls every migration down to 0
#   3. migrates up to head again           -> captures the ROUND-TRIP snapshot
#   4. DIFFS the two normalized snapshots; ANY difference fails the drill.
#
# The snapshot is `pg_dump --schema-only --no-owner --no-privileges`, normalized
# (comments / session GUCs / pg18 \restrict guard tokens stripped) so only real
# DDL is compared. pg_dump emits objects in a stable type+name order, so a clean
# up and a round-trip up are byte-identical unless a Down leg is asymmetric.
#
#   DATABASE_URL=postgres://…/wowapi ./scripts/migration_reversibility_drill.sh
#
# Requires: pg_dump + psql (a client >= the server major version), and `go`
# (to drive internal/tools/migrate up|reset). Safe to re-run; the scratch
# database is dropped on exit. NEVER point this at a production database:
# `migrate reset` rolls every migration back to 0.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

SRC_URL="${DATABASE_URL:-postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable}"
# Admin/maintenance connection (for CREATE/DROP DATABASE): the SRC server's
# `postgres` database.
ADMIN_URL="$(echo "$SRC_URL" | sed -E 's#/[^/?]+(\?|$)#/postgres\1#')"
SCRATCH_DB="wowapi_revdrill_$$"
SCRATCH_URL="$(echo "$SRC_URL" | sed -E "s#/[^/?]+(\?|$)#/${SCRATCH_DB}\1#")"

SNAP_CLEAN="$(mktemp -t wowapi_revdrill_clean_XXXXXX.sql)"
SNAP_ROUND="$(mktemp -t wowapi_revdrill_round_XXXXXX.sql)"

cleanup() {
    rm -f "$SNAP_CLEAN" "$SNAP_ROUND"
    psql "$ADMIN_URL" -qAtc "DROP DATABASE IF EXISTS ${SCRATCH_DB};" >/dev/null 2>&1 || true
}
trap cleanup EXIT

# migrate runs internal/tools/migrate against the scratch DB in the given mode
# (up|reset). Kept as a function so PATH/module resolution stays in REPO_ROOT.
migrate() {
    ( cd "$REPO_ROOT" && DATABASE_URL="$SCRATCH_URL" go run ./internal/tools/migrate "$1" )
}

# snapshot writes a deterministic, normalized schema-only dump of the scratch DB
# to $1. Stripped: psql meta/comments (^--), blank lines, session GUCs (SET /
# set_config), and the pg18 client's \restrict/\unrestrict guard lines (they
# carry a random token that would otherwise be a false diff every run).
snapshot() {
    pg_dump --schema-only --no-owner --no-privileges "$SCRATCH_URL" \
        | grep -vE '^--|^$|^SET |^SELECT pg_catalog\.set_config|^\\restrict |^\\unrestrict ' \
        > "$1"
}

echo "revdrill: create throwaway database ${SCRATCH_DB}"
psql "$ADMIN_URL" -qAtc "DROP DATABASE IF EXISTS ${SCRATCH_DB};" >/dev/null
psql "$ADMIN_URL" -qAtc "CREATE DATABASE ${SCRATCH_DB};" >/dev/null

echo "revdrill: [1/3] migrate up (clean) -> snapshot"
migrate up
snapshot "$SNAP_CLEAN"
clean_lines="$(wc -l < "$SNAP_CLEAN" | tr -d ' ')"
echo "revdrill:      clean schema = ${clean_lines} DDL lines"

echo "revdrill: [2/3] migrate reset (down to 0)"
migrate reset

echo "revdrill: [3/3] migrate up again -> snapshot"
migrate up
snapshot "$SNAP_ROUND"

echo "revdrill: diff clean vs up->down->up schema"
if diff -u "$SNAP_CLEAN" "$SNAP_ROUND"; then
    echo "revdrill: OK — up->down->up schema is byte-identical to a clean up (${clean_lines} DDL lines)"
else
    echo "revdrill: FAIL — schema drift after up->down->up (see diff above)." >&2
    echo "revdrill:        A Down leg is asymmetric: it does not exactly reverse its Up" >&2
    echo "revdrill:        (a table/index/policy/function/default/constraint was left behind or not recreated)." >&2
    exit 1
fi
