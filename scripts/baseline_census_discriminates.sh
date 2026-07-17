#!/bin/sh
# Mutation test for the semantic catalog oracle.  Each case compares complete
# sorted manifests and requires changed semantic lines in the expected catalog
# class.  Several cases deliberately preserve line counts, preventing a
# count-only census from passing.
#
# Usage: scripts/baseline_census_discriminates.sh [ADMIN_DSN]
set -eu

ADMIN="${1:-postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable}"
DIR=$(CDPATH='' cd -- "$(dirname "$0")" && pwd)
ROOT=$(CDPATH='' cd -- "$DIR/.." && pwd)
WORK=$(mktemp -d "${TMPDIR:-/tmp}/wowapi-census-discrimination.XXXXXX")
SUFFIX=$(basename "$WORK" | tr -cd 'a-zA-Z0-9' | tail -c 16)
BASE="wowapi_census_base_$$_$SUFFIX"
CURRENT=""
CASE_DSN=""
PROBE_ROLE="wowapi_census_probe_$$_$SUFFIX"
BASE_DSN=$(printf '%s' "$ADMIN" | sed -E "s#/[^/?]+(\?|$)#/${BASE}\1#")
CREATED_BASE=0
FAIL=0
CASE_NO=0

cleanup() {
    status=$?
    if [ -n "$CURRENT" ]; then
        psql -X "$ADMIN" -qAtc "DROP DATABASE IF EXISTS \"$CURRENT\"" \
            >>"$WORK/cleanup.stdout" 2>>"$WORK/cleanup.stderr" || true
    fi
    if [ "$CREATED_BASE" -eq 1 ]; then
        psql -X "$ADMIN" -qAtc "DROP DATABASE IF EXISTS \"$BASE\"" \
            >>"$WORK/cleanup.stdout" 2>>"$WORK/cleanup.stderr" || true
    fi
    # Probe roles are unique to this invocation.  Membership is cluster-wide,
    # so clean it even if a case aborts between mutation and comparison.
    psql -X "$ADMIN" -qAtc "DROP ROLE IF EXISTS \"$PROBE_ROLE\"" \
        >>"$WORK/cleanup.stdout" 2>>"$WORK/cleanup.stderr" || true

    if [ "$status" -eq 0 ] && [ "$FAIL" -eq 0 ]; then
        rm -rf "$WORK"
    else
        echo "census discrimination diagnostics retained at $WORK" >&2
    fi
}
trap cleanup EXIT
trap 'exit 130' HUP INT TERM

admin_sql() {
    label=$1
    sql=$2
    if ! psql -X -v ON_ERROR_STOP=1 "$ADMIN" -qAtc "$sql" \
        >"$WORK/$label.stdout" 2>"$WORK/$label.stderr"; then
        echo "  setup failure [$label]" >&2
        sed 's/^/    /' "$WORK/$label.stderr" >&2
        return 1
    fi
}

db_sql() {
    label=$1
    dsn=$2
    sql=$3
    if ! psql -X -v ON_ERROR_STOP=1 "$dsn" -qAtc "$sql" \
        >"$WORK/$label.stdout" 2>"$WORK/$label.stderr"; then
        echo "  SQL failure [$label]" >&2
        sed 's/^/    /' "$WORK/$label.stderr" >&2
        return 1
    fi
}

census() {
    label=$1
    dsn=$2
    out=$3
    if ! psql -X -v ON_ERROR_STOP=1 -qAt "$dsn" -f "$DIR/baseline_census.sql" \
        >"$WORK/$label.unsorted" 2>"$WORK/$label.stderr"; then
        echo "  census failure [$label]" >&2
        sed 's/^/    /' "$WORK/$label.stderr" >&2
        return 1
    fi
    LC_ALL=C sort "$WORK/$label.unsorted" >"$out"
}

new_case_db() {
    CASE_NO=$((CASE_NO + 1))
    CURRENT="wowapi_census_case_$$_${CASE_NO}_$SUFFIX"
    admin_sql "case_${CASE_NO}_create" "CREATE DATABASE \"$CURRENT\" TEMPLATE \"$BASE\""
    CASE_DSN=$(printf '%s' "$ADMIN" | sed -E "s#/[^/?]+(\?|$)#/${CURRENT}\1#")
}

