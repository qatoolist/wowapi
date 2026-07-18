# Integration Credential Rotation (CA-14 residual)

This runbook is the wowapi-specific procedure for rotating a **per-tenant (or platform-default)
integration-provider credential** with zero downtime. It closes the CA-14 residual: the generic
secret-handling paragraph in the blueprint did not spell out the concrete, wowapi-specific steps for
rotating a provider credential that a tenant row references.

It is domain-neutral: "provider" here means any external-provider adapter a module registers
(payment, messaging, identity, storage, device — see `foundation/integration/integration.go` `validKinds`),
not any particular society/business concept.

## How a provider credential is stored

A provider credential is **never** stored as plaintext. Each row in `integration_providers` holds only a
**reference** to a secret, in the `credential_ref` column:

- Schema: `integration_providers.credential_ref text` — a secret-provider key, never plaintext
  (defined in the clean kernel baseline).
- The reference is a `secretref://<provider>/<path>` string
  (`kernel/secrets/secrets.go:17` `Scheme`, `:29` `ParseRef`).
- The write path rejects anything that is not a `secretref://` reference, so plaintext cannot enter:
  `if in.CredentialRef != "" && !secrets.IsRef(in.CredentialRef) { … "credential_ref must be a
  secretref:// reference, never plaintext" }` (`foundation/integration/store.go:54`).
- The row is a platform+tenant hybrid: `tenant_id` NULL is the platform default; a non-NULL `tenant_id`
  is a tenant override. RLS admits a tenant to its own rows plus platform defaults, and a tenant-bound
  session cannot forge a platform row (clean-baseline policy `integration_providers_tenant`).

At use time the kernel resolves the reference to a value on demand and wraps it in a structurally
redacted `config.Secret`:

- `integration.Config.Credential` is a `config.Secret` (`foundation/integration/integration.go:42`).
- `Store.Resolve` reads the row on the caller's `TenantDB` (RLS-scoped: a tenant override wins over the
  platform default), parses `credential_ref`, calls the secrets provider, and stores the result as
  `config.NewSecret(*credRef, val)` (`foundation/integration/store.go:88`–`128`). Resolution happens **per
  `Resolve` call**, not once at boot — so a new secret value is picked up on the next resolution
  without a redeploy.
- The secrets provider is the port `secrets.Provider.Resolve` (`kernel/secrets/secrets.go:48`). The
  framework ships the env adapter (`secretref://env/<VAR>` → process environment,
  `adapters/secrets/envprovider/envprovider.go`); cloud secret managers are product-supplied adapters
  implementing the same port.

## Redaction guarantees during rotation

The plaintext credential exists in memory only between `secrets.Provider.Resolve` and the point of use.
Everywhere else it is a `config.Secret`, which redacts on **every** standard rendering path — so logs,
`%v`/`%+v`/`%#v`/`%q`, JSON dumps, and `slog` records emit `[redacted:<ref>]`, never the value
(`kernel/config/secret.go`: `String`, `Format`, `GoString`, `MarshalJSON`, `MarshalText`, `LogValue`).
The raw value is reachable only via `Secret.Reveal`, whose call sites are restricted by boundary lint to
`adapters/`, `app/`, and tests (`kernel/config/secret.go:25`). Consequences for rotation:

- The `credential_ref` string is **safe to log and to print in dumps** — it names the secret, it is not
  the secret. `secretref://…` values are safe to echo (`secrets.Ref.String`, `kernel/secrets/secrets.go:43`).
- A malformed reference passed where a reference was expected is echoed as `****`, never verbatim, in
  case it was itself a raw secret (`kernel/secrets/secrets.go:57`).
- Nothing in the rotation steps below prints a credential value; the SQL and Go touchpoints operate on
  the `credential_ref` only.

## Who may rotate

Provider config is behavior-changing and is kept **off the module role** (SEC-13): the module role
(`app_rt`) may only `SELECT` `integration_providers`; only `app_platform` may `INSERT`/`UPDATE` it
(`app_rt` receives SELECT only in the clean baseline). Every write path below therefore runs with
platform privilege (`kernel/database/txmanager.go:126` `Manager.Platform` on a pool connected AS
`app_platform`, e.g. `app/maintenance.go:25`).

## Zero-downtime rotation

Use **version-pinned references** (a distinct `<path>` per version, e.g. `.../v1`, `.../v2`) so the old
and new credential coexist during the cutover. Add-before-switch, switch, verify, then retire.

### 1. Add a new secret version at the provider / secret manager

