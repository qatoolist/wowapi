---
id: W03-E01-S003-T002
type: task
title: Credential-scheme distinction (SEC-01 T7)
status: todo
parent_story: W03-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E01-S003-T001
acceptance_criteria:
  - AC-W03-E01-S003-02
artifacts:
  - ART-W03-E01-S003-002
evidence:
  - EV-W03-E01-S003-002
---

# W03-E01-S003-T002 — Credential-scheme distinction (SEC-01 T7)

## Task Definition

### Task objective

Distinguish user/API-key/webhook/internal credential schemes explicitly at the permission-check
layer, so a permission scoped to `CredentialUser` rejects a valid, correctly-authenticated API-key
actor.

### Parent story

W03-E01-S003 — Assurance freshness and credential-scheme distinction.

### Owner

unassigned

### Status

todo

### Dependencies

W03-E01-S003-T001 — PLAN's own Depends-on column for T7 names T2-T6, i.e., including T6.

### Detailed work

1. Read the current permission-check layer at this task's actual start commit to confirm how (if
   at all) credential schemes are distinguished today.
2. Design and implement an explicit `CredentialScheme`-style classification (user, API-key, webhook,
   internal) consulted at permission-check time. Build this as a scoped, story-local mechanism
   sufficient to satisfy this task's acceptance criterion — not a pre-emptive implementation of
   DX-03's eventual module-DSL `CredentialScheme` design (see `plan.md`'s "Unresolved questions" —
   this is the cross-cut PLAN's own risk note flags: "Cross-cuts DX-03's `CredentialScheme` design —
   sequence together").
3. Write the credential-scheme distinction test: a permission declared as scoped to
   `CredentialUser` rejects a valid, correctly-authenticated API-key actor.
4. Document the mechanism explicitly as a candidate for reconciliation with DX-03 (W06-E01-S001) —
   per `plan.md`'s unresolved-question framing — so a future DX-03 implementer is not surprised by
   this mechanism's existence.

### Expected files or components affected

`kernel/auth` (permission-check layer — exact file to be confirmed at implementation time).

### Expected output

A permission scoped to `CredentialUser` rejects a valid API-key actor; the mechanism is documented
as a DX-03 reconciliation candidate.

### Required artifacts

ART-W03-E01-S003-002 (credential-scheme distinction implementation).

### Required evidence

EV-W03-E01-S003-002 (credential-scheme distinction test report).

### Related acceptance criteria

AC-W03-E01-S003-02.

### Completion criteria

The credential-scheme distinction test passes: `CredentialUser`-scoped permission rejects a valid
API-key actor.

### Verification method

Direct test execution, logged output retained as evidence.

### Risks

PLAN's own risk note: "Cross-cuts DX-03's `CredentialScheme` design — sequence together" — this
task cannot literally sequence with a design that does not yet exist (DX-03 is W06-scoped); the
risk is scoped to a future reconciliation cost, not a current-implementation blocker. See `plan.md`
"Unresolved questions."

### Rollback or recovery considerations

Revert if the credential-scheme distinction incorrectly rejects a legitimate scheme combination not
anticipated by this task's fixture set — investigate the specific misclassification before
reverting.

## Implementation Record

### What was actually implemented

Explicit credential-scheme distinction at the permission-check layer.

- Added `CredentialScheme` type with `CredentialUser`, `CredentialAPIKey`, `CredentialWebhook`, and
  `CredentialInternal` values in `kernel/authz/authz.go`.
- Added `CredentialScheme` field to `authz.Actor`.
- Added `AllowedSchemes []CredentialScheme` to `authz.Permission`.
- Added `defaultCredentialScheme` helper that derives a scheme from `ActorKind`/`Scopes` when the
  actor was constructed without an explicit scheme, preserving backward compatibility.
- The evaluator now rejects an actor whose scheme is not in the permission's `AllowedSchemes` with
  `Reason = "credential_scheme_mismatch"`.
- `auth.Verifier.Actor` sets `CredentialUser` for OIDC/JWT human actors.
- `apikey.Authenticator.Authenticate` sets `CredentialAPIKey` for verified API-key actors.

### Components changed

`kernel/authz`, `kernel/auth`, `kernel/apikey`.

### Files changed

- `kernel/authz/authz.go`
- `kernel/authz/registry.go`
- `kernel/authz/evaluator.go`
- `kernel/auth/auth.go`
- `kernel/apikey/apikey.go`
- `kernel/authz/credential_scheme_test.go` (new)
- `kernel/apikey/apikey_test.go`

### Interfaces introduced or changed

- New type `authz.CredentialScheme` and constants.
- `authz.Actor.CredentialScheme`.
- `authz.Permission.AllowedSchemes`.

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

Prevents credential-scheme confusion: a permission scoped to `CredentialUser` now rejects a valid
API-key actor.

### Observability changes

*Not applicable.*

### Tests added or modified

- `kernel/authz/credential_scheme_test.go`: user-scoped permission rejects API-key, allows user,
  backward compatibility, scheme derivation, webhook/internal scoping.
- `kernel/apikey/apikey_test.go`: asserts API-key actor carries `CredentialAPIKey`.
- `docs/user-guide/auth.md`: documents the credential-scheme distinction and the DX-03
  reconciliation candidate note.

### Commits

*Local working changes; no commit authored in this session.*

### Pull requests

*Not applicable for this session.*

### Implementation dates

2026-07-13.

### Technical debt introduced

A story-local `CredentialScheme` mechanism is now in place. It is explicitly documented as a
DX-03 (W06-E01-S001) reconciliation candidate and should be treated as provisional/internal until
DX-03's module-DSL design lands.

### Known limitations

DB-backed tests in `./kernel/authz/...` and `./kernel/apikey/...` are skipped when `DATABASE_URL`
is unavailable. All non-DB tests pass.

### Follow-up items

- Re-run DB-backed tests with `DATABASE_URL` set.
- Reconcile this story's `CredentialScheme` with DX-03 (W06-E01-S001).

### Relationship to the approved plan

Matches `plan.md`: T7 introduced an explicit, scoped credential-scheme distinction without
pre-empting DX-03's eventual DSL shape.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E01-S003-02 | Valid, correctly-authenticated API-key actor attempts a `CredentialUser`-scoped permission check | Local dev or CI | Permission check rejects the API-key actor | unit/adversarial test report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
