---
id: W02-ACCEPTANCE
type: wave-acceptance
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02 — Wave-level acceptance

## AC-W02-01 — Online migration protocol operational end-to-end

The DATA-09 manifest schema validates every migration and fails CI on a missing field; the
2-second online-DDL lock-timeout enforces abort-and-retry with a bounded retry ceiling; expand-phase
tooling issues non-transactional `CREATE INDEX CONCURRENTLY` and `NOT VALID` constraints without
blocking traffic; the backfill-job harness passes its named interrupted/resumed test with no
reprocessing or skipping; validation-phase tooling produces machine-checked zero-mismatch artifacts;
canary/deploy-N tooling proves N-1 code runs correctly against N-expanded schema both before and
after backfill; switch-phase tooling proves application rollback after switch with no destructive
`Down`; contract-phase tooling proves forward recovery from every failed phase and gates on
evidenced absence of N-1 traffic; all six directive-named drills run in the CI/scheduled pipeline.
Traces to W02-E01-S001, W02-E01-S002, W02-E01-S003.

## AC-W02-02 — Composite tenant foreign keys closed

All 8 confirmed tenant-scoped child-table edges have a `VALIDATE CONSTRAINT`-clean composite FK on
`(tenant_id, id)`; the tenant-FK catalog scanner is wired as a permanent CI gate that fails a new
migration adding a single-column tenant FK; the mismatch audit reports zero cross-tenant mismatches
against staging/prod-shaped data (or a documented, resolved remediation decision per
`risks.md` RISK-W02-002 if a mismatch was found); a seeded cross-tenant insert fails under both
`app_rt` and `app_platform`. Traces to W02-E02-S001, W02-E02-S002.

## AC-W02-03 — Version allocation race-free; orphan blobs garbage-collected

`kernel/artifact.Generate` and `kernel/document.InitiateUpload` allocate versions via a locked
counter or dedicated sequence row, proven race-free under a ≥20-concurrent-caller test; upload
sessions are durable and survive a simulated crash in `status='pending'`; confirmation is atomic
under a racing-confirm test (exactly one succeeds); scheduled GC removes every past-expiry
unconfirmed session and never a referenced object. Traces to W02-E03-S001.

## AC-W02-04 — Aggregate write contract framework-enforced

A module cannot write its business row without the framework also writing the resource mirror in
the same transaction, proven by fault injection at each of 4 stages independently with full
rollback at every stage; real `created_by` actor attribution replaces the `uuid.Nil` placeholder in
`registrar_pg.go`, with a user-initiated write carrying no actor failing fast and system-actor paths
unaffected; the reference handler is migrated onto the new helper; `kernel/resource` documentation
matches the mandatory-mirror-contract implementation. Traces to W02-E04-S001.

## AC-W02-05 — Production seed-sync path closes the empty-catalog gap

A `wowapi seed sync`-shaped path (or equivalent, per the design-investigation's resolved catalog
manifest format) is idempotent, RLS-respecting, versioned, supports dry-run, and produces an audit
record; a prod-profile boot against an empty catalog database no longer silently reaches a
deny-everything ready state — readiness returns a named failure until seed-sync has run, and the
readiness payload reports the seed/catalog hash once it has. Traces to W02-E05-S001.

## AC-W02-06 — Independent review passed

Every W02 story has passed independent review per mandate §14. E01 and E02 stories (all P0) and E05
(P0-prod) are specifically checked for: no silent scope reduction against PLAN's own T-row
acceptance criteria; the E01-S002 minimal-checkpoint-lease deviation correctly recorded (not hidden)
per RISK-W02-001; the E02-S002 mismatch-audit outcome correctly recorded per RISK-W02-002 whichever
way it resolves; and FBL-02's design-investigation decision correctly documented before its
implementation tasks were started.

## Acceptance authority

Data/reliability lead, per `wave.md`'s "Acceptance authority" (PLAN §5.3's own accountable role for
PF-DATA, applied uniformly across this wave including FBL-02).
