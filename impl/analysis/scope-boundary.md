---
id: ANALYSIS-SCOPE-BOUNDARY
type: analysis
title: Scope boundary — framework vs product, the five PROD items and the test applied to each
status: complete
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Scope boundary

Per mandate §2.3 ("Framework-first scope"). The mandate requires: do not introduce
housing-society-specific, legal-domain-specific, product-specific, or application-specific concepts
into the framework unless the source architecture explicitly requires a generic abstraction
supporting such use cases. Where a requirement belongs to a downstream product rather than the
framework, the mandate's test is to: **record it; classify it as product-level; exclude it from
framework implementation; provide the rationale; identify any generic framework capability that must
exist to support it.**

This is the standing architectural rule CLAUDE.md itself states: "wowapi is a domain-neutral Go
platform kernel; wowsociety is the product built on it" — society/committee/policy-specific
vocabulary stays out of the kernel. This document enforces that rule for every product-level item
surfaced by the source analysis.

## Anchor table — reproduced verbatim from requirement-inventory.md table D

| ID | Title | Rationale | Enabling framework capability |
|---|---|---|---|
| PROD-01 | wowsociety `policy_override` composite FK | Product schema fix | DATA-01 T1 (parent unique index) + DATA-09 protocol |
| PROD-02 | wowsociety `kernel/mfa` import migration (5 identity files) | Product code migration | FBL-01 re-home ships deprecated forwarding shim |
| PROD-03 | wowsociety readiness/timeout backports to committed main.go | Product hand-edit | DX-07 T1 + FBL-09 fix the templates |
| PROD-04 | SEC-01 impersonation cutover (whoami/impersonation/tests) | Product auth flow rework | SEC-01 T1/T5 grant contract + coordinated rollout plan |
| PROD-05 | DATA-08 W6 staging audit re-verification before version bump | Product compliance drill | hash_version branch verification (D-04) |

## Applying the mandate §2.3 test to each item

### PROD-01 — wowsociety `policy_override` composite FK

**Why this is product-level, not framework-level:** the defective row — `policy_override`, with its
`rule_version_id` foreign key missing the tenant-scoping composite that the framework's own tenant-FK
pattern requires — is a table that exists **inside wowsociety's own schema**, not inside any
`kernel/*` migration. The framework's job (DATA-01) is to define and enforce the composite-tenant-FK
*pattern* and provide the online-migration *protocol* (DATA-09) generically; applying that pattern to
a specific product table that encodes wowsociety's own `policy_override` domain concept is a
product-side schema fix wowsociety's own migration must perform, using the framework's tooling. The
framework must never special-case `policy_override` by name.

- **Record:** done — PROD-01 in the anchor table above, sourced from `requirement-inventory.md`
  DATA-01's own note ("wowsociety has own instance (product-level, PROD-01)").
- **Classify:** product-level.
- **Exclude:** DATA-01's framework-side task breakdown (target W02-E02-S001..S002) does not touch
  wowsociety's schema; it defines the composite-FK pattern and migration primitives only.
- **Rationale:** the specific FK belongs to a wowsociety-domain table (`policy_override`), which is
  application/product schema, not kernel schema.
- **Enabling framework capability:** DATA-01 T1 (the parent unique-index pattern the composite FK
  depends on) + the DATA-09 online expand/backfill/validate/contract protocol (so wowsociety can apply
  the fix without a breaking single-shot migration).

### PROD-02 — wowsociety `kernel/mfa` import migration (5 identity files)

**Why this is product-level, not framework-level:** FBL-01 (kernel re-home) is a framework-internal
package-layout correction — moving `kernel/mfa` and 8 sibling packages into a `foundation/` layer per
the corrected four-level architecture (REVIEW §J). The framework's job is to perform that move and
ship a deprecated forwarding shim so existing importers don't break at the moment of the move. But
wowsociety's own `internal/modules/identity/` package (5 files, per REVIEW §29 answer 17's grep-
verified fact) imports the old `kernel/mfa` path directly — updating those 5 wowsociety files to the
new import path is a change to wowsociety's source code, which the framework programme cannot make on
wowsociety's behalf and must not reach into wowsociety's repository to do.

- **Record:** done — PROD-02 in the anchor table, sourced from REVIEW §29 answer 17 ("FBL-01
  `kernel/mfa` re-home — a scoped, auth-critical migration across 5 identity-module files — the only
  re-homed package wowsociety imports; the other 8 are zero-impact").
- **Classify:** product-level.
- **Exclude:** FBL-01's framework-side task breakdown (target W05-E05-S001..S002) performs the
  re-home and ships the shim; it does not edit any wowsociety file.
- **Rationale:** the 5 files needing an import-path update live in wowsociety's own
  `internal/modules/identity/`, outside the wowapi repository boundary entirely.
- **Enabling framework capability:** FBL-01's re-home, specifically its deprecated forwarding shim,
  which is the generic mechanism that keeps wowsociety compiling during the coordination window before
  wowsociety's own migration lands.

### PROD-03 — wowsociety readiness/timeout backports to committed `main.go`

**Why this is product-level, not framework-level:** DX-07 (truthful readiness/config diagnostics) and
FBL-09 (HTTP server timeouts + CSRF body bound) both fix their respective defects in the framework's
*generated scaffold templates* — the code a new module gets when the CLI generator runs. wowsociety's
own `cmd/api/main.go`, however, is a **hand-edited, already-committed file** generated from an older
template version; it does not get automatically regenerated when the templates change. Per PLAN §7's
own wowsociety-specific findings list: DX-07 is "inherited-but-latent, needs a manual backport since
generated code isn't regenerated." Backporting the fix into wowsociety's committed `main.go` is a
manual, product-side edit.

- **Record:** done — PROD-03 in the anchor table, sourced from PLAN §7's DX-07 wowsociety-impact note
  and `requirement-inventory.md` FBL-09's "template-delivery model (wowsociety backport = PROD-03)".
- **Classify:** product-level.
- **Exclude:** DX-07 T1 and FBL-09's framework-side work (targets W04-E04-S003 and W01-E03-S001
  respectively) fix the *templates* only, so future-generated scaffolds are correct; neither touches
  wowsociety's existing `main.go`.
