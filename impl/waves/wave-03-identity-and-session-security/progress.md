---
id: W03-PROGRESS
type: wave-progress
wave: W03
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03 progress

Per mandate §16.2. Story-level planning and task plans are complete; all stories are `ready` and
task files have been created under each story's `tasks/` directory.

## Epic status

| Epic | Title | Status | Stories | Story status breakdown |
|---|---|---|---|---|
| W03-E01 | server-side-session-state | planned | 4 | 4 ready |
| W03-E02 | outbound-security-governance | planned | 1 | 1 ready |
| W03-E03 | webhook-authenticated-replay | planned | 1 | 1 ready |
| W03-E04 | relationship-semantics | planned | 1 | 1 ready |
| W03-E05 | workflow-privileged-completion | planned | 1 | 1 ready |

## Story status

| Story | Title | Status | Task count | Task status breakdown |
|---|---|---|---|---|
| W03-E01-S001 | grant-schema-and-membership | ready | 4 | 4 todo (T001–T004) |
| W03-E01-S002 | capacity-and-privileged-resolver | ready | 3 | 3 todo (T001–T003) |
| W03-E01-S003 | assurance-and-credential-schemes | ready | 3 | 3 todo (T001–T003) |
| W03-E01-S004 | cross-repo-cutover-plan | ready | 3 | 3 todo (T001–T003) |
| W03-E02-S001 | outbound-security-governance | ready | 6 | 6 todo (T001–T006) |
| W03-E03-S001 | webhook-authenticated-replay | ready | 5 | 5 todo (T001–T005) |
| W03-E04-S001 | relationship-semantics | ready | 4 | 4 todo (T001–T004) |
| W03-E05-S001 | workflow-privileged-completion | ready | 3 | 3 todo (T001–T003) |

## Task plan summary

Each story now has a `tasks/` directory with descriptive task files and a `tasks/index.md`.

- W03-E01-S001: 4 tasks — `identity_grant` migration, active-tenant-access membership,
  zero/unknown-tenant rejection, independent review.
- W03-E01-S002: 3 tasks — capacity selection, privileged-session resolver, independent review.
- W03-E01-S003: 3 tasks — assurance freshness, credential-scheme distinction, independent review.
- W03-E01-S004: 3 tasks — sequencing plan, staging-validation plan, rollback plan.
- W03-E02-S001: 6 tasks — fingerprint-scope confirmation, boot-time egress report, allowlist
  change-audit, JWKS-client governance gate, fitness check, independent review.
- W03-E03-S001: 5 tasks — Verifier/Envelope interface, HMAC envelope synthesis, `HandleInbound`
  rewire, provider-verifier contract document, independent review.
- W03-E04-S001: 4 tasks — party-subject evaluation, subject-kind matrix, mutation governance,
  independent review.
- W03-E05-S001: 3 tasks — ratification decision/implementation, durable override audit,
  independent review.

## Blocked items

No story is `blocked` at the planning stage. W03-E04-S001 remains logically gated on W03-E01's
acceptance (hard dependency, PLAN §5.3) and will be re-flagged `blocked` if implementation work
reaches it before E01 is accepted. W03-E05-S001-T002 is gated on W03-E01-S001's `identity_grant`
shape for the grant-ID field, but the story as a whole is ready to start with T001/T003.

## Critical dependencies

- W03-E04-S001 (DATA-07) hard-depends on W03-E01 reaching `accepted` — PLAN's own "do not schedule
  before it lands" language.
- W03-E01-S001's grant-ID field is a dependency of W03-E05-S001-T002 (durable audit).
- W03-E04-S001-T003's cache-invalidation acceptance criterion is deferred-linked to W05-E04-S002
  (SEC-04 epoch table, D-06) — soft dependency, does not block E04's own acceptance for the
  non-cache-related ACs.
- W03-E01-S001 through S003 depend on W00-E02-S003's ADR-ification of D-01 (`ADR-W00-E02-S003-001`);
  W03-E02-S001 depends on the same story's ADR-ification of D-07 (`ADR-W00-E02-S003-007`).

## Open decisions

DEC-Q1 (IdP `grant_id` claim contract) remains open and human-blocked at wave start; W03-E01-S001
proceeds against its documented safe default (REVIEW §F row 1 / MATRIX CS-07) rather than waiting
for resolution. See `wave.md` "Assumptions" and W03-E01-S001's `story.md` "Assumptions" section.

## Open risks

See `risks.md`.

## Artifact completeness

0/8 story-level artifact sets populated.

## Evidence completeness

0 evidence records registered.

## Review state

Planning reviewed as part of this status update. No implementation or independent-review evidence
yet.

## Exit-gate readiness

Not ready. 0 of 8 stories accepted. Wave remains `planned` pending W02 acceptance per mandate §15.
