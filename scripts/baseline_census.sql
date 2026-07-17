-- Schema-complete semantic catalog census for the framework-owned database
-- surface.  The output is deliberately line-oriented and lexically sortable:
-- the historical migration chain and the clean baseline must emit identical
-- lines, not merely identical object counts.
--
-- Framework-owned schemas are declared explicitly below.  goose history
-- relations are execution bookkeeping, not application schema, and are the
-- only name-based exclusion.  Extension-owned objects are excluded using
-- pg_depend membership (deptype = 'e'), never name heuristics.

SET search_path = pg_catalog;

CREATE OR REPLACE FUNCTION pg_temp.census_norm(value text)
RETURNS text
LANGUAGE sql
IMMUTABLE
RETURNS NULL ON NULL INPUT
AS $$
    SELECT regexp_replace(btrim(value), E'[\\n\\r\\t ]+', ' ', 'g')
$$;

-- ACL role rendering is portable across databases owned by different login
-- roles.  PUBLIC and framework roles retain their semantic identity; the
-- object owner is normalized because its LOGIN/password are deployment-owned.
CREATE OR REPLACE FUNCTION pg_temp.census_acl_role(role_oid oid, owner_oid oid)
RETURNS text
LANGUAGE sql
STABLE
AS $$
    SELECT CASE
        WHEN role_oid = 0 THEN 'PUBLIC'
        WHEN role_oid = owner_oid THEN 'OWNER'
        ELSE coalesce((SELECT rolname FROM pg_roles WHERE oid = role_oid), 'OID:' || role_oid::text)
    END
$$;

SELECT 'SCHEMASET migration,public';

-- Schema identity and ownership.  ACLs are emitted separately below.
SELECT 'SCHEMA ' || quote_ident(n.nspname) || ' owner=OWNER'
  FROM pg_namespace n
 WHERE n.nspname IN ('public', 'migration')
 ORDER BY 1;

-- Extensions are part of clean-install semantics, including the exact version
-- and installation schema.  Their member objects are excluded from all other
-- sections through pg_depend.
SELECT 'EXT ' || quote_ident(e.extname) || ' version=' || e.extversion
       || ' schema=' || quote_ident(n.nspname)
  FROM pg_extension e
  JOIN pg_namespace n ON n.oid = e.extnamespace
 ORDER BY 1;

