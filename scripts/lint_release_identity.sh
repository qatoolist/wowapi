#!/bin/sh
# Keep the clean-line module, release policy, generated imports, and active
# documentation on one identity. Usage: lint_release_identity.sh [repo-root].
set -eu

root="${1:-.}"
want_module="github.com/qatoolist/wowapi"
want_bootstrap="v1.2.0"
fail=0

module=$(sed -n 's/^module[[:space:]][[:space:]]*//p' "$root/go.mod" | head -n 1)
if [ "$module" != "$want_module" ]; then
  echo "release identity: go.mod module is '$module', want '$want_module'" >&2
  fail=1
fi

if ! grep -q "ModulePath[[:space:]]*=[[:space:]]*\"$want_module\"" "$root/internal/buildinfo/buildinfo.go"; then
  echo "release identity: buildinfo.ModulePath does not match $want_module" >&2
  fail=1
fi

if ! grep -q '"bootstrap_tag"[[:space:]]*:[[:space:]]*"v1.2.0"' "$root/ci/release-line.json"; then
  echo "release identity: ci/release-line.json does not declare bootstrap v1.2.0" >&2
  fail=1
fi

bad_imports=$(grep -RIn --include='*.go' --include='*.tmpl' --include='go.mod' \
  'github\.com/qatoolist/wowapi/v2\(/\|"\|[[:space:]]\)' \
  "$root/app" "$root/adapters" "$root/cmd" "$root/foundation" "$root/internal" \
  "$root/kernel" "$root/module" "$root/testkit" "$root/go.mod" 2>/dev/null || true)
if [ -n "$bad_imports" ]; then
  echo "release identity: stale /v2 framework import(s):" >&2
  echo "$bad_imports" | sed 's/^/  /' >&2
  fail=1
fi

for doc in README.md CHANGELOG.md docs/SRS.md docs/operations/upgrade-and-deprecation-policy.md; do
  if ! grep -q "$want_bootstrap" "$root/$doc"; then
    echo "release identity: $doc does not name clean bootstrap $want_bootstrap" >&2
    fail=1
  fi
done

exit "$fail"
