#!/usr/bin/env bash
# check_test_skips.sh — surface the "green-but-hollow" risk: tests that t.Skip.
#
# A green suite that skips the meaningful (DB/integration/E2E) tests is not proof.
# This lists every t.Skip / t.Skipf / t.SkipNow site so a reviewer can confirm none
# of them mask real coverage, and reminds you that the authoritative gate runs with
# WOWAPI_REQUIRE_DB=1 (which turns DB-test skips into failures).
#
# When to run: during the review gate, and any time you touch integration tests.
# Read-only. Exit 0 always (informational); prints a count + the sites.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "== t.Skip audit (green-but-hollow guard) =="
# -n line numbers, exclude vendor; testkit's RequireDB()->Fatal path is the correct pattern.
hits="$(grep -rn --include='*_test.go' -E '\bt\.Skip(f|Now)?\(' . 2>/dev/null || true)"

if [ -z "$hits" ]; then
    echo "OK: no t.Skip sites found."
else
    n="$(printf '%s\n' "$hits" | grep -c . || true)"
    echo "$hits"
    echo "---"
    echo "$n skip site(s). Verify NONE hide DB/integration/E2E coverage."
    echo "Reminder: run 'make ci-container' (WOWAPI_REQUIRE_DB=1) — DB tests must FAIL, not skip."
fi
