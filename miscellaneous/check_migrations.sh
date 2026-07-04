#!/usr/bin/env bash
# check_migrations.sh — audit the kernel migration ledger for the traps that have
# bitten this project (unregistered files, missing Down blocks, numbering gaps).
#
# What it checks:
#   1. every migrations/NNNNN_*.sql is registered in migrations/migrations_test.go
#   2. every migration has BOTH `-- +goose Up` and `-- +goose Down` markers
#      (a missing Down breaks the reversibility drill — already caught the 00010 bug)
#   3. migration numbers are contiguous (no gaps / duplicates)
#
# When to run: after adding or editing any migration, before `make ci-container`.
# Read-only; never modifies files. Exit 0 = clean, 1 = issues found.
set -euo pipefail
cd "$(dirname "$0")/.."

mig_dir="migrations"
reg_file="migrations/migrations_test.go"
fail=0

echo "== migration ledger audit =="

prev=0
for f in "$mig_dir"/[0-9]*.sql; do
    base="$(basename "$f")"
    num=$((10#${base%%_*}))

    # 1. registered in the test ledger
    if ! grep -q "\"$base\"" "$reg_file"; then
        echo "FAIL: $base is NOT registered in $reg_file (add it to expectedFiles)"
        fail=1
    fi
    # 2. Up + Down markers
    grep -q -- '-- +goose Up' "$f"   || { echo "FAIL: $base missing '-- +goose Up'"; fail=1; }
    grep -q -- '-- +goose Down' "$f" || { echo "FAIL: $base missing '-- +goose Down' (reversibility drill will fail)"; fail=1; }
    # 3. contiguity
    if [ "$num" -ne $((prev + 1)) ]; then
        echo "FAIL: numbering gap/dup at $base (expected $((prev + 1)), got $num)"
        fail=1
    fi
    prev=$num
done

if [ "$fail" -eq 0 ]; then
    echo "OK: $prev migrations, all registered, reversible, contiguous."
fi
exit "$fail"
