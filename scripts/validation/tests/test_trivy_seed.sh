#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
fixture="$root/scripts/validation/fixtures/security/vulnerable-package-lock.json.seeded"
work=$(mktemp -d "${TMPDIR:-/tmp}/wowapi-trivy-seed.XXXXXX")
trap 'rm -rf "$work"' EXIT HUP INT TERM

cp "$fixture" "$work/package-lock.json"
set +e
trivy fs --scanners vuln --severity CRITICAL,HIGH --exit-code 1 "$work" >"$work/failing.log" 2>&1
failed_status=$?
set -e
if [ "$failed_status" -eq 0 ]; then
  cat "$work/failing.log" >&2
  echo "seeded vulnerable dependency unexpectedly passed Trivy" >&2
  exit 1
fi
if ! grep -q 'CVE-' "$work/failing.log"; then
  cat "$work/failing.log" >&2
  echo "Trivy failed without reporting a seeded CVE" >&2
  exit 1
fi
printf '%s\n' 'seeded vulnerability: rejected as expected'

cat >"$work/package-lock.json" <<'EOF'
{"name":"wowapi-security-fixture","version":"1.0.0","lockfileVersion":2,"packages":{"":{"name":"wowapi-security-fixture","version":"1.0.0"}},"dependencies":{}}
EOF
trivy fs --scanners vuln --severity CRITICAL,HIGH --exit-code 1 "$work" >"$work/passing.log" 2>&1
printf '%s\n' 'removed vulnerability: accepted as expected'
