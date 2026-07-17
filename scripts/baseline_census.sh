#!/bin/sh
# Regenerate the catalog-equivalence census (migrations/baseline/census-reference.txt).
# Applies a migration set to a throwaway database and inventories the
# framework-owned schema objects, normalizing environment-owned details
# (owner, extension-provided functions, OIDs, ACL/dump ordering) so a future
# "equal counts but different objects" change cannot false-pass.
#
# Usage: scripts/baseline_census.sh [ADMIN_DSN]
#   ADMIN_DSN defaults to the local compose DSN. A scratch DB
#   wowapi_baseline_census is created and dropped.
set -eu
ADMIN="${1:-postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable}"
SCRATCH="wowapi_baseline_census"
REF="$(printf '%s' "$ADMIN" | sed -E "s#/[^/?]+(\?|\$)#/${SCRATCH}\1#")"

psql "$ADMIN" -c "DROP DATABASE IF EXISTS $SCRATCH" >/dev/null
psql "$ADMIN" -c "CREATE DATABASE $SCRATCH" >/dev/null
trap 'psql "$ADMIN" -c "DROP DATABASE IF EXISTS $SCRATCH" >/dev/null 2>&1 || true' EXIT

DATABASE_URL="$REF" go run ./internal/tools/migrate >/dev/null

psql "$REF" -tA <<'SQL'
SELECT 'TABLE ' || table_name FROM information_schema.tables WHERE table_schema='public' AND table_name NOT LIKE 'goose_%' ORDER BY 1;
SELECT 'COL ' || table_name||'.'||column_name||' '||data_type||' null='||is_nullable||' def='||coalesce(column_default,'-') FROM information_schema.columns WHERE table_schema='public' AND table_name NOT LIKE 'goose_%' ORDER BY 1;
SELECT 'CONSTRAINT ' || conrelid::regclass||' '||conname||' '||contype FROM pg_constraint WHERE connamespace='public'::regnamespace ORDER BY 1;
SELECT 'INDEX ' || indexname||' '||indexdef FROM pg_indexes WHERE schemaname='public' AND tablename NOT LIKE 'goose_%' ORDER BY 1;
SELECT 'RLS ' || relname||' enabled='||relrowsecurity||' forced='||relforcerowsecurity FROM pg_class WHERE relnamespace='public'::regnamespace AND relkind='r' AND relname NOT LIKE 'goose_%' ORDER BY 1;
SELECT 'POLICY ' || tablename||' '||policyname||' '||cmd||' '||coalesce(regexp_replace(qual,'\s+',' ','g'),'-')||' / '||coalesce(regexp_replace(with_check,'\s+',' ','g'),'-') FROM pg_policies WHERE schemaname='public' ORDER BY 1;
SELECT 'FUNC ' || p.proname||'('||pg_get_function_identity_arguments(p.oid)||')' FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='public' ORDER BY 1;
SELECT 'GRANT ' || grantee||' '||privilege_type||' ON '||table_name FROM information_schema.role_table_grants WHERE table_schema='public' AND table_name NOT LIKE 'goose_%' AND grantee IN ('app_rt','app_platform') ORDER BY 1;
SELECT 'EXT ' || extname FROM pg_extension ORDER BY 1;
SQL
