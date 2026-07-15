---
id: ART-W07-E01-S002-001
type: audit-report
parent_story: W07-E01-S002
producing_task: W07-E01-S002-T001
status: produced
created_at: 2026-07-14
source_revision: 733ef3e930cbb3f89f5bbc53d8f562c60e426513
---

# `rule_versions` index-definition audit

## Sequencing

This T0 audit was executed and recorded before any W07-E01-S002 query-design, migration, or production-code change. The inspected source revision was `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

## Inspected migration

`migrations/00008_rules.sql` is the only migration whose filename matches `*rules*.sql`. It defines:

- an active-only GiST exclusion constraint across `rule_key`, `scope_kind`, normalized `scope_id`, normalized `tenant_id`, and the effective-time range;
- one B-tree lookup index:

```sql
CREATE INDEX rule_versions_lookup
    ON rule_versions (rule_key, scope_kind, effective_from)
    WHERE status = 'active';
```

## Finding

The directive's claim is **confirmed**: the only explicit lookup index is partial on `status = 'active'`, so it favors active-only lookup. The live resolver predicate is broader: it reads `status IN ('active','superseded')`, filters tenant/scope identity and temporal validity, and orders by `effective_from DESC`. Therefore the existing index does not fully match either the current/historical resolver predicate or all of its equality columns. T1 query design and T2 index work must account for this observed migration reality rather than assuming current and historical predicates are already covered.
