---
id: W03-E01-S003
type: story
title: Assurance freshness and credential-scheme distinction
status: accepted
wave: W03
epic: W03-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - SEC-01
depends_on:
  - W03-E01-S001
blocks: []
acceptance_criteria:
  - AC-W03-E01-S003-01
  - AC-W03-E01-S003-02
artifacts: []
evidence: []
decisions: []
risks: []
---

# W03-E01-S003 — Assurance freshness and credential-scheme distinction

## Story ID

W03-E01-S003

## Title

Assurance freshness and credential-scheme distinction

## Objective

Bind `auth_time`/`acr`/`amr` into the assurance model and enforce freshness for step-up, so a stale
`auth_time` with an otherwise-valid `amr` still fails step-up; and distinguish user/API-key/webhook/
internal credential schemes explicitly, so a permission scoped to `CredentialUser` rejects a valid
API-key actor. This is PLAN SEC-01 T6 and T7.

## Value to the framework

This story closes the "expired step-up" leg of SEC-01's mandatory required test-class list (PLAN §6
SEC-05: "token substitution, zero-tenant, stale membership, revoked capacity, expired step-up,
issuer/audience/key rotation, JWKS failure") and gives the framework's permission model a
first-class distinction between credential schemes, so a permission author can scope a check to
"human user only" and have that scoping actually enforced against an API-key or webhook-originated
actor, rather than an implicit assumption that goes unchecked.

## Problem statement

PLAN §5.2 SEC-01's task table, T6 and T7 rows, cited verbatim: T6 — "Bind `auth_time`/`acr`/`amr`
into assurance; enforce freshness for step-up," acceptance criterion "Stale `auth_time` with valid
`amr` still fails step-up," risk note "`AMR` plumbing already exists — additive, moderate risk." T7
— "Distinguish user/API-key/webhook/internal credential schemes explicitly," acceptance criterion
"Permission scoped to `CredentialUser` rejects a valid API-key actor," risk note "Cross-cuts DX-03's
`CredentialScheme` design — sequence together."

## Source requirements

SEC-01 (T6, T7). Cross-referenced: PLAN §6 SEC-05's required test-class list (this story's primary
responsibility is "expired step-up"); DX-03 (module DSL design, `requirement-inventory.md` row
DX-03: class ARCH/FUT, disposition `deferred`, target `W06-E01-S001`, notes "Design-investigation
story only") — see "Dependencies" below for how this story handles the cross-cut PLAN itself flags.

## Current-state assessment

Per PLAN §5.2's own evidence and risk notes (to be re-confirmed at this story's own execution
commit, consistent with this wave's other stories' pattern of re-running fail-first checks rather
than trusting a cited snapshot blindly):

- T6's own risk note states "`AMR` plumbing already exists — additive, moderate risk" — unlike
  W03-E01-S001's genuinely greenfield `identity_grant` schema, this story's T6 work extends an
  existing plumbing path (AMR handling) rather than building one from nothing. The exact current
  extent of that plumbing (what already binds `acr`/`amr` today, and what freshness enforcement, if
  any, currently exists for `auth_time`) is not itself detailed in PLAN's evidence citation for
  SEC-01 and must be confirmed by reading `kernel/auth/auth.go` and its step-up code path at this
  story's actual start commit.
- T7's own risk note flags a design cross-cut: "Cross-cuts DX-03's `CredentialScheme` design —
  sequence together." No current evidence citation in PLAN describes today's credential-scheme
  handling in detail beyond the implicit assumption that `CredentialUser`, API-key, webhook, and
  internal credential schemes are not currently distinguished at the permission-check layer (the
  acceptance criterion — "permission scoped to `CredentialUser` rejects a valid API-key actor" —
  would otherwise already hold).

## Desired state

`auth_time`, `acr`, and `amr` are bound into the framework's assurance model; a stale `auth_time`
with an otherwise-valid `amr` fails step-up rather than being silently accepted. User, API-key,
webhook, and internal credential schemes are distinguished explicitly at the permission-check
layer, such that a permission declared as scoped to `CredentialUser` rejects a valid, correctly-
authenticated API-key actor rather than treating all credential schemes as interchangeable.

## Scope

- T6: assurance-freshness binding and step-up enforcement.
- T7: explicit credential-scheme distinction at the permission-check layer.
- A best-effort `CredentialScheme` distinction sufficient to satisfy T7's acceptance criterion now,
  built with the explicit expectation that it may need reconciliation once DX-03's module DSL design
  (W06-E01-S001) exists — see "Dependencies" and `plan.md`'s "Unresolved questions."

## Out of scope

- Grant schema, membership enforcement, capacity selection, and the privileged-session resolver —
  W03-E01-S001 and W03-E01-S002's scope, both prerequisites to this story.
- DX-03's module DSL design itself (`CredentialScheme` as a first-class DSL concept) — that is
  W06-E01-S001's scope, a design-investigation story per `requirement-inventory.md` row DX-03. This
  story does not wait for DX-03 to exist; it builds a working credential-scheme distinction now,
  against SEC-01 T7's own acceptance criterion, without inventing or pre-empting DX-03's eventual
  DSL shape.
- Full coverage of every SEC-01 required test class — this story is responsible for "expired
  step-up" specifically; the others (token substitution, zero-tenant, stale membership, revoked
  capacity, issuer/audience/key rotation, JWKS failure) are covered across S001/S002 and, for
  JWKS-specific failure modes, W03-E02 (SEC-06).

## Assumptions

