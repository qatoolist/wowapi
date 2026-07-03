#!/bin/sh
# Boundary lint (blueprint 04 §1 import law + 00 §5 vocabulary rule; D-0005).
# Runs on the framework repo; `wowapi lint boundaries` (Phase 10) reuses these
# rules for product repos. POSIX sh; depends only on go + grep + awk.
#
# Production imports and test imports are checked separately (review finding
# ARCH-2/ARCH-3): production code must honor the full import law; _test.go
# files may additionally import testkit.
set -eu

MOD=github.com/qatoolist/wowapi
fail=0

prod="$(mktemp)"
tsts="$(mktemp)"
trap 'rm -f "$prod" "$tsts"' EXIT

# "<import-path>: <imports>" — production vs test import sets, kept separate.
go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./... > "$prod"
go list -f '{{.ImportPath}}: {{join .TestImports " "}} {{join .XTestImports " "}}' ./... > "$tsts"

# check_rule <imports-file> <package-prefix> <forbidden-prefix> <reason>
# Path-segment-aware prefix matching (review finding ARCH-1): "kernel"
# matches kernel and kernel/config, never a future kernelx sibling.
check_rule() {
  file="$1"; prefix="$2"; forbidden="$3"; reason="$4"
  bad=$(awk -v p="$MOD/$prefix" -v f="$MOD/$forbidden" '
    $1 == p":" || index($1, p"/") == 1 {
      for (i = 2; i <= NF; i++) if ($i == f || index($i, f"/") == 1) print $1 " imports " $i
    }' "$file" || true)
  if [ -n "$bad" ]; then
    echo "BOUNDARY VIOLATION ($reason):"
    echo "$bad" | sed 's/:$//' | sed 's/^/  /'
    fail=1
  fi
}

# Production import law (prod imports only — test imports handled below).
for f in module app adapters testkit examples internal/testmodules; do
  check_rule "$prod" "kernel" "$f" "kernel must not import $f"
done
for f in app adapters testkit examples internal/testmodules; do
  check_rule "$prod" "module" "$f" "module must not import $f"
done
for f in module app testkit examples internal/testmodules; do
  check_rule "$prod" "adapters" "$f" "adapters must not import $f"
done
for f in testkit examples internal/testmodules; do
  check_rule "$prod" "app" "$f" "app must not import $f"
done
for f in testkit examples internal/testmodules; do
  check_rule "$prod" "cmd" "$f" "cmd must not import $f"
done

for f in module testkit examples internal/testmodules; do
  check_rule "$prod" "internal/cli" "$f" "internal/cli must not import $f"
done

# HARD rule: no production package imports testkit (test imports are fine).
bad=$(awk -v m="$MOD" '
  {
    self = $1; sub(/:$/, "", self)
    if (self == m"/testkit" || index(self, m"/testkit/") == 1) next
    for (i = 2; i <= NF; i++)
      if ($i == m"/testkit" || index($i, m"/testkit/") == 1) print self
  }' "$prod" | sort -u || true)
if [ -n "$bad" ]; then
  echo "BOUNDARY VIOLATION (production code imports testkit):"
  echo "$bad" | sed 's/^/  /'
  fail=1
fi

# Test imports may use testkit, but must still respect the layering law for
# everything else (e.g. kernel tests must not reach into app).
for f in module app adapters; do
  check_rule "$tsts" "kernel" "$f" "kernel tests must not import $f"
done
check_rule "$tsts" "module" "app" "module tests must not import app"

# Vocabulary denylist: product-domain nouns must not enter framework code.
# Word-boundary match; intentionally omits over-generic words (building, wing,
# flat, member) which are covered by review + the Phase 5 AST lint (D-0009).
DENY='society|housing|chairman|treasurer|defaulter|conveyance|redevelopment|agm|maintenance_bill'
hits=$(grep -rniE "\\b($DENY)\\b" kernel module app cmd adapters 2>/dev/null | grep -v '_test.go:' || true)
if [ -n "$hits" ]; then
  echo "VOCABULARY VIOLATION (product-domain terms in framework code):"
  echo "$hits" | sed 's/^/  /'
  fail=1
fi

# Secret.Reveal() call sites: only adapters/, app/, and tests may reveal.
reveals=$(grep -rn '\.Reveal()' kernel module cmd 2>/dev/null | grep -v '_test.go:' | grep -v 'kernel/config/secret.go' || true)
if [ -n "$reveals" ]; then
  echo "SECRET VIOLATION (Reveal() outside adapters/app/tests):"
  echo "$reveals" | sed 's/^/  /'
  fail=1
fi

if [ "$fail" -ne 0 ]; then
  echo "boundary lint: FAILED"
  exit 1
fi
echo "boundary lint: OK"
