---
id: W02-E01-ACCEPTANCE
type: epic-acceptance
epic: W02-E01
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E01 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W02-01 there maps onto this epic).

## AC-W02-E01-01 — Manifest schema and lock budget enforced

The migration manifest schema validates online/maintenance classification, rows/bytes estimate,
lock/statement timeout, N/N-1 compatibility flag, backfill owner, validation query, and rollback/
forward-fix plan on every migration; a missing required field fails CI via a negative fixture test.
The 2-second online-DDL lock-timeout enforcement aborts cleanly against a concurrently-locked table
with a bounded retry ceiling — no unbounded retry. Traces to W02-E01-S001.

## AC-W02-E01-02 — Expand, backfill, and validate tooling proven

Expand-phase tooling issues `CREATE INDEX CONCURRENTLY` and `NOT VALID` constraints without
blocking traffic, proven by an old-reader-compatibility test. The backfill-job harness's named
interrupted/resumed test passes with no reprocessing or skipping of any row. Validation-phase
tooling's zero-mismatch report is a machine-checked, artifact-schema-conformant record. Traces to
W02-E01-S002.

## AC-W02-E01-03 — Canary, switch, contract, and CI drills proven

Canary/deploy-N tooling's named test proves N-1 alongside N-expanded-schema both before and after
backfill. Switch-phase tooling's named test proves application rollback after switch with no
destructive `Down`. Contract-phase tooling's named test proves forward recovery from every failed
phase and gates the contract step on an evidenced no-N-1-remains precondition. All six directive-
named drills run in the CI/scheduled pipeline with a passing run artifact retained as evidence.
Traces to W02-E01-S003.

## AC-W02-E01-04 — Independent review passed

All three stories (S001, S002, S003) have passed independent review per mandate §14. S002's review
specifically confirms the interim-checkpoint-lease deviation is recorded honestly (bounded scope,
explicit forward reference to W04-E01-S001), not silently presented as a complete DATA-02 T1
substitute. S003's review specifically confirms the soak-threshold judgment gap is recorded as an
accepted residual risk, not silently resolved with an unjustified numeric value.

## Acceptance authority

Data/reliability lead, per `../../wave.md`'s wave-level acceptance authority (PLAN §5.3's
accountable role for PF-DATA).
