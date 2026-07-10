# Authentication & Authorization

wowapi is **deny-by-default**. Authentication (who are you?) is deployment-specific and supplied by the
product; authorization (may you do this?) is enforced by the kernel on every non-public route before your
handler runs. This page covers the actor model, the route gate, how to wire real authentication, and
record-level checks. (`kernel/httpx/authz_gate.go`, `kernel/authz/`, `kernel/apikey/`.)

## The two halves

| | Authentication | Authorization |
|---|---|---|
| Question | Who is calling? | Are they allowed? |
| Owner | **Your product** (OIDC/JWT/API-key → `Actor`) | The **kernel** (RBAC→ReBAC→ABAC + scopes) |
| Contract | `httpx.Authenticator` | `authz.Evaluator` + `RouteMeta.Permission` |
| Default | `DenyAllAuthenticator` → 401 | deny-by-default (no grant = refused) |

## The request gate

`Router.SecureHandler(auth, eval, txm)` wraps every route. For each **non-public** route, `gateRoute` runs
the deny-by-default chain *before* the handler (`kernel/httpx/authz_gate.go`):

```
Authenticate(r) → Actor         (401 KindUnauthenticated on failure)
   → bind tenant + actor into ctx and open the tenant tx
      → Evaluate(actor, RouteMeta.Permission, target)   (403 KindForbidden, or step-up challenge)
         → your handler runs, already authorized
```

A route the framework **cannot authenticate or authorize never reaches the handler**. Public routes
(`RouteMeta.Public: true`) are served directly.

## `RouteMeta` — the per-route contract

Every route declares metadata at registration; invariants are checked at **boot** (`kernel/httpx/router.go`):

```go
type RouteMeta struct {
    Permission string          // required permission key; empty ONLY if Public
    Public     bool            // opt out of authz (health, pre-verification webhooks)
    Scope      ScopeExtractor  // derive the authz target from the request (optional)
    Idempotent bool            // enable idempotency-key handling for unsafe methods
    Sensitive  bool            // force an audit record even for reads
}
```

Rules enforced at boot:

- `Public: true` **and** a `Permission` set → boot error ("choose one").
- Not public **and** no `Permission` → boot error. Every route is explicitly one or the other.

```go
r.Handle("GET",  "/requests/healthz", httpx.RouteMeta{Public: true}, h.Healthz)
r.Handle("POST", "/requests",         httpx.RouteMeta{Permission: "requests.request.create"}, h.Create)
```

Permission keys are conventionally `<module>.<resource>.<action>` and **must be declared** in the module's
`Permissions()` registry — an unregistered permission can never be authorized (deny-by-default,
boot-validated).

## The actor model

Authentication produces an `authz.Actor` (`kernel/authz/authz.go`):

```go
type Actor struct {
    Kind       ActorKind  // "user" | "system" | "webhook"
    UserID     uuid.UUID
    CapacityID uuid.UUID  // the user's active capacity in the tenant (zero for system/webhook)
    System     string     // e.g. "outbox-relay", "webhook:payments"
    TenantID   uuid.UUID
    ImpersonatorUserID uuid.UUID // set when support impersonates; both identities audited
    BreakGlass bool       // operating under an activated break-glass grant (audited + bannered)
    Scopes     []string   // explicit permission set of a machine principal (ActorSystem only)
    AMR        []string   // auth-methods-references from the IdP (e.g. ["pwd","mfa"]) — drives step-up
}
```

The actor carries its **own tenant** — this is how the gate binds `SET LOCAL app.tenant_id` for the
transaction, so RLS and authz share one tenant scope.

## Wiring real authentication (required before serving business routes)

A freshly scaffolded product uses `DenyAllAuthenticator`, so **every business route returns 401** until you
implement `httpx.Authenticator`:

```go
type Authenticator interface {
    Authenticate(r *http.Request) (authz.Actor, error)  // return a KindUnauthenticated error for 401
}
```

Your implementation typically: verifies the bearer token (OIDC/JWT), resolves the tenant, maps the subject
+ capacity, and populates `AMR`/`Scopes`. Then pass it to `SecureHandler` where the app builds the mux.

> **This is an intentional product responsibility, not a gap in the framework.** The identity source
> (which IdP, how tenants are resolved) is deployment-specific, so the kernel never hardcodes it. The
> framework guarantees the *default is safe* (deny-all) rather than guessing an identity strategy.

### Machine principals (API keys)

The kernel ships an API-key authenticator (`kernel/apikey`): `apikey.Authenticator` implements
`httpx.Authenticator` and yields an `ActorSystem` actor whose `Scopes` authorize like RBAC grants but remain
subject to ABAC deny policies. Use it for service-to-service callers; use your OIDC authenticator for humans.
You can compose authenticators (try API key, then bearer token) in your product's `Authenticator`.

## Authorization model

The `Evaluator` is deny-by-default and layered — **RBAC → ReBAC → ABAC**, plus machine scopes and step-up
(`kernel/authz/authz.go`):

```go
type Evaluator interface {
    Evaluate(ctx, db database.TenantDB, a Actor, perm string, t Target) (Decision, error)
    Filter(ctx, db database.TenantDB, a Actor, perm string, rt string) (ListFilter, error)
}
```