Create the new credential value under a **new path**, leaving the current one in place.

- **env adapter** (`secretref://env/<VAR>`): set a new environment variable in the deployment (do not
  overwrite the current one), e.g. add `PROVIDER_CRED_V2` alongside `PROVIDER_CRED_V1`. Placeholder:
  ```
  PROVIDER_CRED_V2=<new-credential-value>   # set in your deployment's secret store, never in a config file
  ```
- **cloud secret manager** (`secretref://<provider>/<path>`): create the new secret/version at the new
  path your product adapter resolves (e.g. `secretref://aws/prod/<tenant>/<provider>/v2`).

The new reference is not yet used by any tenant, so this step is inert until step 2.

### 2. Point the tenant row at the new reference

Update the tenant's (or platform default's) `credential_ref` to the new `secretref://…`. Two equivalent,
in-repo touchpoints:

**A. Go write API (preferred)** — `integration.Store.Upsert`, run inside a platform transaction:

```go
// db is an app_platform DBTX (e.g. via database.Manager.Platform). For a tenant
// override, bind the tx to that tenant so RLS admits the write; omit TenantID for
// the platform default row.
_, err := intStore.Upsert(ctx, db, integration.UpsertIn{
    TenantID:      tenantID,                 // uuid.Nil → platform default (tenant_id NULL)
    Key:           "core.exampleprovider",   // module.name
    Kind:          "messaging",              // one of validKinds
    Settings:      settings,                 // unchanged non-secret config
    CredentialRef: "secretref://env/PROVIDER_CRED_V2",
})
```

`Upsert` re-validates the reference (`store.go:54`), bumps `version`, sets `status='active'`, and stamps
`updated_at`/`updated_by` on conflict (`foundation/integration/store.go:70`–`79`). This is the path that
guarantees no plaintext can be written.

**B. Direct SQL** as `app_platform` (when no product admin surface is wired). This bypasses the Go
validation, so you MUST supply a `secretref://…` value yourself:

```sql
-- Run as the app_platform role. Tenant override:
UPDATE integration_providers
   SET credential_ref = 'secretref://env/PROVIDER_CRED_V2',
       version        = version + 1,
       status         = 'active',
       updated_at     = now(),
       updated_by     = '<operator-actor-uuid>'
 WHERE key = 'core.exampleprovider'
   AND tenant_id = '<tenant-uuid>';        -- omit / use IS NULL for the platform default
```

Because `Resolve` re-reads the row and re-resolves per call (`store.go:88`), the next request/job for
that tenant uses the new credential — no redeploy, no restart.

### 3. Verify the switch

- **Programmatic:** re-run the provider's health probe. `Store.HealthChecks` resolves every configured
  provider visible to the tenant and calls each adapter's `HealthCheck`, returning `key → error`
  (`foundation/integration/store.go:130`); products surface this in readiness detail. A single provider can
  be checked via `Store.Resolve` + the adapter's `HealthCheck` (`foundation/integration/integration.go:52`).
- **Audit / trail:** confirm the row moved. The `integration_providers` row carries its own change trail
  — `version` (incremented), `updated_at`, and `updated_by`:
  ```sql
  SELECT key, tenant_id, credential_ref, version, updated_at, updated_by
    FROM integration_providers
   WHERE key = 'core.exampleprovider' AND tenant_id = '<tenant-uuid>';
  ```
  `credential_ref` is safe to display here — it is a reference, not the secret. Confirm it now shows the
  new `…/v2` path and that `version` advanced.

### 4. Retire the old secret

Only after step 3 confirms the new reference is live and healthy, remove the previous version at the
provider (delete/disable `PROVIDER_CRED_V1`, or destroy the old cloud secret version). Retiring it before
the switch is verified would break in-flight resolution; retiring it after guarantees the old material is
no longer resolvable.

## Rollback

If step 3 fails, re-run step 2 with the **previous** `secretref://…/v1` value (which still exists because
step 4 has not run yet). The switch is a single-row update and reverts the same way.

## Checklist

- [ ] New secret version created at a **new** path; current version untouched.
- [ ] Tenant/platform row `credential_ref` updated via `Store.Upsert` (or `app_platform` SQL) to the new `secretref://…`.
- [ ] `version` incremented; `updated_at`/`updated_by` stamped.
- [ ] Provider health / a live resolution verified against the new credential.
- [ ] Old secret version retired at the provider.
- [ ] No credential **value** appeared in any command, log, or dump — only `secretref://…` references.
