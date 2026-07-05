#!/usr/bin/env bash
# Launch the wowapi product-dev box against a product working directory.
#
#   scripts/product-dev.sh /path/to/your/product-dir
#
# Brings up postgres/minio/mailpit, bootstraps a product database + the app_rt/
# app_platform LOGIN roles (so the API runs as a non-superuser with RLS enforced),
# then drops you into an interactive shell in /workspace. See
# docs/operations/product-dev-container.md.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

PRODUCT_DIR_ARG="${1:-}"
if [ -z "$PRODUCT_DIR_ARG" ]; then
  echo "usage: $0 <product-working-dir>" >&2
  exit 2
fi

# Absolute path; create it if missing (an empty dir is fine — you scaffold into it).
mkdir -p "$PRODUCT_DIR_ARG"
PRODUCT_DIR="$(cd "$PRODUCT_DIR_ARG" && pwd)"
export PRODUCT_DIR

# Kept in sync with deployments/product-dev.yaml via the same defaults.
export PRODUCT_DB="${PRODUCT_DB:-wowproduct}"
export APP_RT_PASSWORD="${APP_RT_PASSWORD:-app-local-only}"

COMPOSE=(docker compose -f "$ROOT/deployments/compose.yaml" -f "$ROOT/deployments/product-dev.yaml")

echo "==> product dir : $PRODUCT_DIR"
echo "==> product db  : $PRODUCT_DB"
echo "==> starting services (postgres, minio, mailpit) ..."
"${COMPOSE[@]}" up -d --wait postgres minio mailpit

echo "==> bootstrapping roles (app_rt, app_platform LOGIN) ..."
# app_rt/app_platform are cluster-global and may already exist NOLOGIN (created by
# a prior migration/test run). Force LOGIN + the local password so the API can
# connect as app_rt. WE are the out-of-band ops grant the framework expects.
"${COMPOSE[@]}" exec -T postgres psql -v ON_ERROR_STOP=1 -U wowapi -d postgres <<SQL
DO \$\$ BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='app_rt') THEN
    ALTER ROLE app_rt LOGIN PASSWORD '${APP_RT_PASSWORD}';
  ELSE
    CREATE ROLE app_rt LOGIN PASSWORD '${APP_RT_PASSWORD}';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='app_platform') THEN
    ALTER ROLE app_platform LOGIN PASSWORD '${APP_RT_PASSWORD}';
  ELSE
    CREATE ROLE app_platform LOGIN PASSWORD '${APP_RT_PASSWORD}';
  END IF;
END \$\$;
SQL

echo "==> ensuring product database '$PRODUCT_DB' ..."
if ! "${COMPOSE[@]}" exec -T postgres \
    psql -tAc "SELECT 1 FROM pg_database WHERE datname='${PRODUCT_DB}'" -U wowapi -d postgres | grep -q 1; then
  "${COMPOSE[@]}" exec -T postgres psql -v ON_ERROR_STOP=1 -U wowapi -d postgres \
    -c "CREATE DATABASE ${PRODUCT_DB} OWNER wowapi;"
fi
"${COMPOSE[@]}" exec -T postgres psql -v ON_ERROR_STOP=1 -U wowapi -d postgres \
  -c "GRANT CONNECT ON DATABASE ${PRODUCT_DB} TO app_rt, app_platform;"
# The scaffolded api's platform pool defaults to the runtime DSN + SET ROLE
# app_platform, so app_rt must be a member of app_platform. (In production you'd
# instead give the platform process its own app_platform login via db.platform_dsn.)
"${COMPOSE[@]}" exec -T postgres psql -v ON_ERROR_STOP=1 -U wowapi -d postgres \
  -c "GRANT app_platform TO app_rt;"

echo "==> entering devbox shell (workspace: $PRODUCT_DIR)"
exec "${COMPOSE[@]}" run --rm --service-ports devbox