-- All framework relation kinds.  Partition, foreign-table, storage, access
-- method, persistence and reloptions semantics live on this line.  Columns,
-- indexes, sequences and view definitions are expanded in dedicated sections.
SELECT 'REL ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' kind=' || CASE c.relkind
            WHEN 'r' THEN 'table'
            WHEN 'p' THEN 'partitioned_table'
            WHEN 'f' THEN 'foreign_table'
            WHEN 'S' THEN 'sequence'
            WHEN 'v' THEN 'view'
            WHEN 'm' THEN 'materialized_view'
          END
       || ' persistence=' || c.relpersistence::text
       || ' access=' || coalesce(am.amname, '-')
       || ' tablespace=' || coalesce(ts.spcname, '-')
       || ' partition_key=' || coalesce(pg_temp.census_norm(pg_get_partkeydef(c.oid)), '-')
       || ' partition_bound=' || coalesce(pg_temp.census_norm(pg_get_expr(c.relpartbound, c.oid, true)), '-')
       || ' parents={' || coalesce((
            SELECT string_agg(quote_ident(pn.nspname) || '.' || quote_ident(pc.relname), ',' ORDER BY pn.nspname, pc.relname)
              FROM pg_inherits i
              JOIN pg_class pc ON pc.oid = i.inhparent
              JOIN pg_namespace pn ON pn.oid = pc.relnamespace
             WHERE i.inhrelid = c.oid
          ), '') || '}'
       || ' options={' || coalesce((
            SELECT string_agg(opt, ',' ORDER BY opt)
              FROM unnest(c.reloptions) opt
          ), '') || '}'
       || ' foreign_server=' || coalesce(quote_ident(fs.srvname), '-')
       || ' foreign_options={' || coalesce((
            SELECT string_agg(o.option_name || '=' || o.option_value, ',' ORDER BY o.option_name)
              FROM pg_options_to_table(ft.ftoptions) o
          ), '') || '}'
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_am am ON am.oid = c.relam
  LEFT JOIN pg_tablespace ts ON ts.oid = c.reltablespace
  LEFT JOIN pg_foreign_table ft ON ft.ftrelid = c.oid
  LEFT JOIN pg_foreign_server fs ON fs.oid = ft.ftserver
 WHERE n.nspname IN ('public', 'migration')
   AND c.relkind IN ('r', 'p', 'f', 'S', 'v', 'm')
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- Columns include view/composite attributes as well as table columns. Physical
-- attnum gaps are deliberately excluded: they are abandoned DDL history, not a
-- named-column contract, and retaining them would force the clean baseline to
-- recreate and drop legacy columns. Stable schema/relation/column identity is
-- the ordering key. Setting search_path to pg_catalog makes user-defined types
-- schema-qualified while preserving built-in type names.
SELECT 'COL ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname) || '.' || quote_ident(a.attname)
       || ' type=' || format_type(a.atttypid, a.atttypmod)
       || ' nullable=' || (NOT a.attnotnull)
       || ' default=' || coalesce(pg_temp.census_norm(pg_get_expr(ad.adbin, ad.adrelid, true)), '-')
       || ' identity=' || CASE a.attidentity::text WHEN '' THEN '-' ELSE a.attidentity::text END
       || ' generated=' || CASE a.attgenerated::text WHEN '' THEN '-' ELSE a.attgenerated::text END
       || ' collation=' || coalesce(quote_ident(cn.nspname) || '.' || quote_ident(coll.collname), '-')
       || ' compression=' || CASE a.attcompression::text WHEN '' THEN '-' ELSE a.attcompression::text END
       || ' storage=' || a.attstorage::text
  FROM pg_attribute a
  JOIN pg_class c ON c.oid = a.attrelid
  JOIN pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_attrdef ad ON ad.adrelid = a.attrelid AND ad.adnum = a.attnum
  LEFT JOIN pg_collation coll ON coll.oid = a.attcollation
  LEFT JOIN pg_namespace cn ON cn.oid = coll.collnamespace
 WHERE n.nspname IN ('public', 'migration')
   AND c.relkind IN ('r', 'p', 'f', 'v', 'm', 'c')
   AND c.relname NOT LIKE 'goose_%'
   AND a.attnum > 0
   AND NOT a.attisdropped
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- Constraints include table and domain constraints, complete definitions,
-- validation/deferral/inheritance state and parent identity.
SELECT 'CONSTRAINT '
       || CASE
            WHEN con.conrelid <> 0 THEN quote_ident(rn.nspname) || '.' || quote_ident(rc.relname)
            WHEN con.contypid <> 0 THEN quote_ident(tn.nspname) || '.' || quote_ident(t.typname)
            ELSE quote_ident(con_n.nspname) || '.-'
          END
       || ' name=' || quote_ident(con.conname)
       || ' type=' || con.contype::text
       || ' def=' || pg_temp.census_norm(pg_get_constraintdef(con.oid, true))
       || ' validated=' || con.convalidated
       || ' deferrable=' || con.condeferrable
       || ' deferred=' || con.condeferred
       || ' local=' || con.conislocal
       || ' inherited=' || con.coninhcount
       || ' noinherit=' || con.connoinherit
       || ' parent=' || coalesce(quote_ident(pc.conname), '-')
  FROM pg_constraint con
  JOIN pg_namespace con_n ON con_n.oid = con.connamespace
  LEFT JOIN pg_class rc ON rc.oid = con.conrelid
  LEFT JOIN pg_namespace rn ON rn.oid = rc.relnamespace
  LEFT JOIN pg_type t ON t.oid = con.contypid
  LEFT JOIN pg_namespace tn ON tn.oid = t.typnamespace
  LEFT JOIN pg_constraint pc ON pc.oid = con.conparentid
 WHERE con_n.nspname IN ('public', 'migration')
   AND (rc.relname IS NULL OR rc.relname NOT LIKE 'goose_%')
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_constraint'::regclass
          AND d.objid = con.oid
          AND d.deptype = 'e'
   )
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = con.conrelid
          AND d.deptype = 'e'
   )
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_type'::regclass
          AND d.objid = con.contypid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- Index definitions plus catalog state that pg_get_indexdef alone omits.
