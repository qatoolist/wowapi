-- Semantic framework-owned catalog census (shared by scripts/baseline_census.sh
-- and the mutation-discrimination check). Emits one normalized text line per
-- object. See scripts/baseline_census.sh for the semantics captured.
-- Extensions: name + version + install schema.
SELECT 'EXT ' || e.extname||' v'||e.extversion||' @'||n.nspname
  FROM pg_extension e JOIN pg_namespace n ON n.oid=e.extnamespace ORDER BY 1;

-- Tables (framework-owned, excludes goose bookkeeping).
SELECT 'TABLE ' || c.relname
  FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
 WHERE n.nspname='public' AND c.relkind='r' AND c.relname NOT LIKE 'goose_%' ORDER BY 1;

-- Columns: full type (format_type incl. typmod/precision), nullability, default,
-- identity, generated, collation.
SELECT 'COL ' || c.relname||'.'||a.attname
       ||' '||format_type(a.atttypid, a.atttypmod)
       ||' null='||(NOT a.attnotnull)
       ||' def='||coalesce(pg_get_expr(ad.adbin, ad.adrelid),'-')
       ||' identity='||(CASE a.attidentity::text WHEN '' THEN '-' ELSE a.attidentity::text END)
       ||' generated='||(CASE a.attgenerated::text WHEN '' THEN '-' ELSE a.attgenerated::text END)
       ||' coll='||coalesce((SELECT collname FROM pg_collation WHERE oid=a.attcollation),'-')
  FROM pg_attribute a
  JOIN pg_class c ON c.oid=a.attrelid
  JOIN pg_namespace n ON n.oid=c.relnamespace
  LEFT JOIN pg_attrdef ad ON ad.adrelid=a.attrelid AND ad.adnum=a.attnum
 WHERE n.nspname='public' AND c.relkind='r' AND c.relname NOT LIKE 'goose_%'
   AND a.attnum>0 AND NOT a.attisdropped ORDER BY 1;

-- Constraints: full definition (FK actions, referenced cols, check exprs),
-- validation + deferrability state.
SELECT 'CONSTRAINT ' || conrelid::regclass||' '||conname||' '
       ||regexp_replace(pg_get_constraintdef(oid), '\s+', ' ', 'g')
       ||' validated='||convalidated||' deferrable='||condeferrable||' deferred='||condeferred
  FROM pg_constraint WHERE connamespace='public'::regnamespace ORDER BY 1;

-- Indexes.
SELECT 'INDEX ' || indexname||' '||regexp_replace(indexdef, '\s+', ' ', 'g')
  FROM pg_indexes WHERE schemaname='public' AND tablename NOT LIKE 'goose_%' ORDER BY 1;

-- RLS enable/force per table.
SELECT 'RLS ' || relname||' enabled='||relrowsecurity||' forced='||relforcerowsecurity
  FROM pg_class WHERE relnamespace='public'::regnamespace AND relkind='r'
   AND relname NOT LIKE 'goose_%' ORDER BY 1;

-- Policies: command, permissive/restrictive, roles, USING + WITH CHECK.
SELECT 'POLICY ' || tablename||' '||policyname||' '||cmd
       ||' permissive='||permissive
       ||' roles={'||array_to_string(roles,',')||'}'
       ||' using='||coalesce(regexp_replace(qual,'\s+',' ','g'),'-')
       ||' check='||coalesce(regexp_replace(with_check,'\s+',' ','g'),'-')
  FROM pg_policies WHERE schemaname='public' ORDER BY 1;

-- Functions: FRAMEWORK-OWNED ONLY (excludes extension-provided via pg_depend).
-- Body captured as a stable md5 so a body change is detected without dumping it.
SELECT 'FUNC ' || p.proname||'('||pg_get_function_identity_arguments(p.oid)||')'
       ||' ret='||pg_get_function_result(p.oid)
       ||' lang='||l.lanname
       ||' vol='||p.provolatile::text||' strict='||p.proisstrict||' secdef='||p.prosecdef
       ||' cfg={'||coalesce(array_to_string(p.proconfig,','),'')||'}'
       ||' bodymd5='||md5(coalesce(p.prosrc,''))
  FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace JOIN pg_language l ON l.oid=p.prolang
 WHERE n.nspname='public'
   AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid=p.oid AND d.deptype='e')
 ORDER BY 1;

-- Grants: table + column + sequence + function privileges for the app roles,
-- with grant-option flag; plus schema usage and default privileges.
SELECT 'GRANT.TABLE ' || grantee||' '||privilege_type||' grantopt='||is_grantable||' ON '||table_name
  FROM information_schema.role_table_grants
 WHERE table_schema='public' AND table_name NOT LIKE 'goose_%'
   AND grantee IN ('app_rt','app_platform') ORDER BY 1;
SELECT 'GRANT.COLUMN ' || grantee||' '||privilege_type||' ON '||table_name||'.'||column_name
  FROM information_schema.role_column_grants
 WHERE table_schema='public' AND table_name NOT LIKE 'goose_%'
   AND grantee IN ('app_rt','app_platform') ORDER BY 1;
SELECT 'GRANT.SEQ ' || r.rolname||' '||
       array_to_string(ARRAY(SELECT unnest FROM unnest(ARRAY['USAGE','SELECT','UPDATE'])
         WHERE has_sequence_privilege(r.rolname, c.oid, unnest)),',')||' ON '||c.relname
  FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
  CROSS JOIN (SELECT rolname FROM pg_roles WHERE rolname IN ('app_rt','app_platform')) r
 WHERE n.nspname='public' AND c.relkind='S'
   AND has_sequence_privilege(r.rolname, c.oid, 'USAGE,SELECT,UPDATE') ORDER BY 1;
SELECT 'GRANT.FUNC ' || r.rolname||' EXECUTE ON '||p.proname
  FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
  CROSS JOIN (SELECT rolname FROM pg_roles WHERE rolname IN ('app_rt','app_platform')) r
 WHERE n.nspname='public'
   AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid=p.oid AND d.deptype='e')
   AND has_function_privilege(r.rolname, p.oid, 'EXECUTE') ORDER BY 1;
SELECT 'GRANT.SCHEMA ' || r.rolname||' '||priv
  FROM (SELECT rolname FROM pg_roles WHERE rolname IN ('app_rt','app_platform')) r
  CROSS JOIN (VALUES ('USAGE'),('CREATE')) AS p(priv)
 WHERE has_schema_privilege(r.rolname, 'public', p.priv) ORDER BY 1;
