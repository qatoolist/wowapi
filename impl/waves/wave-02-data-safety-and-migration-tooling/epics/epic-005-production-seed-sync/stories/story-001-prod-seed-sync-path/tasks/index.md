---
id: W02-E05-S001-TASKS-INDEX
type: tasks-index
parent_story: W02-E05-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W02-E05-S001 — Tasks index

Per mandate §16.4. Task files are single-file per the repository's documented adaptation (see
`governance/naming-conventions.md` "Adaptation 1") — each task file below contains its task
definition, implementation record, verification record, and deviations record as internal sections.

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W02-E05-S001-T001](task-001-catalog-manifest-design-investigation.md) | Catalog manifest design investigation | W02SeedRerun | done | none | Documented design decision (manifest format, versioning, CLI shape, idempotency mechanism, RLS/role posture, dry-run format, audit integration) with rationale | AC-W02-E05-S001-01 | completed | completed |
| [W02-E05-S001-T002](task-002-seed-sync-command-and-manifest-implementation.md) | Seed-sync command and manifest implementation | W02SeedRerun | done | T001 | Seed-sync command/path + catalog manifest schema/loader; idempotency + RLS/role posture proven | AC-W02-E05-S001-02 | completed | completed |
| [W02-E05-S001-T003](task-003-dry-run-and-audit-record.md) | Dry-run mode and audit-record production | W02SeedRerun | done | T001 | Dry-run reporting + per-run audit record | AC-W02-E05-S001-03 | completed | completed |
| [W02-E05-S001-T004](task-004-readiness-check-registration.md) | Readiness-check registration | W02SeedRerun | done | T001, T002 | Named readiness check failing until seed-sync has run against an empty catalog database | AC-W02-E05-S001-04 | completed | completed |
| [W02-E05-S001-T005](task-005-readiness-hash-reporting-and-docs.md) | Readiness-payload hash reporting and documentation | W02SeedRerun | done | T001, T002, T004 | Readiness payload reports seed/catalog hash; manifest/CLI/readiness/audit documentation | AC-W02-E05-S001-04 | completed | completed |
| [W02-E05-S001-T006](task-006-independent-review.md) | Independent review | W02ReviewGate | done | T001, T002, T003, T004, T005 | Independent-review record per mandate §14 | AC-W02-E05-S001-01, AC-W02-E05-S001-02, AC-W02-E05-S001-03, AC-W02-E05-S001-04, AC-W02-E05-S001-05 | completed | completed |

## Grouping rationale

This is the single most important piece of narrative reasoning in this epic: **the design-
investigation task (T001) is sequenced strictly before every implementation task (T002–T005), and no
implementation task may begin until T001's decision record exists.** This ordering is not a stylistic
preference — it is required by the source material's own framing and by mandate §18. MATRIX CS-21
states explicitly that FBL-02's "design detail [is] to be ratified in Phase 5, but the acceptance bar
is fixed now" — meaning the *what must be true* (idempotent, RLS-respecting, versioned, dry-run,
audited; readiness gated on seed-sync; hash reported) is settled, but the *how it is built* (manifest
format, versioning scheme, CLI shape, idempotency mechanism, RLS/role posture, dry-run format, audit
integration) is explicitly not. Mandate §18 states directly: "Where implementation details cannot yet
be known, state what must be determined during the story rather than inventing specifics." Building
T002–T005 before T001 resolves would force this plan to either (a) invent the very specifics mandate
§18 forbids inventing, or (b) leave T002–T005 so vague they carry no real implementation content —
both unacceptable. Sequencing T001 first converts an unresolvable planning problem (implementation
tasks for a design that does not yet exist) into a resolvable one (implementation tasks gated on a
design-investigation task whose own job is to produce that design).

**Why T001's output is a documented design decision, not a new ADR.** `epic.md` "Required decisions"
states explicitly: "None in this programme's D-0N ADR sense — no D-01..D-09 decision targets FBL-02."
`../../../../wave.md` "Assumptions" confirms this negatively across the whole wave: no D-0N ADR
targets DATA-09, DATA-01, DATA-05, DATA-06, or FBL-02. This wave's governing brief accordingly states
that no W02 epic enacts a new programme-level ADR — T001's resolution of the catalog-manifest-format
question is real design content, but it is a **story-scoped implementation decision**, recorded as
this task's own "Expected output," not a `decisions/`-directory ADR entry. This story's directory tree
accordingly has no `decisions/` subdirectory (see `epic.md`'s explicit escalation safeguard: if T001's
findings turn out to be of genuinely D-0N caliber — a new framework-wide convention intended to outlive
this story, or a new external dependency — that must be escalated through the programme's decision
register rather than silently absorbed here; this is a contingency to detect and record, not this
task's expected outcome).

**Why T002–T005 are four separate tasks, not one "implementation" task.** Per mandate §12's
decomposition triggers ("contain multiple independent acceptance outcomes," "need separate
evidence," "have materially different risks"): T002 (seed-sync core + idempotency + RLS posture) and
T003 (dry-run + audit) produce evidence for different acceptance criteria (AC-02 vs AC-03) and carry
different risk profiles (T002's risk is the RLS/role-posture safety question, RISK-W02-E05-001; T003's
risk is comparatively low — dry-run and audit are additive reporting concerns). T004 (readiness-check
registration) and T005 (hash reporting + documentation) both touch the same readiness seam and both
feed AC-04, but are kept separate because T005 additionally depends on T002 (the hash must reflect the
actual synced manifest) in a way T004 does not, and because T005 bundles the documentation deliverable
that closes out the story's "Documentation requirements" — collapsing T004/T005 into one task would
either delay hash-reporting work behind full readiness-check completion unnecessarily or force
documentation to be written before the mechanisms it documents are stable. A four-way implementation
split was also considered against a two-way split (core-sync-path vs readiness-wiring); the four-way
split was chosen because AC-W02-E05-S001-02 and -03's evidence (idempotency/RLS proof vs dry-run/audit
proof) are genuinely separable and reviewable independently, consistent with the same
separate-evidence reasoning epic-001-online-migration-protocol's story-001 applied to its own T001/T002
split.

**Why T006 (independent review) exists as its own task.** This story is P0-prod — the highest priority
grade this programme uses, reserved for a confirmed production-blocking gap (CS-21: "prod boots with
deny-everything catalogs... prod-blocking"). Per mandate §14's independent-review requirement for
critical stories, and consistent with epic-001-online-migration-protocol's story-001-T003 precedent, a
dedicated review task is added rather than folding review into T005's own completion claim — this
story's dominant residual risk (a design-investigation decision quietly reverse-justified after
implementation, rather than genuinely predating it) is exactly the failure mode an independent,
dependency-gated review task exists to catch. No separate evidence-aggregation task is added beyond
T006 — six tasks (one investigation, four implementation, one review) is not large enough to warrant a
seventh aggregation-only task, per mandate §12's fragmentation-avoidance guidance applied the same way
epic-001's story-001 applied it to its own T003.