drop_case_db() {
    admin_sql "case_${CASE_NO}_drop" "DROP DATABASE IF EXISTS \"$CURRENT\"" || true
    CURRENT=""
}

# check NAME SETUP MUTATION EXPECTED_DIFF_REGEX SAME_COUNT
check() {
    name=$1
    setup=$2
    mutation=$3
    expected=$4
    same_count=$5
    if ! new_case_db; then
        FAIL=1
        return
    fi
    dsn=$CASE_DSN
    before="$WORK/case_${CASE_NO}.before"
    after="$WORK/case_${CASE_NO}.after"
    delta="$WORK/case_${CASE_NO}.diff"

    if [ -n "$setup" ] && ! db_sql "case_${CASE_NO}_setup" "$dsn" "$setup"; then
        echo "  FAIL: $name — fixture setup failed"
        FAIL=1
        drop_case_db
        return
    fi
    if ! census "case_${CASE_NO}_before" "$dsn" "$before"; then
        echo "  FAIL: $name — baseline census failed"
        FAIL=1
        drop_case_db
        return
    fi
    if ! db_sql "case_${CASE_NO}_mutation" "$dsn" "$mutation"; then
        echo "  FAIL: $name — mutation failed"
        FAIL=1
        drop_case_db
        return
    fi
    if ! census "case_${CASE_NO}_after" "$dsn" "$after"; then
        echo "  FAIL: $name — mutated census failed"
        FAIL=1
        drop_case_db
        return
    fi

    if diff -u "$before" "$after" >"$delta"; then
        echo "  DISCRIMINATION FAIL: $name — semantic manifest did not change"
        FAIL=1
    elif ! grep -E "^[+-]${expected}" "$delta" >/dev/null; then
        echo "  DISCRIMINATION FAIL: $name — changed, but not in expected class /$expected/"
        sed -n '1,80p' "$delta" | sed 's/^/    /'
        FAIL=1
    elif [ "$same_count" = "same" ] && [ "$(wc -l <"$before" | tr -d ' ')" != "$(wc -l <"$after" | tr -d ' ')" ]; then
        echo "  DISCRIMINATION FAIL: $name — fixture was meant to preserve semantic-line count"
        FAIL=1
    else
        if [ "$same_count" = "same" ]; then
            echo "  ok: $name detected (same-count semantic change)"
        else
            echo "  ok: $name detected"
        fi
    fi
    drop_case_db
}

echo "census discrimination: building one migrated template database"
admin_sql create_base "CREATE DATABASE \"$BASE\""
CREATED_BASE=1
if ! (cd "$ROOT" && DATABASE_URL="$BASE_DSN" go run ./internal/tools/migrate) \
    >"$WORK/migrate.stdout" 2>"$WORK/migrate.stderr"; then
    echo "migration setup failed" >&2
    sed 's/^/  /' "$WORK/migrate.stderr" >&2
    exit 1
fi

echo "census discrimination: semantic mutation matrix"

check "migration checkpoint composite identity" "" \
  "ALTER TABLE migration.backfill_checkpoint DROP CONSTRAINT backfill_checkpoint_pkey; ALTER TABLE migration.backfill_checkpoint ADD CONSTRAINT backfill_checkpoint_pkey PRIMARY KEY (job_id);" \
  '(CONSTRAINT migration\.backfill_checkpoint|INDEX migration\.backfill_checkpoint_pkey)' same

check "ordinary versus partitioned relation kind" \
  "CREATE TABLE public.census_relation_kind (id integer NOT NULL);" \
  "DROP TABLE public.census_relation_kind; CREATE TABLE public.census_relation_kind (id integer NOT NULL) PARTITION BY RANGE (id);" \
  'REL public\.census_relation_kind ' same

check "foreign-table server options" \
  "CREATE EXTENSION postgres_fdw; CREATE SERVER census_foreign_server FOREIGN DATA WRAPPER postgres_fdw; CREATE FOREIGN TABLE public.census_foreign (id integer) SERVER census_foreign_server OPTIONS (schema_name 'public', table_name 'before_table');" \
  "ALTER FOREIGN TABLE public.census_foreign OPTIONS (SET table_name 'after_table');" \
  'REL public\.census_foreign ' same

check "sequence parameters" "" \
  "ALTER SEQUENCE public.audit_anchors_id_seq INCREMENT BY 7;" \
  'SEQUENCE public\.audit_anchors_id_seq ' same

