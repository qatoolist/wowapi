#!/usr/bin/env bash
# find_duplicate_tests.sh — ADVISORY: list Go test names defined in more than one
# file, so you can confirm each pair tests genuinely different things (per-package
# `TestTenantIsolation` etc. is legitimate and common here) rather than copy-paste
# drift. Go rejects same-name funcs within a package, so every hit is cross-package
# and MAY be intentional — this is an attention list, never a hard failure.
#
# When to run: before adding tests, and during the review gate (no-duplicate-tests).
# Read-only. Always exits 0 (advisory).
set -uo pipefail
cd "$(dirname "$0")/.."

echo "== duplicate test-name audit (advisory) =="
dups="$(grep -rhoE --include='*_test.go' '^func (Test|Benchmark|Fuzz)[A-Za-z0-9_]+' . 2>/dev/null \
        | awk '{print $2}' | sort | uniq -d || true)"

if [ -z "$dups" ]; then
    echo "OK: no repeated test/benchmark/fuzz function names."
    exit 0
fi

echo "Repeated test names (each in >1 file) — confirm they test DIFFERENT things:"
for name in $dups; do
    echo "  $name:"
    grep -rn --include='*_test.go' -E "^func ${name}\b" . | sed 's/^/    /'
done
echo "(advisory — cross-package same-names are legal; look for copy-paste drift only)"
exit 0
