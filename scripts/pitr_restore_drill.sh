#!/usr/bin/env bash
# Point-in-time-recovery (PITR) restore drill (roadmap O5, backlog B-5).
#
# The logical drill (scripts/backup_restore_drill.sh) proves a pg_dump ->
# pg_restore round-trip. That does NOT exercise the mechanism a production
# restore actually relies on: a physical base backup + WAL replay to a chosen
# timestamp. This drill rehearses exactly that, end to end and self-contained,
# so the *procedure* (base backup, continuous WAL archiving, recovery_target_time
# replay, promote) is proven in CI/local — not just trusted from the runbook.
#
# It is deliberately isolated from the compose stack: it spins up its OWN
# throwaway postgres:16-alpine primary configured for WAL archiving (the shared
# compose postgres has archive_mode off, as an ephemeral dev DB should). Steps:
#   1. start a primary with archive_mode=on archiving WAL to a scratch dir
#   2. take a physical `pg_basebackup`
#   3. write a row, mark timestamp T (the recovery target), switch WAL
#   4. write a SECOND row AFTER T (the "mistake" to be recovered past)
#   5. restore the base backup into a fresh server with restore_command +
#      recovery_target_time=T, let it replay WAL and promote
#   6. assert the restored DB contains the pre-T row and NOT the post-T row
#
# Production PITR itself is delegated to the managed provider (see
# docs/operations/backup-restore.md and decision D-0080); this drill validates
# that we understand and can execute the recovery procedure against real WAL.
#
#   ./scripts/pitr_restore_drill.sh          # needs docker; no compose stack required
#
# Requires: docker. Re-runnable; all containers/volumes are removed on exit.
set -euo pipefail

IMG="${PITR_PG_IMAGE:-postgres:16-alpine}"
PGPW="pitr-drill-local-only"
PRIMARY="wowapi_pitr_primary_$$"
RESTORE="wowapi_pitr_restore_$$"
WORK="$(mktemp -d -t wowapi_pitr_XXXXXX)"
ARCHIVE="${WORK}/archive"
BASE="${WORK}/base"
ENTRY="${WORK}/restore-entry.sh"
mkdir -p "$ARCHIVE" "$BASE"
# Containers run postgres as uid 70 (alpine) / 999; make the shared dirs writable.
chmod 777 "$WORK" "$ARCHIVE" "$BASE"

cleanup() {
    docker rm -f "$PRIMARY" "$RESTORE" >/dev/null 2>&1 || true
    rm -rf "$WORK"
}
trap cleanup EXIT

# wait_ready blocks until the named container answers pg_isready (or fails).
wait_ready() {
    local name="$1"
    for _ in $(seq 1 30); do
        if docker exec "$name" pg_isready -U postgres >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done
    echo "pitr: FAIL — ${name} did not become ready" >&2
    docker logs "$name" 2>&1 | tail -20 >&2
    return 1
}

# The restore entrypoint: seed PGDATA from the base backup, point recovery at
# the archived WAL + target time, and let postgres replay and promote. Written
# as a mounted file so no nested shell quoting leaks into the drill.
cat > "$ENTRY" <<'ENTRYEOF'
#!/bin/sh
set -e
PGDATA=/var/lib/postgresql/data/pgdata
export PGDATA
rm -rf "$PGDATA"
cp -r /restore "$PGDATA"
chmod 700 "$PGDATA"
{
    echo "restore_command = 'cp /archive/%f %p'"
    echo "recovery_target_time = '${RECOVERY_TARGET_TIME}'"
    echo "recovery_target_action = 'promote'"
} >> "$PGDATA/postgresql.auto.conf"
touch "$PGDATA/recovery.signal"
exec docker-entrypoint.sh postgres
ENTRYEOF
chmod +x "$ENTRY"

echo "pitr: [1/6] start primary (WAL archiving on) — ${IMG}"
docker run -d --name "$PRIMARY" -e POSTGRES_PASSWORD="$PGPW" -v "$ARCHIVE:/archive" "$IMG" \
    -c wal_level=replica \
    -c archive_mode=on \
    -c "archive_command=test ! -f /archive/%f && cp %p /archive/%f" \
    -c max_wal_senders=3 >/dev/null
wait_ready "$PRIMARY"

echo "pitr: [2/6] physical base backup"
docker exec "$PRIMARY" psql -U postgres -qc "CREATE TABLE ledger (id int PRIMARY KEY, note text);" >/dev/null
docker exec -e PGPASSWORD="$PGPW" "$PRIMARY" \
    pg_basebackup -U postgres -D /basebackup -Fp -Xs -c fast >/dev/null 2>&1
docker cp "$PRIMARY:/basebackup/." "$BASE/" >/dev/null

echo "pitr: [3/6] write pre-target row, capture recovery target T, switch WAL"
docker exec "$PRIMARY" psql -U postgres -qc "INSERT INTO ledger VALUES (1, 'pre-target keep');" >/dev/null
docker exec "$PRIMARY" psql -U postgres -qc "CHECKPOINT; SELECT pg_switch_wal();" >/dev/null
sleep 1
TARGET="$(docker exec "$PRIMARY" psql -U postgres -qAtc "SELECT now();")"
echo "pitr:      recovery_target_time = ${TARGET}"
sleep 2

echo "pitr: [4/6] write POST-target row (the change to recover past)"
docker exec "$PRIMARY" psql -U postgres -qc "INSERT INTO ledger VALUES (2, 'post-target DROP');" >/dev/null
docker exec "$PRIMARY" psql -U postgres -qc "SELECT pg_switch_wal(); CHECKPOINT;" >/dev/null
sleep 1
docker rm -f "$PRIMARY" >/dev/null

echo "pitr: [5/6] restore base backup + replay WAL to T, promote"
docker run -d --name "$RESTORE" -e POSTGRES_PASSWORD="$PGPW" -e RECOVERY_TARGET_TIME="$TARGET" \
    -v "$ARCHIVE:/archive:ro" -v "$BASE:/restore:ro" -v "$ENTRY:/restore-entry.sh:ro" \
    --entrypoint /restore-entry.sh "$IMG" >/dev/null
wait_ready "$RESTORE"
# Give recovery a moment to finish replay + promote before querying.
sleep 2

echo "pitr: [6/6] verify — pre-target row present, post-target row absent"
rows="$(docker exec "$RESTORE" psql -U postgres -qAtc "SELECT count(*) FROM ledger;")"
keep="$(docker exec "$RESTORE" psql -U postgres -qAtc "SELECT count(*) FROM ledger WHERE id = 1;")"
dropped="$(docker exec "$RESTORE" psql -U postgres -qAtc "SELECT count(*) FROM ledger WHERE id = 2;")"

fail=0
if [ "$keep" != "1" ]; then
    echo "pitr: FAIL — pre-target row (id=1) missing after recovery (found ${keep})" >&2
    fail=1
fi
if [ "$dropped" != "0" ]; then
    echo "pitr: FAIL — post-target row (id=2) survived recovery to T (found ${dropped}); WAL was replayed past the target" >&2
    fail=1
fi
if [ "$fail" -ne 0 ]; then
    exit 1
fi

echo "pitr: OK — recovered to T with ${rows} row(s): pre-target kept, post-target correctly excluded"
