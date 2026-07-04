# Hardening H3 — Machine authentication — proof bundle

Plan: [../../hardening-plan.md](../../hardening-plan.md). Decision: [D-0071](../../decisions.md).

## S1 — machine authentication (API keys / service principals)

| Verdict | Fix |
|---|---|
| real (P0) — only OIDC user JWTs; non-human callers (gate devices, integrations) had no credential | `kernel/apikey`: issuable, **scoped**, rotatable, revocable, expirable keys over `api_keys` (migration 00019). Only `sha256(secret)` is stored; the public prefix is the lookup handle. `Issue`/`Revoke`/`List` are tenant-scoped; `Verify` is cross-tenant (app_platform — the tenant is unknown pre-auth). `apikey.Authenticator` satisfies the `httpx.Authenticator` port from H1 and maps a verified key to an `ActorSystem` carrying its tenant + scopes. |

**Authz integration (the flagged design decision).** Rather than couple machine keys to the
capacity/party model, `authz.Actor` gained a `Scopes []string` and `Evaluate` got a machine fast-path:
a machine (`ActorSystem`) actor is authorized when the permission is in its scopes — the scope acts like
an RBAC grant, placed right after the RBAC loop so **ABAC deny policies still override it**. Deny-by-
default holds (allow only on an explicit scope match), and internal system actors (relay/webhook) carry
no scopes, so their authority is unchanged — proven by a test.

Security details: secret compared in constant time (`crypto/subtle`), and the hash is compared even when
the prefix is unknown, so there is no timing oracle distinguishing "unknown key" from "wrong secret".
All verification failures return a single non-specific `KindUnauthenticated`. Token format
`wowapi_<prefix>_<secret>` is distinguishable from a JWT, so a composite authenticator can fall through
to OIDC for non-`wowapi_` bearers.

Tests: `kernel/apikey/apikey_test.go` — issue→verify round-trip (tenant + scopes); wrong secret / unknown
prefix / malformed token all unauthenticated; revoked key denied; expired key denied; Authenticator maps
a valid key to `ActorSystem`+scopes and passes over a JWT bearer (+ compile-time `httpx.Authenticator`
assertion). `kernel/authz/machine_scope_test.go` — in-scope allowed (`machine_scope`), out-of-scope
denied, scopeless system actor denied. Gate: 0 FAIL, 0 SKIP, 82 packages; boundary lint + 00019
reversibility pass; all pre-existing authz tests still green with the `Scopes` addition.

Follow-ups (documented): rotation convenience (issue-new + revoke-old is available today via the two
calls); a `wowapi apikey` CLI; per-key rate limits (composes with S2 `KeyByActor`).
