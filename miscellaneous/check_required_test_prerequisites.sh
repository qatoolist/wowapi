#!/usr/bin/env bash
# Proves authoritative DB/S3 gates fail closed when their dependencies are absent.
set -euo pipefail
cd "$(dirname "$0")/.."

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

expect_failure() {
    name="$1"
    expected="$2"
    shift 2
    log="$tmp/$name.log"
    set +e
    "$@" >"$log" 2>&1
    status=$?
    set -e
    cat "$log"
    if [ "$status" -eq 0 ]; then
        echo "ERROR: $name prerequisite fixture passed; required dependency was silently skipped" >&2
        return 1
    fi
    if ! grep -Fq "$expected" "$log"; then
        echo "ERROR: $name failed without actionable diagnosis: expected '$expected'" >&2
        return 1
    fi
    echo "PASS: $name exited $status with actionable fail-closed diagnosis"
}

expect_failure db "WOWAPI_REQUIRE_DB" \
    env DATABASE_URL= WOWAPI_TEST_DSN= WOWAPI_REQUIRE_DB=1 \
    go test ./internal/tools/tenantfk -run '^TestScannerEnumerateFixture$' -count=1

expect_failure s3 "WOWAPI_REQUIRE_S3=1 but S3/minio unreachable" \
    env S3_TEST_ENDPOINT=127.0.0.1:1 WOWAPI_REQUIRE_S3=1 \
    go test ./adapters/storage/s3 -run '^TestContract_S3$' -count=1