check "view definition" \
  "CREATE VIEW public.census_view AS SELECT 1::integer AS value;" \
  "CREATE OR REPLACE VIEW public.census_view AS SELECT 2::integer AS value;" \
  'VIEW public\.census_view ' same

check "materialized-view definition" \
  "CREATE MATERIALIZED VIEW public.census_matview AS SELECT 1::integer AS value;" \
  "DROP MATERIALIZED VIEW public.census_matview; CREATE MATERIALIZED VIEW public.census_matview AS SELECT 2::integer AS value;" \
  'MATVIEW public\.census_matview ' same

check "trigger enabled state" \
  "CREATE FUNCTION public.census_trigger_fn() RETURNS trigger LANGUAGE plpgsql AS \$body\$ BEGIN RETURN NEW; END \$body\$; CREATE TRIGGER census_trigger BEFORE INSERT ON public.tenants FOR EACH ROW EXECUTE FUNCTION public.census_trigger_fn();" \
  "ALTER TABLE public.tenants DISABLE TRIGGER census_trigger;" \
  'TRIGGER public\.tenants ' same

check "rewrite rule definition" \
  "CREATE TABLE public.census_rule_table (id integer); CREATE RULE census_rule AS ON INSERT TO public.census_rule_table DO INSTEAD NOTHING;" \
  "DROP RULE census_rule ON public.census_rule_table; CREATE RULE census_rule AS ON UPDATE TO public.census_rule_table DO INSTEAD NOTHING;" \
  'RULE public\.census_rule_table ' same

check "enum labels" \
  "CREATE TYPE public.census_enum AS ENUM ('alpha', 'beta');" \
  "ALTER TYPE public.census_enum ADD VALUE 'gamma';" \
  'TYPE public\.census_enum ' same

check "domain constraint" \
  "CREATE DOMAIN public.census_domain AS integer CONSTRAINT census_domain_check CHECK (VALUE > 0);" \
  "ALTER DOMAIN public.census_domain DROP CONSTRAINT census_domain_check; ALTER DOMAIN public.census_domain ADD CONSTRAINT census_domain_check CHECK (VALUE >= 0);" \
  '(TYPE public\.census_domain |CONSTRAINT public\.census_domain )' same

check "range subtype" \
  "CREATE TYPE public.census_range AS RANGE (subtype = integer);" \
  "DROP TYPE public.census_range; CREATE TYPE public.census_range AS RANGE (subtype = bigint);" \
  'TYPE public\.census_range ' same

check "composite attributes" \
  "CREATE TYPE public.census_composite AS (first_value integer);" \
  "ALTER TYPE public.census_composite RENAME ATTRIBUTE first_value TO renamed_value;" \
  'TYPE public\.census_composite ' same

check "column default" "" \
  "ALTER TABLE public.tenants ALTER COLUMN display_name SET DEFAULT '';" \
  'COL public\.tenants\.display_name ' same

check "generated-column state" "" \
  "ALTER TABLE public.tenants ADD COLUMN census_generated integer GENERATED ALWAYS AS (1) STORED;" \
  'COL public\.tenants\.census_generated ' changed

check "index definition" "" \
  "DROP INDEX public.outbox_pending; CREATE INDEX outbox_pending ON public.events_outbox (occurred_at DESC) WHERE dispatch_status = 'pending';" \
  'INDEX public\.outbox_pending ' same

check "RLS force state" "" \
  "ALTER TABLE public.acting_capacities NO FORCE ROW LEVEL SECURITY;" \
  'RLS public\.acting_capacities ' same

check "policy role" "" \
  "ALTER POLICY acting_capacities_tenant_isolation ON public.acting_capacities TO app_rt;" \
  'POLICY public\.acting_capacities ' same

check "FK deferrability" "" \
  "ALTER TABLE public.acting_capacities ALTER CONSTRAINT acting_capacities_party_id_tenant_fkey DEFERRABLE INITIALLY DEFERRED;" \
  'CONSTRAINT public\.acting_capacities ' same

check "function security-definer" "" \
  "ALTER FUNCTION public.app_tenant_id() SECURITY DEFINER;" \
  'FUNC public\.app_tenant_id\(\)' same

