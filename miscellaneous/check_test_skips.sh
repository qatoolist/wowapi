#!/usr/bin/env bash
# check_test_skips.sh — enforce the owner/rationale approval manifest.
#
# The Go AST validator finds t.Skip/t.Skipf/t.SkipNow calls in every *_test.go
# file. A new call, a stale approval, or incomplete owner/rationale metadata is
# a CI failure. Required integration skips must additionally name their
# WOWAPI_REQUIRE_DB/S3 fail-closed guard.
set -euo pipefail
cd "$(dirname "$0")/.."

exec go run ./internal/tools/testskipmanifest \
    -root . \
    -manifest miscellaneous/test-skip-manifest.json
