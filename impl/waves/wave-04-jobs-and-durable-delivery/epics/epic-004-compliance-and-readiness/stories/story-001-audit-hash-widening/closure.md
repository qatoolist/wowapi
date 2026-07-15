---
id: CLOSURE-W04-E04-S001
type: closure-record
parent_story: W04-E04-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Closure — W04-E04-S001

## Acceptance-criteria completion

- **AC-W04-E04-S001-01**: PASS. `chainHash` now covers every persisted field, including canonicalized
  `metadata` and `tx_id`. `TestIntegrationAuditChainDetectsPerFieldTampering` mutates each of 17
  declared fields independently and asserts verification fails.
- **AC-W04-E04-S001-02**: PASS. Migration `00037_audit_hash_version.sql` adds `hash_version smallint
  NOT NULL DEFAULT 1`. `Verify` branches by row version: v1 rows use the historical 15-field scheme;
  v2 rows include `metadata` and `tx_id`. `TestIntegrationAuditHashVersionBranching` and
  `TestIntegrationAuditUnknownHashVersionFailsClosed` prove both branches and the fail-closed
  behavior.
- **AC-W04-E04-S001-03**: PASS. The migration carries a `+wowapi:manifest` block with classification
  `online`, lock-timeout budget 2000 ms, validation query, and rollback plan, satisfying W02-E01's
  manifest schema. `migrations/manifest_test.go` validates it.

## Task completion

- W04-E04-S001-T001 — Audit hash-chain widening, `hash_version` migration, and version-branched
  verification: COMPLETE.
- W04-E04-S001-T002 — Independent review: PENDING (must be completed per mandate §14).

## Artifact completeness

- Widened `chainHash` implementation and metadata canonicalization: `kernel/audit/audit.go`.
- `hash_version` migration: `migrations/00037_audit_hash_version.sql`.
- Version-branched `Verify`: `kernel/audit/audit.go`.
- Documentation of widened field list and version-branch semantics: inline comments in
  `kernel/audit/audit.go`.

## Evidence completeness

- Per-field tamper test output: `go test ./kernel/audit/... -run
  TestIntegrationAuditChainDetectsPerFieldTampering -count=1 -v`.
- Version-branch verification test output: `go test ./kernel/audit/... -run
  'TestIntegrationAuditHashVersionBranching|TestIntegrationAuditUnknownHashVersionFailsClosed' -count=1 -v`.
- Migration manifest validation: `go test ./migrations/... -run TestKernelMigrationsHaveManifests -count=1 -v`.

## Unresolved findings

None.

## Accepted risks

- RISK-W04-002 (highest-risk-in-wave status, breaking format change hitting wowsociety's live audit
  rows): accepted as residual risk pending the product-side PROD-05 staging drill.
- RISK-W04-E04-S001-001 (incorrect metadata canonicalization could reintroduce non-reproducibility):
  mitigated by using deterministic `json.Marshal` on the unmarshaled map; accepted as resolved for
  framework-side closure.

## Deferred work

- PROD-05 (wowsociety-side staging audit re-verification drill) is recorded as a non-blocking
  coordination item, not implemented by this story.

## Reviewer conclusion

PENDING. The reviewer must confirm:
1. The per-field tamper test genuinely covers every declared field independently.
2. D-04's version-branch design (`hash_version=1` historical, `hash_version=2` widened) was
   implemented exactly as ratified.

## Acceptance authority

Data/reliability lead, per epic-level `acceptance.md`.

## Closure date

2026-07-13 (framework-side implementation and verification complete; pending independent review).

## Final status

`closed-pending-review` — implementation and evidence complete; awaiting mandatory independent
review per mandate §14.