- **DX-03 cross-cut timing.** PLAN's own risk note for T7 says "Cross-cuts DX-03's `CredentialScheme`
  design — sequence together," but DX-03 is scheduled in W06 (`requirement-inventory.md` row DX-03:
  target `W06-E01-S001`), materially later than this W03 story. At this wave's point in the
  programme sequence, DX-03's `CredentialScheme` design does not yet exist. This story therefore
  cannot literally "sequence together" with DX-03 as PLAN's ideal-case note suggests — it must
  proceed with its own best-effort, SEC-01-T7-scoped credential-scheme distinction now, and record
  explicitly (per mandate §18: "state what must be determined during the story rather than
  inventing specifics") that this distinction may need reconciliation once DX-03's design lands in
  W06. This is not a decision this story makes unilaterally about DX-03's eventual shape — it is a
  scoping note about sequencing reality that PLAN's own cross-cutting language did not fully
  resolve. See `plan.md`'s "Unresolved questions" for the precise open question this creates.
- AMR plumbing (T6's stated prerequisite state) is assumed to exist in a form this story can extend
  additively; if the fresh re-read at story start finds it materially thinner than PLAN's risk note
  implies, that is recorded as a deviation, not silently absorbed into a larger redesign.

## Dependencies

Depends on W03-E01-S001 (PLAN: T6 depends on T2; T7 depends on T2-T6, i.e., transitively on T2 via
T6 plus directly on T6). No dependency on W03-E01-S002 — T6/T7 do not appear in PLAN's Depends-on
column for T4/T5, and T4/T5 do not depend on T6/T7 either, so S002 and S003 can in principle proceed
in parallel once S001 is accepted, though S003's T7 does depend on S003's own T6 (see `plan.md`'s
task breakdown). Cross-cuts DX-03 (W06-E01-S001) per PLAN's own note — see "Assumptions" above; this
is a coordination note, not a blocking `depends_on` entry, since DX-03 is scheduled materially later
than this wave.

## Affected packages or components

`kernel/auth/` (the step-up/assurance code path binding `auth_time`/`acr`/`amr`; the
permission-check layer distinguishing credential schemes — exact files to be confirmed at
implementation time).

## Compatibility considerations

Not flagged as breaking in PLAN's wowsociety-impact prose for SEC-01 specifically at the T6/T7
level (the overall SEC-01 wowsociety-impact prose focuses on T1/T2/T5's impersonation-flow breaking
change). T6's freshness enforcement is additive/stricter behavior (a request that previously passed
step-up with a stale `auth_time` would now correctly fail) — a behavioral tightening, not a
structural break, consistent with the pattern PLAN describes for T2 in W03-E01-S001. T7's explicit
credential-scheme distinction could, in principle, newly reject an API-key actor against a
`CredentialUser`-scoped permission where that combination was previously (incorrectly) allowed — to
be confirmed for wowsociety-specific impact at implementation time, since PLAN's own SEC-01
wowsociety-impact prose does not call out T6/T7 individually.

## Security considerations

T6 closes the "expired step-up" gap in SEC-01's mandatory required test-class list — a stale
`auth_time` currently either is not checked for freshness or is checked incompletely; this story
makes that check both present and enforced. T7 prevents credential-scheme confusion at the
permission layer — a security-relevant distinction, since a permission author's intent ("only a
human, step-up-capable user may perform this action") is meaningless if an API-key actor can
satisfy it.

## Performance considerations

Neither T6 nor T7 is expected to introduce a new database round-trip beyond what the assurance/
credential-scheme data already available on the verified token or resolved actor provides — both
are in-memory checks against already-available claim/actor data, not new external lookups, subject
to confirmation at implementation time.

## Observability considerations

Not mandated by this story's acceptance criteria; a metric or log line distinguishing step-up
rejections by cause (stale `auth_time` vs. other) is a reasonable implementation-time addition, not
required scope.

## Migration considerations

None. No schema or data migration.

## Documentation requirements

Document the assurance-freshness contract (what "stale" means for `auth_time`, how it interacts
with `acr`/`amr`) and the credential-scheme distinction (the four schemes — user, API-key, webhook,
internal — and how a permission declares which it accepts) in whatever documentation currently
covers the auth/permission model. Record the DX-03 cross-cut coordination note explicitly in this
documentation so a future DX-03 implementer knows this story's credential-scheme distinction exists
and may need reconciliation.

## Acceptance criteria

- **AC-W03-E01-S003-01**: A stale `auth_time` with an otherwise-valid `amr` still fails step-up,
  proven by a test.
- **AC-W03-E01-S003-02**: A permission scoped to `CredentialUser` rejects a valid, correctly-
  authenticated API-key actor, proven by a test.

## Required artifacts

- Assurance-freshness binding/enforcement code change (T6).
- Credential-scheme distinction implementation at the permission-check layer (T7).
See `artifacts/index.md`.

## Required evidence

- Step-up freshness test output (stale `auth_time`, valid `amr`, still fails).
- Credential-scheme distinction test output (`CredentialUser`-scoped permission rejects a valid
  API-key actor).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`:
`story.md` and `plan.md` complete, acceptance criteria numbered and measurable, dependency on
W03-E01-S001 recorded, the DX-03 cross-cut coordination note recorded rather than silently
resolved, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically covering the "expired step-up" required test class.

## Risks

None elevated beyond this epic's inherited risk set at this story's specific scope; T6/T7 are
additive/moderate-risk per PLAN's own risk notes, not greenfield-schema-risk like S001. The DX-03
cross-cut is tracked as an unresolved question in `plan.md`, not as a formal risk register entry,
since PLAN's own framing treats it as a sequencing note rather than a named risk.

## Residual-risk expectations

Once T6/T7 are implemented against their stated acceptance criteria, no residual risk is expected
beyond the DX-03 reconciliation noted above, which is explicitly deferred to W06-E01-S001's own
scope rather than resolved here.

## Plan

See `plan.md`.
