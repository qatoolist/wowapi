---
id: W07-ACCEPTANCE
type: wave-acceptance
wave: W07
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W07 — Wave-level acceptance

## AC-W07-01 — Performance programme relative evidence published; absolute SLOs correctly conditional

PERF-02..05's relative/container comparison evidence is published against `perf/reference-v1.json` for
each finding's own task table; every absolute-SLO acceptance criterion is explicitly recorded as
conditional on DEC-Q9, not silently asserted unconditionally. BENCH_PKGS covers the 7 MATRIX CS-16-named
hot-path packages with passing bench-budget entries. Traces to W07-E01-S001..S004.

## AC-W07-02 — Security verification profile and coverage-truthfulness closure complete

SEC-05's control map leaves zero open Critical/High findings (or each has an approved waiver), backed by
an external assessment. REL-04 T5-T8 are complete: fail-not-skip E2E, machine-checked skip manifest,
race-integration schedule, real time-bounded coverage-guided fuzzing (owning PERF-06 T3/T4's scope).
Traces to W07-E02-S001, W07-E02-S002.

## AC-W07-03 — Product-alignment verification complete, framework-side only

Every PROD-01..05 row has a documented framework-side coordination artifact confirming existence and a
documented product upgrade path, with zero wowsociety-repository code change performed by this wave.
Traces to W07-E03-S001.

## AC-W07-04 — Programme-wide final verification gate re-run and closure decision package complete

The REVIEW §30-style final approval gate has been re-run across the whole 8-wave programme against
current HEAD; the traceability matrix shows every `requirement-inventory.md` row with a disposition and
no silent drop; the disposition audit confirms every item genuinely reached its recorded disposition. A
closure report and a separate, explicit production-readiness claim-upgrade decision package exist for
the human authority — this wave does not itself declare the framework production-ready. Traces to
W07-E04-S001, W07-E04-S002.

## AC-W07-05 — Independent review passed

Every W07 story with a P0/critical priority, and SEC-05 specifically (given its closure-gate role), has
passed independent review per mandate §14. W07-E01's stories are specifically checked for their
absolute-SLO acceptance criteria being genuinely conditional on DEC-Q9, not silently unconditional.
W07-E04-S001 is specifically checked for genuinely re-running the gate against current HEAD, not
restating REVIEW's own prior 2026-07-11 conclusions without re-verification.

## Acceptance authority

Performance/SRE lead (W07-E01); product-security lead (W07-E02); a cross-functional authority spanning
every prior wave's own accountable role (W07-E03/E04) — see `wave.md` "Acceptance authority" for the
full rationale.