- **`Evaluate`** decides a single action on a `Target` (tenant / org / resource-type / resource scope).
- **`Filter`** returns a `ListFilter` so list endpoints embed authorization **in SQL** — you fetch only
  visible rows instead of load-then-filter.
- Both take the caller's `TenantDB`, so the authz read runs in the **same transaction and MVCC snapshot** as
  the business write (an authz check right after a mirror-row write sees that write).

`Decision.Reason` names the matched grant/policy (safe to log for audit, e.g.
`role:requests.org.approver`, `policy:deny_locked`).

### Route-level vs. record-level checks

The gate already enforced the route's `RouteMeta.Permission` before your handler ran. Inside a handler you
make **fine-grained, resource-scoped** checks against a concrete target when needed — that's what the
injected `authz.Evaluator` (`mc.Authz()`) is for:

```go
dec, err := h.authz.Evaluate(ctx, db, actor, "requests.request.approve",
    authz.Target{Scope: authz.ScopeResource, Resource: resource.Ref{Type: "requests.request", ID: id}})
if err != nil { return err }
if !dec.Allowed { return kerr.E(kerr.KindForbidden, "permission_denied", "not allowed to approve this request") }
```

### Step-up / MFA

When an actor *would* be allowed but the permission demands an elevated auth factor they haven't satisfied,
`Decision.StepUpRequired` is set. The HTTP gate turns this into a **re-authentication challenge** — 401 with
`WWW-Authenticate: Bearer error="insufficient_user_authentication", step_up="mfa"` — not a flat 403.

**Declaring a step-up permission** is a seed field, not out-of-band wiring:

```yaml
permissions:
  - key: identity.impersonation.assign
    description: assign an impersonation grant
    step_up: true
```

`kernel/seeds` strict-decodes `step_up` (a typo fails the load), `app.Boot` propagates it into
`authz.Permission.StepUp` when it registers your seed's permissions, and `seeds.Sync` persists it to
`permissions.step_up` — re-syncing after you flip the flag updates the existing catalog row (idempotent,
not insert-only).

**Satisfying the factor** flows from your IdP token straight through, with nothing to reparse: the standard
`amr` claim (RFC 8176, e.g. `["pwd","mfa"]`) is a field on `auth.Claims`, and `Verifier.Actor` copies it onto
`authz.Actor.AMR`, which the evaluator checks. A product authenticator built on `kernel/auth` gets step-up for
free — it does not need to re-verify the bearer token to recover `amr` itself.

**Implementing the factor that produces `amr`** is a separate concern from step-up itself, and lives in
`kernel/mfa`: reusable, standards-compliant TOTP (RFC 6238) and HOTP (RFC 4226) code generation/verification,
numeric OTP generation with salted constant-time hashing, and pure TTL/attempt-limit challenge-policy helpers,
plus `Sender` delivery-port interfaces (SMS/email) with log/fake adapters. There is no import relationship
between the two packages — `kernel/authz` never imports `kernel/mfa` and vice versa. The connection is a
convention: a product's own auth/MFA service uses `kernel/mfa` to verify a TOTP or delivered-OTP code, and
*on success* appends the corresponding factor (e.g. `"mfa"`, `"otp"`) to the `amr` slice it puts on the
authenticated actor — at which point `kernel/authz`'s step-up check (above) is satisfied. `kernel/mfa`
deliberately does not know about `amr`, permissions, or which actions require which factor: enrollment UX,
factor storage schema, delivery-provider selection, and factor-to-permission policy are all product-owned.
See the `kernel/mfa` package doc for the exact API (`GenerateTOTPSecret`, `TOTPCodeAt`/`VerifyTOTPAt`,
`HOTPCode`, `GenerateOTPCode`/`HashOTPCode`/`VerifyOTPCode`, `ChallengePolicy`, `Sender`).

## Testing auth

`testkit` issues real signed tokens so you can exercise the gate end to end
(`testkit/auth.go`):

```go
ti := testkit.NewTokenIssuer()                 // local RSA keypair
ks := ti.KeySource()                           // wire into your verifier
tok := ti.Issue(subjectID, tenantID, capacityID,
    testkit.WithAudience("myapp"), testkit.WithBreakGlass(true)) // options: WithIssuer/WithAudience/WithExpiry/WithImpersonator/WithBreakGlass(bool)/WithAMR(...string)
req.Header.Set("Authorization", "Bearer "+tok)
```

See [Testing](testing.md) for the full harness.

## Common problems

| Symptom | Cause | Fix |
|---|---|---|
| Every business route → 401 | `DenyAllAuthenticator` still wired | Implement + wire a real `httpx.Authenticator`. |
| Route → 403 for a valid user | permission not granted, or not declared in `Permissions()` | Grant the role/permission; ensure the permission is registered. |
| Boot fails: "Public but also sets Permission" | route marked both | Pick one in `RouteMeta`. |
| Boot fails: route has no permission | non-public route missing `Permission` | Add a `Permission` or mark `Public`. |
| 403 where you expected a re-auth prompt | `StepUpRequired` not handled by the client | Handle the step-up challenge; ensure `AMR` reflects satisfied factors. |
| List returns rows the user shouldn't see | not using `Filter` | Build the query from `Evaluator.Filter`'s `ListFilter`. |

Next: [Validation & errors](validation-errors.md) · [Testing](testing.md).
