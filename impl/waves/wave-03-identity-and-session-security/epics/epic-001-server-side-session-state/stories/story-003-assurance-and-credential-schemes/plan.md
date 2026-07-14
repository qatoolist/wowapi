---
id: PLAN-W03-E01-S003
type: plan
parent_story: W03-E01-S003
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E01-S003

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

No new package. T6 extends the existing assurance/step-up code path in `kernel/auth` with freshness
enforcement bound to `auth_time`/`acr`/`amr`. T7 adds an explicit credential-scheme distinction at
the permission-check layer — a new, small classification (user/API-key/webhook/internal) consulted
wherever a permission declares a credential-scheme scope.

## Implementation strategy

1. Re-read `kernel/auth`'s current step-up/assurance code path at this story's actual start commit
   to confirm exactly what AMR plumbing already exists (per T6's own risk note) and what, if
   anything, currently checks `auth_time` freshness.
2. Implement T6: bind `auth_time`/`acr`/`amr` into the assurance model; add or correct freshness
   enforcement so a stale `auth_time` fails step-up even when `amr` itself is valid.
3. Write the step-up freshness test: stale `auth_time`, valid `amr` → step-up fails.
4. Re-read the current permission-check layer to confirm how (if at all) credential schemes are
   distinguished today.
5. Implement T7: an explicit `CredentialScheme` classification (user, API-key, webhook, internal)
   consulted at permission-check time, sufficient to satisfy T7's acceptance criterion. Build this
   as a scoped, story-local mechanism — not a pre-emptive implementation of DX-03's eventual DSL
   shape (see "Unresolved questions").
6. Write the credential-scheme distinction test: a permission scoped to `CredentialUser` rejects a
   valid, correctly-authenticated API-key actor.

## Expected package or module changes

`kernel/auth` (step-up/assurance code path; permission-check layer — exact files to be confirmed at
implementation time).

## Expected file changes where determinable

Not yet determinable with file:line precision — PLAN's own evidence citation for SEC-01 does not
name specific files/lines for T6/T7 (unlike T1/T2 for W03-E01-S001). To be confirmed by reading
`kernel/auth`'s step-up and permission-check code at this story's actual start commit.

## Contracts and interfaces

T7 likely introduces a new small type or enum (`CredentialScheme` or equivalent) and a way for a
permission declaration to scope itself to one or more schemes. Exact shape to be determined at
implementation time, built narrowly enough to satisfy T7's acceptance criterion without pre-empting
DX-03's eventual design.

## Data structures

None new beyond whatever minimal `CredentialScheme` classification T7 requires (see "Contracts and
interfaces").

## APIs

No public HTTP API surface change anticipated; this is an internal assurance/permission-check
change.

## Configuration changes

None anticipated.

## Persistence changes

None. No schema or data migration.

## Migration strategy

Not applicable.

## Concurrency implications

None beyond what the existing step-up/permission-check paths already handle.

## Error-handling strategy

Step-up freshness failure and credential-scheme mismatch both return distinguishable, testable
errors/rejection reasons — not merely a generic "forbidden" that would make the specific failure
mode untestable.

## Security controls

T6 closes the "expired step-up" gap in SEC-01's mandatory required test-class list. T7 prevents
credential-scheme confusion at the permission layer.

## Observability changes

Not mandated; a metric/log line distinguishing step-up rejection cause or credential-scheme
mismatch is a reasonable implementation-time addition, not required scope.

## Testing strategy

- Fail-first: confirm today's actual behavior (a stale `auth_time` with valid `amr` currently
  passes step-up, or is only partially checked) before the fix.
- Step-up freshness test: stale `auth_time`, valid `amr` → fails (T6, AC-W03-E01-S003-01).
- Credential-scheme test: `CredentialUser`-scoped permission rejects a valid API-key actor (T7,
  AC-W03-E01-S003-02).
- These map to the "expired step-up" leg of SEC-01's required test classes (PLAN §6 SEC-05).

## Regression strategy

Both tests, run in CI, are the regression guard for their respective behaviors.

## Compatibility strategy

Additive/stricter behavior for T6 (a previously-passing stale-`auth_time` request now correctly
fails); T7's credential-scheme distinction could newly reject a previously-passing
scheme-mismatched request — both are accepted, intended tightenings per this story's acceptance
criteria, not treated as regressions.

## Rollout strategy

Single PR/commit for T6; single PR/commit for T7; no phased rollout anticipated given the narrow,
additive scope of both changes.

## Rollback strategy

Revert T6's or T7's change independently if either surfaces an unexpected volume of newly-rejected,
previously-passing requests that indicate a fixture/assumption gap rather than a genuine security
gap — investigate before reverting, consistent with this repository's general practice of
diagnosing root cause rather than reflexively reverting.

## Implementation sequence

T6 before T7, per PLAN's own Depends-on column (T7 depends on T2-T6, i.e., including T6).

## Task breakdown

- **W03-E01-S003-T001** — Assurance freshness binding and step-up enforcement — SEC-01 T6.
- **W03-E01-S003-T002** — Credential-scheme distinction at the permission-check layer — SEC-01 T7.
- **W03-E01-S003-T003** — Independent review (mandate §14), scoped to the "expired step-up" required
  test class and the DX-03 cross-cut coordination note.

## Expected artifacts

Assurance-freshness code change; credential-scheme distinction implementation.

## Expected evidence

Step-up freshness test log; credential-scheme distinction test log.

## Unresolved questions

- **DX-03 cross-cut timing (the central open question for this story).** PLAN's own risk note for
  T7 says "Cross-cuts DX-03's `CredentialScheme` design — sequence together," but DX-03 (module DSL
  design) is scheduled at `W06-E01-S001` — materially later than this W03 story, per
  `requirement-inventory.md` row DX-03 (disposition `deferred`, "Design-investigation story only").
  This story cannot literally sequence together with a design that does not yet exist at this point
  in the programme. What must be determined, and by whom, at or before W06-E01-S001's execution: (a)
  whether this story's T7 `CredentialScheme` classification becomes DX-03's starting point, is
  superseded by it, or coexists as a lower-level mechanism DX-03's DSL wraps; (b) whether any
  interface this story introduces for T7 needs to be treated as provisional/internal rather than a
  stable public contract, specifically to avoid a breaking-change collision when DX-03 lands. This
  story does not resolve this question — it records it here per mandate §18 and flags it explicitly
  in `story.md`'s documentation requirements so a future DX-03 implementer is not surprised by this
  story's existing mechanism.
- Exact current file/line locations for the step-up/assurance code path and the permission-check
  layer — PLAN's own evidence citation for SEC-01 does not name specific files for T6/T7 (unlike
  T1/T2); to be confirmed by reading `kernel/auth` at implementation time.
- Exact current extent of AMR plumbing (T6's risk note: "already exists") — to be confirmed by
  reading the code, not assumed complete.

## Approval conditions

This plan is approved for implementation once: (a) the unresolved questions above are answered by a
first re-read of `kernel/auth`'s step-up and permission-check code at story start, (b) the DX-03
cross-cut question is at minimum recorded and flagged (not silently ignored) even if it cannot be
fully resolved before W06, and (c) the owner and reviewer are assigned.
