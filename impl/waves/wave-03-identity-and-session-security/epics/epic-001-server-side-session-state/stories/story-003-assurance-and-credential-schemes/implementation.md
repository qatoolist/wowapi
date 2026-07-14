---
id: IMPL-W03-E01-S003
type: implementation-record
parent_story: W03-E01-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Implementation record — W03-E01-S003

## What was actually implemented

SEC-01 T6 (assurance freshness) and T7 (credential-scheme distinction).

- `auth.Claims` now carries `AuthTime` and `ACR`.
- `authz.Actor` now carries `AuthTime`, `ACR`, and `CredentialScheme`.
- `auth.Verifier.Actor` propagates these fields and assigns `CredentialUser` to JWT-authenticated
  human actors.
- `apikey.Authenticator.Authenticate` assigns `CredentialAPIKey` to verified API-key actors.
- `authz.StepUpPolicy` gained `MaxAge`; `authz.Options` gained `StepUpMaxAge`.
- The evaluator enforces step-up freshness before the AMR check and rejects stale/zero `AuthTime`
  with `Reason = "step_up_freshness_required"`.
- The evaluator enforces credential-scheme scoping via `Permission.AllowedSchemes`, rejecting
  mismatched schemes with `Reason = "credential_scheme_mismatch"`.
- Backward-compatible scheme derivation from `ActorKind`/`Scopes` supports actors constructed
  without an explicit scheme.

## Components changed

`kernel/auth`, `kernel/authz`, `kernel/apikey`, `testkit`.

## Files changed

- `kernel/auth/auth.go`
- `kernel/authz/authz.go`
- `kernel/authz/registry.go`
- `kernel/authz/evaluator.go`
- `kernel/apikey/apikey.go`
- `testkit/auth.go`
- `kernel/authz/assurance_freshness_test.go` (new)
- `kernel/authz/credential_scheme_test.go` (new)
- `kernel/auth/assurance_internal_test.go` (new)
- `kernel/apikey/apikey_test.go`

## Interfaces introduced or changed

- `auth.Claims`: added `AuthTime`, `ACR`.
- `authz.Actor`: added `AuthTime`, `ACR`, `CredentialScheme`.
- `authz.CredentialScheme` type and constants.
- `authz.Permission`: added `AllowedSchemes`.
- `authz.StepUpPolicy`: added `MaxAge`.
- `authz.Options`: added `StepUpMaxAge`.
- `testkit.TokenIssuer`: added `WithAuthTime`, `WithACR`.

## Configuration changes

None.

## Schema or migration changes

None.

## Security changes

- Closes the "expired step-up" required test-class gap (SEC-01 T6).
- Prevents credential-scheme confusion at the permission layer (SEC-01 T7).

## Observability changes

None required.

## Tests added or modified

- `kernel/authz/assurance_freshness_test.go`
- `kernel/authz/credential_scheme_test.go`
- `kernel/auth/assurance_internal_test.go`
- `kernel/apikey/apikey_test.go` (assertion added)
- `docs/user-guide/auth.md` (documentation of assurance freshness and credential schemes)

## Commits

Local working changes only.

## Pull requests

None for this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

The story-local `CredentialScheme` mechanism is a candidate for reconciliation with DX-03
(W06-E01-S001). It is documented as provisional/internal.

## Known limitations

DB-backed tests in `./kernel/authz/...` and `./kernel/apikey/...` are skipped without
`DATABASE_URL`. All non-DB tests pass.

## Follow-up items

- Re-run DB-backed tests with `DATABASE_URL` set.
- Reconcile `CredentialScheme` with DX-03 (W06-E01-S001).

## Relationship to the approved plan

Matches `plan.md`; no deviations.
