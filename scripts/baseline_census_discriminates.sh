#!/bin/sh
# Negative test for the census oracle: for each object-semantic class, apply a
# minimal mutation and prove the census manifest CHANGES. A census that cannot
# detect an FK-action / policy-role / function-security / grant / extension /
# generated-column change is not a valid squash acceptance oracle.
#
# Usage: scripts/baseline_census_discriminates.sh [ADMIN_DSN]
set -eu
ADMIN="${1:-postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable}"
DIR="$(dirname "$0")"
fail=0

census() { psql "$1" -tA -f "$DIR/baseline_census.sql" | grep -E '^(EXT|TABLE|COL|CONSTRAINT|INDEX|RLS|POLICY|FUNC|GRANT)' | sort; }

check() { # name, mutation-SQL
  name="$1"; mut="$2"
  scratch="wowapi_census_disc_$$_$(echo "$name" | tr -cd 'a-z0-9')"
  ref="$(printf '%s' "$ADMIN" | sed -E "s#/[^/?]+(\?|\$)#/${scratch}\1#")"
  psql "$ADMIN" -c "DROP DATABASE IF EXISTS $scratch" >/dev/null
  psql "$ADMIN" -c "CREATE DATABASE $scratch" >/dev/null
  DATABASE_URL="$ref" go run ./internal/tools/migrate >/dev/null
  a=$(census "$ref")
  psql "$ref" -c "$mut" >/dev/null
  b=$(census "$ref")
  psql "$ADMIN" -c "DROP DATABASE IF EXISTS $scratch" >/dev/null
  if [ "$a" = "$b" ]; then echo "  DISCRIMINATION FAIL: $name — census did not change"; fail=1
  else echo "  ok: $name detected"; fi
}

echo "census discrimination:"
check "fk-deferrability"  "ALTER TABLE acting_capacities ALTER CONSTRAINT acting_capacities_party_id_tenant_fkey DEFERRABLE INITIALLY DEFERRED;"
check "policy-role"       "ALTER POLICY acting_capacities_tenant_isolation ON acting_capacities TO app_rt;"
check "function-secdef"   "ALTER FUNCTION app_tenant_id() SECURITY DEFINER;"
check "grant-added"       "GRANT SELECT ON tenants TO app_rt;"
check "extension-added"   "CREATE EXTENSION IF NOT EXISTS pg_trgm;"
check "generated-column"  "ALTER TABLE tenants ADD COLUMN census_probe int GENERATED ALWAYS AS (1) STORED;"

[ "$fail" -eq 0 ] && echo "census oracle: ALL classes discriminate" || { echo "census oracle: NOT DISCRIMINATING"; exit 1; }
