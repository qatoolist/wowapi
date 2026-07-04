#!/usr/bin/env bash
# review_gate.sh — the mechanical half of the Independent Review Gate, in one run.
#
# It runs the fast, deterministic project checks so a reviewer (human or AI) can
# focus on judgement (requirement coverage, wiring, correctness). It does NOT
# replace the fresh-reviewer pass or the authoritative gate; it front-loads the
# things a machine can verify. See docs/working/quality-gate-checklist.md.
#
# Checks: gofmt, go vet, boundary lint, migration ledger, test skips, duplicate
# tests, doc overclaims, stray generation artifacts, and (optionally) the full
# authoritative gate with --full.
#
# Usage:   miscellaneous/review_gate.sh          # fast checks
#          miscellaneous/review_gate.sh --full   # + make ci-container (needs infra)
# Read-only except that --full builds. Exit 0 = all clean, 1 = something to fix.
set -uo pipefail
cd "$(dirname "$0")/.."
here="miscellaneous"
rc=0
step() { echo; echo "### $1"; }

step "gofmt (formatting)"
unformatted="$(gofmt -l kernel internal app module adapters testkit migrations 2>/dev/null || true)"
if [ -n "$unformatted" ]; then echo "FAIL: gofmt needed:"; echo "$unformatted"; rc=1; else echo "OK"; fi

step "go vet"
if go vet ./... >/tmp/rg_vet.log 2>&1; then echo "OK"; else echo "FAIL:"; tail -20 /tmp/rg_vet.log; rc=1; fi

step "boundary lint (import law)"
if sh scripts/lint_boundaries.sh >/tmp/rg_bl.log 2>&1; then tail -1 /tmp/rg_bl.log; else echo "FAIL:"; tail -20 /tmp/rg_bl.log; rc=1; fi

step "migration ledger"
sh "$here/check_migrations.sh" || rc=1

step "test skips (green-but-hollow guard)"
sh "$here/check_test_skips.sh" || true

step "duplicate tests (advisory)"
sh "$here/find_duplicate_tests.sh" || true

step "doc overclaims"
sh "$here/check_overclaims.sh" || true

step "stray generation artifacts (leaked tool-call tags in generated docs)"
# Leaked closing tags from Write/tool calls (</content>, </invoke>, </parameter>,
# </...>) that occasionally slip into generated markdown. Scan docs only —
# the miscellaneous/ scripts reference these strings on purpose to detect them.
stray="$(grep -rlE '</(content|invoke|parameter|antml:[a-z]+)>' --include='*.md' docs 2>/dev/null || true)"
if [ -n "$stray" ]; then echo "FAIL: stray tool-call tags in:"; echo "$stray"; rc=1; else echo "OK"; fi

step "stray committed binaries"
bins="$(git ls-files 2>/dev/null | grep -E '(^|/)(bin/|.*\.test$|.*\.dump$)' || true)"
if [ -n "$bins" ]; then echo "WARN: tracked binaries/dumps:"; echo "$bins"; else echo "OK"; fi

if [ "${1:-}" = "--full" ]; then
    step "authoritative gate (make ci-container — DB tests forced)"
    if make ci-container >/tmp/rg_ci.log 2>&1; then
        fails="$(grep -icE '(^| )FAIL|permitted to log in|REQUIRE_DB is set' /tmp/rg_ci.log || true)"
        skips="$(grep -cE '(^| )SKIP|--- SKIP' /tmp/rg_ci.log || true)"
        oks="$(grep -cE '^ok ' /tmp/rg_ci.log || true)"
        echo "ci-container: FAIL=$fails SKIP=$skips ok=$oks"
        { [ "$fails" -ne 0 ] || [ "$skips" -ne 0 ]; } && rc=1
    else
        echo "FAIL: make ci-container errored"; tail -20 /tmp/rg_ci.log; rc=1
    fi
fi

echo
if [ "$rc" -eq 0 ]; then
    echo "MECHANICAL CHECKS CLEAN — now do the judgement pass (spawn a fresh reviewer, requirement coverage, wiring)."
else
    echo "ISSUES FOUND — fix before proceeding to the judgement pass."
fi
exit "$rc"