- **Rationale:** the defect only persists in wowsociety because wowsociety's `main.go` is a static,
  hand-maintained artifact that diverged from the template at generation time — the framework cannot
  retroactively regenerate a product's already-customized entry point.
- **Enabling framework capability:** DX-07 T1 (readiness/migration-currency check in the template) +
  FBL-09's fix (timeouts + CSRF body bound in the template) — both are the generic template-delivery
  model wowsociety's own maintainers apply by hand.

### PROD-04 — SEC-01 impersonation cutover (whoami/impersonation/tests)

**Why this is product-level, not framework-level:** SEC-01 builds the framework's generic server-side
`identity_grant` table and resolver — a framework capability with no knowledge of "impersonation" as a
product concept. wowsociety currently has its own `identity_impersonation_session` table and UX flow
(whoami endpoint, impersonation start/stop, its own test suite) that is wowsociety-specific product
behavior built on top of whatever session model the framework provides today. Per D-01 (ratified,
REVIEW §F Q2): "framework owns grant validity/expiry/revocation; wowsociety keeps its table for
product UX/audit only." Migrating wowsociety's impersonation flow to consume the new grant contract —
updating its whoami endpoint, its impersonation start/stop handlers, and its own tests — is a
wowsociety-side auth-flow rework coordinated against, but not performed by, the framework programme.

- **Record:** done — PROD-04 in the anchor table, sourced from PLAN §7's wowsociety-specific findings
  list ("Breaking, high-severity: SEC-01 (impersonation flow)") and REVIEW §F Q2 / decision D-01.
- **Classify:** product-level.
- **Exclude:** SEC-01's framework-side task breakdown (target W03-E01-S001..S004) builds the grant
  table/resolver/Verifier.Actor integration only; it does not modify wowsociety's
  `identity_impersonation_session` table, whoami endpoint, or impersonation tests.
- **Rationale:** "impersonation" as a named UX/audit concern is wowsociety's own product vocabulary;
  the framework's concern stops at generic grant validity/expiry/revocation (D-01's explicit
  boundary).
- **Enabling framework capability:** SEC-01 T1 (the grant table) + T5 (the waiver/claim-contract
  mechanism, safe-defaulted per DEC-Q1) — plus a coordinated rollout plan since this is a BREAKING
  change for wowsociety.

### PROD-05 — DATA-08 W6 staging audit re-verification before version bump

**Why this is product-level, not framework-level:** DATA-08's Wave-6 work widens the framework's
audit-hash contract with a `hash_version` discriminator (D-04) so historical rows remain verifiable
after the hash contract changes. But wowsociety has **live audit rows in production/staging** —
including its own impersonation/policy audit trail — that were written under the old hash contract.
Before wowsociety can safely bump to a wowapi version containing the widened hash contract, wowsociety
must run its own staging drill to confirm its live rows verify correctly under the new
`hash_version`-aware verification logic. That drill runs against wowsociety's own staging data and is
wowsociety's own pre-upgrade compliance gate, not a framework deliverable.

- **Record:** done — PROD-05 in the anchor table, sourced from PLAN §5.6 DATA-08 W6-T1 risk note
  ("including wowsociety's live impersonation/policy audit rows") and `requirement-inventory.md`
  DATA-08's D-04 reference.
- **Classify:** product-level.
- **Exclude:** DATA-08's framework-side task breakdown (target W04-E04-S001..S002, specifically the
  W6-T1 hash-widening task) implements and verifies the `hash_version` migration and dual-branch
  verification logic against the framework's own test fixtures; it does not run against wowsociety's
  actual staging audit data.
- **Rationale:** the compliance risk (a live audit row silently failing to verify after an upgrade) is
  specific to wowsociety's own production/staging dataset, which the framework programme has no access
  to and no authority to drill against.
- **Enabling framework capability:** the `hash_version` branch-verification logic (D-04) itself is the
  generic capability wowsociety's drill exercises — the framework provides the versioned verification
  path; wowsociety proves its own data survives it.

## Framework-first principle (mandate §2.3, restated)

None of the five PROD items above introduce society/committee/policy-specific concepts into any
`kernel/*` or `foundation/*` package. In every case, the framework-side task (DATA-01/DATA-09,
FBL-01, DX-07/FBL-09, SEC-01, DATA-08) delivers a **generic capability** — a pattern, protocol,
template, contract, or migration primitive — that is domain-neutral and has no wowsociety-specific
naming or behavior baked in. The product-specific work (the actual `policy_override` table fix, the
actual import-path edits, the actual `main.go` backport, the actual impersonation-flow rework, the
actual staging drill) is explicitly excluded from framework implementation scope and lives entirely on
the wowsociety side of the repository boundary, tracked here only as a coordination item with a named
rationale and a named enabling framework capability, per the mandate's own required fields.

This is the same rule CLAUDE.md states as the project's standing architectural principle: "wowapi is a
domain-neutral Go platform kernel; wowsociety is the product built on it." This `impl/` programme's
scope boundary — PROD-01..05 recorded, classified, excluded, rationale'd, and cross-referenced to
their enabling framework capability — is the mechanical enforcement of that principle for this
specific implementation programme, and no story in any wave (`impl/waves/`) targets a PROD-0X ID as
its own deliverable; PROD items exist only so the coordination dependency is visible, not so the
framework programme performs the work.