check "function parallel attribute" "" \
  "ALTER FUNCTION public.app_actor_id() PARALLEL SAFE;" \
  'FUNC public\.app_actor_id\(\)' same

check "extension installation" "" \
  "CREATE EXTENSION pg_trgm;" \
  'EXT pg_trgm ' changed

check "table ACL grant option" "" \
  "GRANT SELECT ON public.roles TO app_rt WITH GRANT OPTION;" \
  'ACL\.TABLE public\.roles ' same

check "column ACL grant option" "" \
  "GRANT UPDATE (title) ON public.documents TO app_rt WITH GRANT OPTION;" \
  'ACL\.COLUMN public\.documents\.title ' same

check "sequence ACL identity" \
  "GRANT USAGE ON SEQUENCE public.audit_anchors_id_seq TO app_rt;" \
  "REVOKE USAGE ON SEQUENCE public.audit_anchors_id_seq FROM app_rt; GRANT SELECT ON SEQUENCE public.audit_anchors_id_seq TO app_rt;" \
  'ACL\.SEQUENCE public\.audit_anchors_id_seq ' same

check "schema ACL grant option" "" \
  "GRANT USAGE ON SCHEMA public TO app_rt WITH GRANT OPTION;" \
  'ACL\.SCHEMA public ' same

check "default privileges" \
  "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO app_rt;" \
  "ALTER DEFAULT PRIVILEGES IN SCHEMA public REVOKE SELECT ON TABLES FROM app_rt; ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT INSERT ON TABLES TO app_rt;" \
  'ACL\.DEFAULT ' same

check "overloaded-function privilege identity" \
  "CREATE FUNCTION public.census_overload(integer) RETURNS integer LANGUAGE sql IMMUTABLE AS 'SELECT \$1'; CREATE FUNCTION public.census_overload(text) RETURNS text LANGUAGE sql IMMUTABLE AS 'SELECT \$1'; REVOKE ALL ON FUNCTION public.census_overload(integer), public.census_overload(text) FROM PUBLIC; GRANT EXECUTE ON FUNCTION public.census_overload(integer) TO app_rt;" \
  "REVOKE EXECUTE ON FUNCTION public.census_overload(integer) FROM app_rt; GRANT EXECUTE ON FUNCTION public.census_overload(text) TO app_rt;" \
  'ACL\.FUNCTION public\.census_overload\((integer|text)\)' same

check "PUBLIC function privilege" "" \
  "REVOKE EXECUTE ON FUNCTION public.app_actor_id() FROM PUBLIC;" \
  'ACL\.FUNCTION public\.app_actor_id\(\)' changed

# Cluster roles are not cloned with a database.  Use a unique temporary role,
# mutate membership involving app_rt, census it, and restore it immediately.
if new_case_db; then
    dsn=$CASE_DSN
else
    FAIL=1
    dsn=""
fi
if [ -n "$dsn" ]; then
    before="$WORK/case_${CASE_NO}.before"
    after="$WORK/case_${CASE_NO}.after"
    delta="$WORK/case_${CASE_NO}.diff"
    if admin_sql role_create "CREATE ROLE \"$PROBE_ROLE\" NOLOGIN" \
       && census "case_${CASE_NO}_before" "$dsn" "$before" \
       && admin_sql role_grant "GRANT \"$PROBE_ROLE\" TO app_rt" \
       && census "case_${CASE_NO}_after" "$dsn" "$after"; then
        if diff -u "$before" "$after" >"$delta"; then
            echo "  DISCRIMINATION FAIL: role membership — semantic manifest did not change"
            FAIL=1
        elif ! grep -E '^[+-]ROLE\.MEMBER ' "$delta" >/dev/null; then
            echo "  DISCRIMINATION FAIL: role membership — ROLE.MEMBER line did not change"
            FAIL=1
        else
            echo "  ok: role membership detected"
        fi
    else
        echo "  FAIL: role membership fixture failed"
        FAIL=1
    fi
    admin_sql role_revoke "REVOKE \"$PROBE_ROLE\" FROM app_rt" || true
    admin_sql role_drop "DROP ROLE IF EXISTS \"$PROBE_ROLE\"" || true
    drop_case_db
fi

if [ "$FAIL" -eq 0 ]; then
    echo "census oracle: ALL semantic classes discriminate"
else
    echo "census oracle: NOT DISCRIMINATING"
    exit 1
fi
