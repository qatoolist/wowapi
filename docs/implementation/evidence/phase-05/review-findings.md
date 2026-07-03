# Phase 5 — Review Findings

One comprehensive critique agent reviewed the module-SDK / app-boot / seeds / contract-suite slice
(2026-07-03), covering architecture, boundaries, and seed/permission security, with live probes. It
confirmed the boot gates work (reproduced the unregistered-route-permission boot rejection) and the
import surface is clean, and found two high seed-ownership/privilege issues.

| ID | Sev | Finding | Resolution | Status |
|---|---|---|---|---|
| SEC-32 | high | seed role grants bypass module ownership — a role could grant a foreign module's permission (FK-satisfied in multi-module deploys) — reproduced | `seeds.validate` prefix-checks every `RoleSeed.Permissions` entry (D-0044); `TestLoadRejectsForeignRoleGrant` | **fixed** |
| SEC-33 | high | no production SeedSync path; contract ran Sync as superuser → SEC-13 grant boundary untested | testkit provisions an `app_platform` login + Platform pool; the contract syncs seeds under it, so a seed needing a grant app_platform lacks fails in the suite. A forgiving `app_tenant_id_or_null()` lets platform NULL-template writes pass RLS (D-0045) | **fixed** |
| SEC-34 | med | `granted_via` neither ownership- nor existence-checked | prefix-checked + must name a relationship type the bundle declares; `TestLoadRejects{Foreign,Dangling}GrantedVia`, `TestLoadAcceptsOwnedGrantedVia` | **fixed** |
| ARCH-47 | med | seed reconciliation insert-only — removed grants never pruned (privilege drift) | `Sync` deletes each role's grants not in the seed (D-0044); contract's checksum idempotency guards it | **fixed** |
| ARCH-48 | med | contract RLS check evadable by table naming; zero-table passes | RLS check is now before/after diff-based over the tables the migration actually created (excl. goose); zero-tables-after-migrate fails (D-0046) | **fixed** |
| ARCH-49 | med | contract omits AST cross-import check + asserts seed no-ERROR not no-CHANGE | seed idempotency now asserts a catalog checksum is unchanged on re-sync; the cross-import check is covered by `scripts/lint_boundaries.sh` (documented — the runtime suite is not the place for AST import analysis) | **fixed / documented** |
| ARCH-50 | med | `consumer_test.go` `go mod tidy` network-dependent, not guarded | uses `GOFLAGS=-mod=mod` + the warm module cache; skips (not fails) when resolution needs network on a cold cache | **fixed** |
| ARCH-51 | low | `moduleContext` lazy-init guards inconsistent (perms/rtypes nil would silently fork; boot unguarded) | accepted for Phase 5: boot ALWAYS injects non-nil registries + boot state; the lazy `nil` paths are only reachable from the Phase-4 white-box context tests that call Logger/Config only. A fail-fast constructor is a follow-up (noted) | **accepted (tracked)** |
| ARCH-52 | low | nondeterministic seed merge order (map iteration) | boot iterates module seed names in sorted order | **fixed** |

Reviewer-confirmed solid (verified/reproduced): boot lifecycle order + full error accumulation;
the unregistered-route-permission gate fails boot (reproduced); the seed prefix check is airtight
(`req.` vs `requests.` off-by-prefix defended); `Config().Decode` strictness enforced end-to-end
(reject-invalid contract step proves it); `loggingAudit` impersonation logic correct; import-
direction clean and `replace` does not mask internal leakage (Go blocks external `internal/`
imports, so the scratch consumer exercises the real public surface); Port-before-provider is
structurally impossible (Register runs in dependency order).

Residual risk:
- app_platform seed privilege is now tested in the contract, but a real product's seed RUNNER
  (a `cmd`/`app.RunMigrate`-style entry that opens an app_platform connection at deploy time) is a
  Phase 10/CLI deliverable — Phase 5 proves the SQL works under that privilege, not the deploy glue.
- The durable authz-denial audit remains WARN-log-only until Phase 6 (documented).
- moduleContext fail-fast construction (ARCH-51) is a hardening follow-up.