SELECT 'INDEX ' || quote_ident(ins.nspname) || '.' || quote_ident(ic.relname)
       || ' table=' || quote_ident(tns.nspname) || '.' || quote_ident(tc.relname)
       || ' def=' || pg_temp.census_norm(pg_get_indexdef(i.indexrelid, 0, true))
       || ' valid=' || i.indisvalid
       || ' ready=' || i.indisready
       || ' live=' || i.indislive
       || ' unique=' || i.indisunique
       || ' primary=' || i.indisprimary
       || ' exclusion=' || i.indisexclusion
       || ' immediate=' || i.indimmediate
       || ' nulls_not_distinct=' || i.indnullsnotdistinct
       || ' clustered=' || i.indisclustered
       || ' replica_identity=' || i.indisreplident
  FROM pg_index i
  JOIN pg_class ic ON ic.oid = i.indexrelid
  JOIN pg_namespace ins ON ins.oid = ic.relnamespace
  JOIN pg_class tc ON tc.oid = i.indrelid
  JOIN pg_namespace tns ON tns.oid = tc.relnamespace
 WHERE tns.nspname IN ('public', 'migration')
   AND tc.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = ic.oid
          AND d.deptype = 'e'
   )
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = tc.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- RLS enable/force state and complete policy semantics.
SELECT 'RLS ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' enabled=' || c.relrowsecurity
       || ' forced=' || c.relforcerowsecurity
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
 WHERE n.nspname IN ('public', 'migration')
   AND c.relkind IN ('r', 'p')
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

