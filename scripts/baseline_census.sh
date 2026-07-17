#!/bin/sh
# Semantic catalog manifest of framework-owned schema objects — the acceptance
# oracle the squashed 00001_baseline must reproduce object-for-object. Captures
# full object semantics (constraint defs incl. FK actions/deferrability/
# validation, policy roles + permissive mode, function body-hash + return type
# + language + volatility + strictness + security-definer + config, grants
# across tables/columns/sequences/functions/schema + grant-option, extension
# versions + schema, column identity/generated/collation/type) and NORMALIZES
# environment-owned details (database owner, extension-provided functions via
# pg_depend deptype='e', generated OIDs, ACL/dump ordering).
#
# Usage: scripts/baseline_census.sh [ADMIN_DSN]
# Concurrency-safe: a per-invocation scratch DB (PID-suffixed) is created/dropped.
set -eu
ADMIN="${1:-postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable}"
SCRATCH="wowapi_baseline_census_$$"
REF="$(printf '%s' "$ADMIN" | sed -E "s#/[^/?]+(\?|\$)#/${SCRATCH}\1#")"

psql "$ADMIN" -c "DROP DATABASE IF EXISTS $SCRATCH" >/dev/null
psql "$ADMIN" -c "CREATE DATABASE $SCRATCH" >/dev/null
trap 'psql "$ADMIN" -c "DROP DATABASE IF EXISTS $SCRATCH" >/dev/null 2>&1 || true' EXIT

DATABASE_URL="$REF" go run ./internal/tools/migrate >/dev/null

psql "$REF" -tA -f "$(dirname "$0")/baseline_census.sql"
