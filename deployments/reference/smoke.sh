#!/usr/bin/env bash
# S7 reference-stack smoke test: assert the assumed security-header posture is
# present on a running deployment. Point BASE at the reference nginx (or the api
# directly, which sets the same headers in-process via httpx.SecureHeaders).
#
#   BASE=https://localhost ./deployments/reference/smoke.sh
#
# Exits non-zero if any required header is missing. Intended as a deploy-time
# gate and a quarterly drill; the in-process posture is unit-tested in CI
# (kernel/httpx/edge_test.go), so this covers the proxy/TLS wiring.
set -euo pipefail

BASE="${BASE:-http://localhost:8080}"
# Use /readyz — always present, Public, no auth needed.
URL="${BASE%/}/readyz"

echo "smoke: GET $URL"
headers="$(curl -fsS -k -D - -o /dev/null "$URL")"

require() {
    local name="$1" pattern="$2"
    if ! grep -iqE "^${name}:.*${pattern}" <<<"$headers"; then
        echo "FAIL: missing/incorrect header ${name} (want ~ ${pattern})"
        echo "--- headers seen ---"; echo "$headers"
        exit 1
    fi
    echo "ok: ${name}"
}

require "X-Content-Type-Options" "nosniff"
require "X-Frame-Options"        "DENY"
require "Content-Security-Policy" "frame-ancestors 'none'"
require "Referrer-Policy"        "no-referrer"
# HSTS is set by the app and/or the proxy; required on any TLS deployment.
if [[ "$BASE" == https://* ]]; then
    require "Strict-Transport-Security" "max-age="
fi

echo "smoke: all required security headers present"