SELECT 'POLICY ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' name=' || quote_ident(pol.polname)
       || ' command=' || pol.polcmd::text
       || ' permissive=' || pol.polpermissive
       || ' roles={' || coalesce((
            SELECT string_agg(CASE role_oid WHEN 0 THEN 'PUBLIC' ELSE quote_ident(r.rolname) END, ',' ORDER BY CASE role_oid WHEN 0 THEN 'PUBLIC' ELSE r.rolname END)
              FROM unnest(pol.polroles) role_oid
              LEFT JOIN pg_roles r ON r.oid = role_oid
          ), '') || '}'
       || ' using=' || coalesce(pg_temp.census_norm(pg_get_expr(pol.polqual, pol.polrelid, true)), '-')
       || ' check=' || coalesce(pg_temp.census_norm(pg_get_expr(pol.polwithcheck, pol.polrelid, true)), '-')
  FROM pg_policy pol
  JOIN pg_class c ON c.oid = pol.polrelid
  JOIN pg_namespace n ON n.oid = c.relnamespace
 WHERE n.nspname IN ('public', 'migration')
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_policy'::regclass
          AND d.objid = pol.oid
          AND d.deptype = 'e'
   )
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- Sequence parameters and ownership.  Identity and serial ownership use
-- different dependency types; both are semantic and therefore retained.
SELECT 'SEQUENCE ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' type=' || format_type(s.seqtypid, -1)
       || ' start=' || s.seqstart
       || ' min=' || s.seqmin
       || ' max=' || s.seqmax
       || ' increment=' || s.seqincrement
       || ' cache=' || s.seqcache
       || ' cycle=' || s.seqcycle
       || ' owned_by=' || coalesce(quote_ident(onsp.nspname) || '.' || quote_ident(oc.relname) || '.' || quote_ident(oa.attname), '-')
       || ' ownership_kind=' || coalesce(dep.deptype::text, '-')
  FROM pg_sequence s
  JOIN pg_class c ON c.oid = s.seqrelid
  JOIN pg_namespace n ON n.oid = c.relnamespace
  LEFT JOIN pg_depend dep
    ON dep.classid = 'pg_class'::regclass
   AND dep.objid = c.oid
   AND dep.objsubid = 0
   AND dep.refclassid = 'pg_class'::regclass
   AND dep.deptype IN ('a', 'i')
  LEFT JOIN pg_class oc ON oc.oid = dep.refobjid
  LEFT JOIN pg_namespace onsp ON onsp.oid = oc.relnamespace
  LEFT JOIN pg_attribute oa ON oa.attrelid = dep.refobjid AND oa.attnum = dep.refobjsubid
 WHERE n.nspname IN ('public', 'migration')
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- View and materialized-view definitions are emitted independently of their
-- rewrite rules so a same-count definition change remains visible.
SELECT CASE c.relkind WHEN 'v' THEN 'VIEW ' ELSE 'MATVIEW ' END
       || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' populated=' || CASE c.relkind WHEN 'm' THEN c.relispopulated::text ELSE '-' END
       || ' def=' || pg_temp.census_norm(pg_get_viewdef(c.oid, false))
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
 WHERE n.nspname IN ('public', 'migration')
   AND c.relkind IN ('v', 'm')
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- User triggers only.  Internal RI triggers have generated names/OIDs and are
-- a derived implementation of the already-captured foreign-key constraints.
SELECT 'TRIGGER ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' name=' || quote_ident(t.tgname)
       || ' enabled=' || t.tgenabled::text
       || ' function=' || quote_ident(pn.nspname) || '.' || quote_ident(p.proname)
          || '(' || pg_get_function_identity_arguments(p.oid) || ')'
       || ' def=' || pg_temp.census_norm(pg_get_triggerdef(t.oid, false))
  FROM pg_trigger t
  JOIN pg_class c ON c.oid = t.tgrelid
  JOIN pg_namespace n ON n.oid = c.relnamespace
  JOIN pg_proc p ON p.oid = t.tgfoid
  JOIN pg_namespace pn ON pn.oid = p.pronamespace
 WHERE n.nspname IN ('public', 'migration')
   AND NOT t.tgisinternal
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_trigger'::regclass
          AND d.objid = t.oid
          AND d.deptype = 'e'
   )
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- Rewrite rules include view _RETURN rules and explicit user rules.  Their
-- event and enabled state are catalog semantics outside the rendered body.
SELECT 'RULE ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' name=' || quote_ident(r.rulename)
       || ' event=' || r.ev_type::text
       || ' instead=' || r.is_instead
       || ' enabled=' || r.ev_enabled::text
       || ' def=' || pg_temp.census_norm(pg_get_ruledef(r.oid, false))
  FROM pg_rewrite r
  JOIN pg_class c ON c.oid = r.ev_class
  JOIN pg_namespace n ON n.oid = c.relnamespace
 WHERE n.nspname IN ('public', 'migration')
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_rewrite'::regclass
          AND d.objid = r.oid
          AND d.deptype = 'e'
   )
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- User-defined enum, domain, range/multirange, and standalone composite types.
-- Table row types are omitted because their complete relation/column semantics
-- are already represented above.
SELECT 'TYPE ' || quote_ident(n.nspname) || '.' || quote_ident(t.typname)
       || ' kind=' || CASE t.typtype
            WHEN 'e' THEN 'enum'
            WHEN 'd' THEN 'domain'
            WHEN 'r' THEN 'range'
            WHEN 'm' THEN 'multirange'
            WHEN 'c' THEN 'composite'
          END
       || ' details=' || CASE t.typtype
            WHEN 'e' THEN 'labels={' || coalesce((
                SELECT string_agg(quote_literal(e.enumlabel), ',' ORDER BY e.enumsortorder)
                  FROM pg_enum e WHERE e.enumtypid = t.oid
              ), '') || '}'
            WHEN 'd' THEN 'base=' || format_type(t.typbasetype, t.typtypmod)
                 || ',notnull=' || t.typnotnull
                 || ',default=' || coalesce(pg_temp.census_norm(t.typdefault), '-')
                 || ',collation=' || coalesce(quote_ident(cn.nspname) || '.' || quote_ident(coll.collname), '-')
            WHEN 'r' THEN 'subtype=' || format_type(rng.rngsubtype, -1)
                 || ',collation=' || coalesce(quote_ident(rcn.nspname) || '.' || quote_ident(rcoll.collname), '-')
                 || ',opclass=' || quote_ident(opn.nspname) || '.' || quote_ident(opc.opcname)
                 || ',canonical=' || CASE rng.rngcanonical WHEN 0 THEN '-' ELSE rng.rngcanonical::regprocedure::text END
                 || ',subdiff=' || CASE rng.rngsubdiff WHEN 0 THEN '-' ELSE rng.rngsubdiff::regprocedure::text END
            WHEN 'm' THEN 'range=' || format_type(rngm.rngtypid, -1)
            WHEN 'c' THEN 'attrs={' || coalesce((
                SELECT string_agg(quote_ident(a.attname) || ':' || format_type(a.atttypid, a.atttypmod)
                                  || ':coll=' || coalesce(quote_ident(acn.nspname) || '.' || quote_ident(ac.collname), '-'),
                                  ',' ORDER BY a.attnum)
                  FROM pg_attribute a
                  LEFT JOIN pg_collation ac ON ac.oid = a.attcollation
                  LEFT JOIN pg_namespace acn ON acn.oid = ac.collnamespace
                 WHERE a.attrelid = t.typrelid AND a.attnum > 0 AND NOT a.attisdropped
              ), '') || '}'
          END
  FROM pg_type t
  JOIN pg_namespace n ON n.oid = t.typnamespace
  LEFT JOIN pg_class cr ON cr.oid = t.typrelid
  LEFT JOIN pg_collation coll ON coll.oid = t.typcollation
  LEFT JOIN pg_namespace cn ON cn.oid = coll.collnamespace
  LEFT JOIN pg_range rng ON rng.rngtypid = t.oid
  LEFT JOIN pg_range rngm ON rngm.rngmultitypid = t.oid
  LEFT JOIN pg_collation rcoll ON rcoll.oid = rng.rngcollation
  LEFT JOIN pg_namespace rcn ON rcn.oid = rcoll.collnamespace
  LEFT JOIN pg_opclass opc ON opc.oid = rng.rngsubopc
  LEFT JOIN pg_namespace opn ON opn.oid = opc.opcnamespace
 WHERE n.nspname IN ('public', 'migration')
   AND t.typtype IN ('e', 'd', 'r', 'm', 'c')
   AND (t.typtype <> 'c' OR cr.relkind = 'c')
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_type'::regclass
          AND d.objid = t.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- Framework functions: complete identity and behavioral attributes.  The
