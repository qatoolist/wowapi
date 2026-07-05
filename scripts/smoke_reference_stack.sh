#!/usr/bin/env bash
# B-7 / CA-6 reference-stack header smoke.
#
# Scaffolds a product with `wowapi init`, builds its linux static api/migrate
# binaries, stands up postgres + migrate + api + the reference nginx (TLS) via
# deployments/reference/smoke-compose.yaml, then runs deployments/reference/smoke.sh
# THROUGH nginx — proving the security-header posture is delivered by the reverse
# proxy over TLS, not just in-process (which kernel/httpx/edge_test.go unit-tests).
#
# Requires: go, docker + compose, openssl. Exits non-zero on any failure.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"
COMPOSE_FILE="deployments/reference/smoke-compose.yaml"
HTTPS_PORT="${SMOKE_HTTPS_PORT:-8443}"

for tool in go docker openssl; do
  command -v "$tool" >/dev/null 2>&1 || { echo "smoke: '$tool' is required" >&2; exit 1; }
done
docker compose version >/dev/null 2>&1 || { echo "smoke: 'docker compose' is required" >&2; exit 1; }

WORK="$(mktemp -d)"
PRODUCT="$WORK/product"
BIN="$WORK/bin"
TLS="$WORK/tls"

cleanup() {
  SMOKE_BIN_DIR="$BIN" SMOKE_CFG_DIR="$PRODUCT/configs" SMOKE_TLS_DIR="$TLS" SMOKE_HTTPS_PORT="$HTTPS_PORT" \
    docker compose -f "$COMPOSE_FILE" down -v --remove-orphans >/dev/null 2>&1 || true
  rm -rf "$WORK"
}
trap cleanup EXIT

echo "==> building wowapi CLI"
go build -o "$WORK/wowapi" ./cmd/wowapi

echo "==> scaffolding product (wowapi init)"
mkdir -p "$PRODUCT"
"$WORK/wowapi" init --module smoke.example/app --name app --dir "$PRODUCT" >/dev/null

echo "==> wiring local framework + go mod tidy"
( cd "$PRODUCT" \
    && go mod edit -replace "github.com/qatoolist/wowapi=$REPO_ROOT" \
    && GOFLAGS=-mod=mod go mod tidy )

echo "==> building linux static api + migrate binaries"
mkdir -p "$BIN"
( cd "$PRODUCT" \
    && GOOS=linux CGO_ENABLED=0 GOFLAGS=-mod=mod go build -o "$BIN/api" ./cmd/api \
    && GOOS=linux CGO_ENABLED=0 GOFLAGS=-mod=mod go build -o "$BIN/migrate" ./cmd/migrate )

echo "==> generating self-signed TLS cert (CN=localhost)"
mkdir -p "$TLS"
openssl req -x509 -newkey rsa:2048 -nodes -days 1 \
  -keyout "$TLS/privkey.pem" -out "$TLS/fullchain.pem" -subj "/CN=localhost" >/dev/null 2>&1

echo "==> starting reference stack (postgres + migrate + api + nginx)"
export SMOKE_BIN_DIR="$BIN" SMOKE_CFG_DIR="$PRODUCT/configs" SMOKE_TLS_DIR="$TLS" SMOKE_HTTPS_PORT="$HTTPS_PORT"
docker compose -f "$COMPOSE_FILE" up -d --wait --wait-timeout 180

# --wait releases as soon as nginx (which has no healthcheck) is "running", which
# can be a moment before it binds :443. Poll from the host until TLS is actually
# accepted, so the smoke never races a not-yet-listening proxy (flaky CI).
echo "==> waiting for nginx to accept TLS on :$HTTPS_PORT"
ready=""
for _ in $(seq 1 30); do
  if curl -k -sf -o /dev/null "https://localhost:$HTTPS_PORT/readyz"; then ready=1; break; fi
  sleep 1
done
if [ -z "$ready" ]; then
  echo "smoke: nginx did not accept TLS on :$HTTPS_PORT within 30s" >&2
  docker compose -f "$COMPOSE_FILE" logs --tail 40 nginx api || true
  exit 1
fi

echo "==> smoke: security headers through nginx (TLS) @ https://localhost:$HTTPS_PORT"
BASE="https://localhost:$HTTPS_PORT" "$REPO_ROOT/deployments/reference/smoke.sh"

# The app AND nginx both default HSTS on; nginx must own it authoritatively
# (proxy_hide_header) so the edge emits exactly one header. A missing
# proxy_hide_header would surface as two — assert exactly one.
hsts_count="$(curl -k -sS -D - -o /dev/null "https://localhost:$HTTPS_PORT/readyz" | grep -ic '^strict-transport-security:')"
if [ "$hsts_count" != "1" ]; then
  echo "smoke: expected exactly 1 Strict-Transport-Security header (nginx edge-authoritative), got $hsts_count" >&2
  exit 1
fi
echo "==> HSTS is single (nginx edge-authoritative) ✓"

echo "==> reference-stack smoke PASSED"
