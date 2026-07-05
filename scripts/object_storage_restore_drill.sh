#!/usr/bin/env bash
# Object-storage restore drill (roadmap O5, backlog B-5).
#
# Documents/attachments blobs live in object storage (kernel/storage), keyed by
# document_versions.storage_key. backup-restore.md requires that a restore bring
# the bucket back to a point AT OR AFTER the database point, so every storage_key
# a restored DB row references resolves to a present blob. This drill rehearses
# the object-storage leg of that restore against the local MinIO in the compose
# stack, end to end:
#   1. write a tenant-prefixed blob to the PRIMARY bucket
#   2. back it up (mirror PRIMARY -> BACKUP bucket)
#   3. simulate loss (delete the blob from PRIMARY)
#   4. restore (mirror BACKUP -> PRIMARY)
#   5. assert the blob is back and byte-identical
#   6. assert the referential invariant: a DB-style storage_key resolves to a blob
#
# All work happens in throwaway drill-* buckets, which are removed on exit; it
# does not touch product data. True cross-region replication/versioning is a
# provider capability (see docs/operations/backup-restore.md); this proves the
# restore PROCEDURE and the DB<->blob consistency check.
#
#   ./scripts/object_storage_restore_drill.sh     # needs docker + `make up` (MinIO)
#
# Requires: docker and a running compose MinIO (make up). Re-runnable.
set -euo pipefail

S3_HOST="${OBJSTORE_S3_HOST:-http://minio:9000}"
S3_USER="${OBJSTORE_S3_USER:-wowapi}"
S3_PASS="${OBJSTORE_S3_PASS:-wowapi-local-only}"
MC_IMAGE="${OBJSTORE_MC_IMAGE:-minio/mc:latest}"
STAMP="drill_$$"
PRIMARY="objdrill-primary-${STAMP//_/-}"
BACKUP="objdrill-backup-${STAMP//_/-}"

# Resolve the compose network so the throwaway mc container can reach `minio`.
NET="${OBJSTORE_NETWORK:-}"
if [ -z "$NET" ]; then
    NET="$(docker network ls --format '{{.Name}}' | grep -E 'wowapi(_|-)default' | head -1 || true)"
fi
if [ -z "$NET" ]; then
    echo "objstore: FAIL — could not find the compose network (is the stack up? \`make up\`)." >&2
    echo "objstore:        override with OBJSTORE_NETWORK=<docker network>." >&2
    exit 1
fi

echo "objstore: rehearsing object-storage backup/restore on ${S3_HOST} (net=${NET})"

# The mc round-trip runs inside one throwaway mc container (so the drill needs
# no mc on the host). It is written to a temp file and mounted — mounting rather
# than piping on stdin is what makes the container's step-by-step stdout reach
# the host reliably. The script is self-checking and exits non-zero on any
# mismatch, which propagates out through `docker run`.
MC_SCRIPT="$(mktemp -t wowapi_objdrill_XXXXXX.sh)"
trap 'rm -f "$MC_SCRIPT"' EXIT
cat > "$MC_SCRIPT" <<'MCEOF'
set -e

cleanup() {
    mc rb --force "src/${PRIMARY}" >/dev/null 2>&1 || true
    mc rb --force "src/${BACKUP}"  >/dev/null 2>&1 || true
}
trap cleanup EXIT

KEY="tenant-11111111/documents/doc-1/v1.bin"
CONTENT="wowapi-object-storage-drill-payload-$$"

echo "objstore: [1/5] write blob to primary bucket ${PRIMARY}"
mc mb -p "src/${PRIMARY}" >/dev/null
mc mb -p "src/${BACKUP}"  >/dev/null
printf '%s' "$CONTENT" | mc pipe "src/${PRIMARY}/${KEY}" >/dev/null

echo "objstore: [2/5] back up (mirror primary -> backup)"
mc mirror --overwrite --quiet "src/${PRIMARY}" "src/${BACKUP}" >/dev/null

echo "objstore: [3/5] simulate loss (delete blob from primary)"
mc rm "src/${PRIMARY}/${KEY}" >/dev/null
if mc stat "src/${PRIMARY}/${KEY}" >/dev/null 2>&1; then
    echo "objstore: FAIL — blob still present after simulated loss" >&2
    exit 1
fi

echo "objstore: [4/5] restore (mirror backup -> primary)"
mc mirror --overwrite --quiet "src/${BACKUP}" "src/${PRIMARY}" >/dev/null

echo "objstore: [5/5] verify restored blob is byte-identical, storage_key resolves"
GOT="$(mc cat "src/${PRIMARY}/${KEY}")"
if [ "$GOT" != "$CONTENT" ]; then
    echo "objstore: FAIL — restored blob differs (got '${GOT}')" >&2
    exit 1
fi
# Referential invariant: the storage_key a restored DB row would hold resolves.
if ! mc stat "src/${PRIMARY}/${KEY}" >/dev/null 2>&1; then
    echo "objstore: FAIL — storage_key ${KEY} does not resolve after restore" >&2
    exit 1
fi

echo "objstore: OK — blob restored byte-identical and storage_key resolves"
MCEOF

docker run --rm --network "$NET" \
    -e MC_HOST_src="http://${S3_USER}:${S3_PASS}@${S3_HOST#http://}" \
    -e PRIMARY="$PRIMARY" -e BACKUP="$BACKUP" \
    -v "$MC_SCRIPT:/objdrill.sh:ro" \
    --entrypoint sh "$MC_IMAGE" /objdrill.sh

echo "objstore: drill passed"