-- source digest covers source/binary bodies; the canonical definition digest
-- also covers parsed SQL-standard bodies without dumping implementation text.
SELECT 'FUNC ' || quote_ident(n.nspname) || '.' || quote_ident(p.proname)
       || '(' || pg_get_function_identity_arguments(p.oid) || ')'
       || ' arguments={' || pg_temp.census_norm(pg_get_function_arguments(p.oid)) || '}'
       || ' result=' || pg_get_function_result(p.oid)
       || ' language=' || l.lanname
       || ' kind=' || p.prokind::text
       || ' volatility=' || p.provolatile::text
       || ' strict=' || p.proisstrict
       || ' security_definer=' || p.prosecdef
       || ' leakproof=' || p.proleakproof
       || ' parallel=' || p.proparallel::text
       || ' cost=' || p.procost
       || ' rows=' || p.prorows
       || ' config={' || coalesce((
            SELECT string_agg(cfg, ',' ORDER BY cfg) FROM unnest(p.proconfig) cfg
          ), '') || '}'
       || ' source_md5=' || md5(coalesce(p.prosrc, '') || E'\n--probin--\n' || coalesce(p.probin, ''))
       || ' definition_md5=' || md5(pg_get_functiondef(p.oid))
  FROM pg_proc p
  JOIN pg_namespace n ON n.oid = p.pronamespace
  JOIN pg_language l ON l.oid = p.prolang
 WHERE n.nspname IN ('public', 'migration')
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_proc'::regclass
          AND d.objid = p.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

