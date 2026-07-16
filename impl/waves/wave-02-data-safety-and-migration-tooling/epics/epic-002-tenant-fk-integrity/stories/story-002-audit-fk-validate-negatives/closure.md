---
id: CLOSURE-W02-E02-S002
type: closure-record
parent_story: W02-E02-S002
status: accepted
created_at: 2026-07-13
updated_at: 2026-07-13
---

# Closure — W02-E02-S002

## Acceptance-criteria completion

- AC-W02-E02-S002-01: pass — `TestIntegrationTenantFKMismatchAuditZero` reports
  0 cross-tenant mismatches across all 9 edges using a platform-role connection.
- AC-W02-E02-S002-02: pass — composite FKs added `NOT VALID` for all 9 edges in
  migration 00035.
- AC-W02-E02-S002-03: pass — `VALIDATE CONSTRAINT` executed for all 9 edges and
  redundant single-column FKs dropped in migration 00036.
- AC-W02-E02-S002-04: pass — `TestIntegrationTenantFKCrossTenantInsertBlocked`
  proves seeded cross-tenant inserts fail under both `app_rt` and `app_platform`
  roles; admin probe fails with `foreign_key_violation` (SQLSTATE 23503).
- AC-W02-E02-S002-05: pass — `TestIntegrationTenantFKEdgeCensus` confirms the
  live schema contains exactly the 9 composite tenant FKs with no silent gaps
  (corrected from a documented "8 edges" figure — see review-gate-2026-07-16.md's
  per-story record, task-006-independent-review.md).

— count correction dated 2026-07-16, conductor adjudication (Fable 5), per
review-gate-2026-07-16.md records

## Task completion

- W02-E02-S002-T001: complete.
- W02-E02-S002-T002: complete.
- W02-E02-S002-T003: complete.
- W02-E02-S002-T004: complete.
- W02-E02-S002-T005: complete.
- W02-E02-S002-T006: complete (review gate W02ReviewGate).

## Artifact completeness

- ART-W02-E02-S002-001: mismatch audit query / test.
- ART-W02-E02-S002-002: migration 00035.
- ART-W02-E02-S002-003: migration 00036.
- ART-W02-E02-S002-004: `testkit/tenant_fk_cross_tenant_test.go`.
- ART-W02-E02-S002-005: optional FK cleanup deferred; no blocking work remains.

## Evidence completeness

- EV-W02-E02-S002-001: mismatch audit zero-mismatch report.
- EV-W02-E02-S002-002: edge census.
- EV-W02-E02-S002-003: cross-tenant insert blocked.

## Unresolved findings

None.

## Accepted risks

RISK-W02-002 resolved: zero mismatches found; remediation path not triggered.

## Deferred work

T8 (removal of redundant single-column FKs) was completed in migration 00036 as
part of validation/cleanup, so no separate deferred work remains.

## Reviewer conclusion

Independent review passed (W02ReviewGate, 2026-07-13). Reviewer confirmed the
mismatch audit, composite FK validation, and cross-tenant negative tests.

## Acceptance authority

data/reliability lead.

## Closure date

2026-07-13.

## Final status

accepted.

— CORRECTION (2026-07-16, findings-remediation adjudication): acceptance status rolled back to `implemented` per discovery that the three tasks' named proof artifacts (T001: seeded-mismatch detection test; T002: lock-duration report; T003: concurrent-writer-load test) were never built, while the schema properties themselves (mismatch audit, composite FK validation, cross-tenant insert rejection) were independently re-verified live on 2026-07-16. See story.md status note and programme-deviations.md DEV-PROG-005.
