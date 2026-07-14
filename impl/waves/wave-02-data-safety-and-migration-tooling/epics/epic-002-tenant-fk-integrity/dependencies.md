---
id: W02-E02-DEPS
type: epic-dependencies
epic: W02-E02
wave: W02
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E02 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W02-E01** (this wave, hard dependency) — S002's T4/T5 (composite FK `NOT VALID` add and
  `VALIDATE CONSTRAINT`) must not start before W02-E01-S001 and W02-E01-S002 reach `accepted`, per
  `impl/analysis/wave-allocation-detail.md`'s explicit cross-wave sequencing note: "DATA-01 T4/T5
  must not start before W02-E01 S001+S002 acceptance." This is recorded here at epic scope and again
  at story scope (`stories/story-002-audit-fk-validate-negatives/story.md`'s `depends_on` front
  matter lists `W02-E01-S001` and `W02-E01-S002` explicitly) — not merely as prose, per this
  programme's traceability requirement that dependencies be explicit, not silent (mandate §11.6).
- **W00** (full wave) — per `../../dependencies.md` (wave-level), W02 depends on W00's exit gate.
  This epic has no additional upstream dependency beyond that gate and the W02-E01 dependency above.

## Downstream (epics/waves that depend on this epic)

No epic within this programme currently depends on W02-E02's own output beyond the general
tenant-isolation-integrity improvement it provides framework-wide. wowsociety's own
`policy_override.rule_version_id` migration (tracked as `PROD-01`, product-level) depends on this
epic's T1 (`UNIQUE (tenant_id, id)` on `rule_versions`) landing first, then follows this wave's
DATA-09 protocol for its own rollout — this is a product-level consumer, not a framework-epic
dependency, and is excluded from this epic's own closure per mandate §2.3.

## Internal (within this epic)

**S002 depends on S001.** Concretely: S002's T3 (mismatch audit) is more useful once S001's T2
catalog scanner confirms the exact 8-edge FK inventory is complete and current (a mismatch audit
against an incomplete FK inventory could silently miss an edge); S002's T4/T5 additionally depend on
W02-E01 as stated above. PLAN's own Depends-on column for T4 lists "T1, T3" and for T5 lists "T4" —
T1 is S001's own task, so S002's T4 has both an intra-epic dependency (S001-T1) and the cross-wave
W02-E01 dependency. T7 (negative tests) depends on T5 (PLAN's own Depends-on column). T8 (optional
cleanup) depends on T5 and T7 (PLAN's own Depends-on column: "T5, T7").

S001's own internal task order: T6 (the CI gate) depends only on T2 (the scanner) per PLAN's own
Depends-on column, and PLAN's own risk note recommends doing it "first if sequencing allows" since
it is "cheapest, most durable" — S001's `plan.md` should sequence T6 promptly after T2 rather than
last, even though T1 (parent indexes) has no dependency and can proceed in parallel with T2.

## Cross-wave dependencies

- **W02-E01** (this wave) — see "Upstream" above; this is the epic's single most consequential
  dependency.

## External dependencies

None new. This epic operates entirely within the existing PostgreSQL schema and RLS machinery
already in use.

## Repository dependencies

- **wowsociety** — `PROD-01` (product-level, tracked in `requirement-inventory.md` §D): wowsociety's
  `policy_override.rule_version_id` composite-FK migration depends on this epic's T1 landing in
  wowapi first, then follows this wave's DATA-09 protocol. Not a blocking dependency for this
  epic's own closure.

## Tooling dependencies

None beyond the CI infrastructure this epic's S001-T6 gate is wired into (the same CI infrastructure
W02-E01's manifest-schema validation and lock-timeout enforcement are wired into).

## Decision dependencies

None. See `epic.md` "Required decisions."
