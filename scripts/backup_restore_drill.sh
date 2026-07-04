#!/usr/bin/env bash
# Backup/restore drill (roadmap O5). Proves the dump → restore round-trip against
# a seeded instance: dumps SRC, restores into a scratch database, and verifies the
# schema and a marker row survived. Run quarterly (and after any backup-tooling
# change). This validates the PROCEDURE; production PITR uses the managed
# provider's snapshots + WAL — see docs/operations/backup-restore.md.
#
#   SRC_URL=postgres://…/wowapi ./scripts/backup_restore_drill.sh
#
# Requires: pg_dump, psql (matching the server major version).
set -euo pipefail

SRC_URL="${SRC_URL:-${DATABASE_URL:-postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable}}"
# Admin/maintenance connection (for CREATE/DROP DATABASE); defaults to the SRC
# server's `postgres` database.
ADMIN_URL="${ADMIN_URL:-$(echo "$SRC_URL" | sed -E 's#/[^/?]+(\?|$)#/postgres\1#')}"
STAMP="drill_$$"
RESTORE_DB="wowapi_restore_${STAMP}"
DUMP_FILE="$(mktemp -t wowapi_drill_XXXXXX.dump)"
MARKER="restore-marker-${STAMP}"

cleanup() {
    rm -f "$DUMP_FILE"
    psql "$ADMIN_URL" -qAtc "DROP DATABASE IF EXISTS ${RESTORE_DB};" >/dev/null 2>&1 || true
    # Remove the marker we inserted into SRC.
    psql "$SRC_URL" -qAtc "DELETE FROM schema_migrations_marker WHERE note = '${MARKER}';" >/dev/null 2>&1 || true
    psql "$SRC_URL" -qAtc "DROP TABLE IF EXISTS schema_migrations_marker;" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "drill: seeding a marker in SRC"
psql "$SRC_URL" -qAtc "CREATE TABLE IF NOT EXISTS schema_migrations_marker (note text);" >/dev/null
psql "$SRC_URL" -qAtc "INSERT INTO schema_migrations_marker (note) VALUES ('${MARKER}');" >/dev/null

echo "drill: pg_dump SRC → ${DUMP_FILE}"
pg_dump --format=custom --no-owner --no-privileges --file="$DUMP_FILE" "$SRC_URL"

echo "drill: create restore target ${RESTORE_DB}"
psql "$ADMIN_URL" -qAtc "DROP DATABASE IF EXISTS ${RESTORE_DB};" >/dev/null
psql "$ADMIN_URL" -qAtc "CREATE DATABASE ${RESTORE_DB};" >/dev/null

RESTORE_URL="$(echo "$SRC_URL" | sed -E "s#/[^/?]+(\?|$)#/${RESTORE_DB}\1#")"
echo "drill: pg_restore → ${RESTORE_DB}"
# Do not abort on non-fatal restore warnings (e.g. a newer client tool emitting a
# GUC the older server ignores, like transaction_timeout). The verify step below
# is authoritative: if the schema or marker did not land, it fails the drill.
pg_restore --no-owner --no-privileges --dbname="$RESTORE_URL" "$DUMP_FILE" \
    || echo "drill: pg_restore reported warnings; verifying the actual outcome"

echo "drill: verify"
tables="$(psql "$RESTORE_URL" -qAtc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';")"
marker="$(psql "$RESTORE_URL" -qAtc "SELECT count(*) FROM schema_migrations_marker WHERE note='${MARKER}';")"
if [ "$tables" -lt 5 ]; then
    echo "FAIL: restored schema has only ${tables} public tables"; exit 1
fi
if [ "$marker" != "1" ]; then
    echo "FAIL: marker row did not survive the restore (found ${marker})"; exit 1
fi

echo "drill: OK — ${tables} tables restored, marker row intact"