-- Role contracts.  LOGIN/password/valid-until are deliberately excluded:
-- deployments own credentials, while privilege-bearing attributes are fixed
-- framework semantics.  Membership includes either direction involving a
-- framework role and all three PostgreSQL 16 membership options.
SELECT 'ROLE ' || quote_ident(r.rolname)
       || ' superuser=' || r.rolsuper
       || ' inherit=' || r.rolinherit
       || ' create_role=' || r.rolcreaterole
       || ' create_db=' || r.rolcreatedb
       || ' replication=' || r.rolreplication
       || ' bypass_rls=' || r.rolbypassrls
       || ' connlimit=' || r.rolconnlimit
       || ' config={' || coalesce((SELECT string_agg(cfg, ',' ORDER BY cfg) FROM unnest(r.rolconfig) cfg), '') || '}'
  FROM pg_roles r
 WHERE r.rolname IN ('app_rt', 'app_platform')
 ORDER BY 1;

SELECT 'ROLE.MEMBER role=' || quote_ident(parent.rolname)
       || ' member=' || quote_ident(member.rolname)
       || ' grantor=' || CASE WHEN grantor.rolname IN ('app_rt', 'app_platform') THEN quote_ident(grantor.rolname) ELSE 'ENV_OWNER' END
       || ' admin=' || m.admin_option
       || ' inherit=' || m.inherit_option
       || ' set=' || m.set_option
  FROM pg_auth_members m
  JOIN pg_roles parent ON parent.oid = m.roleid
  JOIN pg_roles member ON member.oid = m.member
  JOIN pg_roles grantor ON grantor.oid = m.grantor
 WHERE parent.rolname IN ('app_rt', 'app_platform')
    OR member.rolname IN ('app_rt', 'app_platform')
 ORDER BY 1;

-- Object ACLs are expanded directly from catalog aclitem[] values.  This
-- captures implicit defaults, PUBLIC, grantor, grantee and grant option without
-- information_schema's current-user filtering or has_*_privilege inheritance.
SELECT 'ACL.' || CASE c.relkind
            WHEN 'S' THEN 'SEQUENCE'
            WHEN 'v' THEN 'VIEW'
            WHEN 'm' THEN 'MATVIEW'
            ELSE 'TABLE'
          END
       || ' ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname)
       || ' grantor=' || pg_temp.census_acl_role(acl.grantor, c.relowner)
       || ' grantee=' || pg_temp.census_acl_role(acl.grantee, c.relowner)
       || ' privilege=' || acl.privilege_type
       || ' grant_option=' || acl.is_grantable
  FROM pg_class c
  JOIN pg_namespace n ON n.oid = c.relnamespace
  CROSS JOIN LATERAL aclexplode(coalesce(c.relacl,
      acldefault(CASE c.relkind WHEN 'S' THEN 's'::"char" ELSE 'r'::"char" END, c.relowner))) acl
 WHERE n.nspname IN ('public', 'migration')
   AND c.relkind IN ('r', 'p', 'f', 'S', 'v', 'm')
   AND c.relname NOT LIKE 'goose_%'
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

SELECT 'ACL.COLUMN ' || quote_ident(n.nspname) || '.' || quote_ident(c.relname) || '.' || quote_ident(a.attname)
       || ' grantor=' || pg_temp.census_acl_role(acl.grantor, c.relowner)
       || ' grantee=' || pg_temp.census_acl_role(acl.grantee, c.relowner)
       || ' privilege=' || acl.privilege_type
       || ' grant_option=' || acl.is_grantable
  FROM pg_attribute a
  JOIN pg_class c ON c.oid = a.attrelid
  JOIN pg_namespace n ON n.oid = c.relnamespace
  CROSS JOIN LATERAL aclexplode(a.attacl) acl
 WHERE n.nspname IN ('public', 'migration')
   AND c.relkind IN ('r', 'p', 'f', 'v', 'm')
   AND c.relname NOT LIKE 'goose_%'
   AND a.attnum > 0
   AND NOT a.attisdropped
   AND a.attacl IS NOT NULL
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_class'::regclass
          AND d.objid = c.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

SELECT 'ACL.FUNCTION ' || quote_ident(n.nspname) || '.' || quote_ident(p.proname)
       || '(' || pg_get_function_identity_arguments(p.oid) || ')'
       || ' grantor=' || pg_temp.census_acl_role(acl.grantor, p.proowner)
       || ' grantee=' || pg_temp.census_acl_role(acl.grantee, p.proowner)
       || ' privilege=' || acl.privilege_type
       || ' grant_option=' || acl.is_grantable
  FROM pg_proc p
  JOIN pg_namespace n ON n.oid = p.pronamespace
  CROSS JOIN LATERAL aclexplode(coalesce(p.proacl, acldefault('f', p.proowner))) acl
 WHERE n.nspname IN ('public', 'migration')
   AND NOT EXISTS (
       SELECT 1 FROM pg_depend d
        WHERE d.classid = 'pg_proc'::regclass
          AND d.objid = p.oid
          AND d.deptype = 'e'
   )
 ORDER BY 1;

SELECT 'ACL.SCHEMA ' || quote_ident(n.nspname)
       || ' grantor=' || pg_temp.census_acl_role(acl.grantor, n.nspowner)
       || ' grantee=' || pg_temp.census_acl_role(acl.grantee, n.nspowner)
       || ' privilege=' || acl.privilege_type
       || ' grant_option=' || acl.is_grantable
  FROM pg_namespace n
 CROSS JOIN LATERAL aclexplode(coalesce(n.nspacl, acldefault('n', n.nspowner))) acl
 WHERE n.nspname IN ('public', 'migration')
 ORDER BY 1;

-- Default ACLs are scoped by owner, schema and target object class.  Owner and
-- non-framework grantors are normalized because the migration login is an
-- environment concern; framework grantees and PUBLIC retain exact identity.
SELECT 'ACL.DEFAULT owner=ENV_OWNER'
       || ' schema=' || coalesce(quote_ident(n.nspname), '-')
       || ' object=' || CASE d.defaclobjtype
            WHEN 'r' THEN 'table'
            WHEN 'S' THEN 'sequence'
            WHEN 'f' THEN 'function'
            WHEN 'T' THEN 'type'
            WHEN 'n' THEN 'schema'
          END
       || ' grantor=' || CASE WHEN grantor.rolname IN ('app_rt', 'app_platform') THEN quote_ident(grantor.rolname) ELSE 'ENV_OWNER' END
       || ' grantee=' || CASE acl.grantee WHEN 0 THEN 'PUBLIC' ELSE quote_ident(grantee.rolname) END
       || ' privilege=' || acl.privilege_type
       || ' grant_option=' || acl.is_grantable
  FROM pg_default_acl d
  LEFT JOIN pg_namespace n ON n.oid = d.defaclnamespace
  CROSS JOIN LATERAL aclexplode(d.defaclacl) acl
  JOIN pg_roles grantor ON grantor.oid = acl.grantor
  LEFT JOIN pg_roles grantee ON grantee.oid = acl.grantee
 WHERE n.nspname IN ('public', 'migration')
    OR d.defaclnamespace = 0
 ORDER BY 1;
