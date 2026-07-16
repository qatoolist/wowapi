---
id: W02-E02
type: epic
title: Tenant foreign-key integrity
status: partially-accepted
wave: W02
owner: W02FKVerAgg
reviewer: W02ReviewGate
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-16
source_requirements:
  - DATA-01
  - CS-18
depends_on:
  - W02-E01
stories:
  - W02-E02-S001
  - W02-E02-S002
decisions: []
risks:
  - RISK-W02-002
---

# W02-E02 — Tenant foreign-key integrity

## Epic objective

Close the confirmed tenant-FK integrity gap: 8 tenant-scoped child tables that today reference only
their parent's bare `id`, never `(tenant_id, id)`, so that row-level security's proof of the child
row's own tenant is backed by a database-enforced guarantee that the parent agrees, not merely by
convention. Deliver this by adding `UNIQUE (tenant_id, id)` on every referenced parent, a permanent
CI-gated catalog scanner, a mismatch audit against staging/prod-shaped data, composite `NOT VALID`
foreign keys validated via the online migration protocol this wave's W02-E01 builds, and seeded
cross-tenant negative tests under both platform and regular roles.

## Problem being solved

`requirement-inventory.md` row DATA-01 records: "Composite tenant FKs (T1–T8) | IMPL | P0 | planned
| W02-E02-S001..S002 | Dep DATA-09 T1–T5 for risky steps; wowsociety has own instance (product-level,
PROD-01)." PLAN's own DATA-01 evidence, quoted exactly: "8 tenant-scoped child tables
(persons/legal_entities/party_contacts/acting_capacities → parties; resources → organizations;
document_versions/document_access_grants/attachments → documents/document_versions) reference only
the parent's `id`, never `(tenant_id, id)`. RLS proves the child row's own tenant; nothing proves
parent and child agree." MATRIX CS-18 states the fail-first framing directly: "platform-role seeded
cross-tenant parent/child insert succeeds today; fails after." This is not a theoretical gap — it is
a confirmed structural hole in the tenant-isolation guarantee the framework's row-level-security
architecture is built to provide, graded P0 in every source that discusses it.

## Scope

- Adding/confirming `UNIQUE (tenant_id, id)` on every referenced parent table, built
  `CONCURRENTLY` (S001, PLAN DATA-01 T1).
- Building a tenant-FK catalog scanner that flags any tenant-table FK not composite on
  `(tenant_id, …)`, enumerating exactly the 8 known FKs with zero silent gaps (S001, PLAN DATA-01
  T2).
- Wiring the T2 scanner into a permanent CI gate that fails a new migration adding a single-column
  tenant FK (S001, PLAN DATA-01 T6 — sequenced early per PLAN's own note: "Cheapest, most durable
  part — do first if sequencing allows").
- A mismatch audit proving `child.tenant_id = parent.tenant_id` for every existing row, failing
  deployment on any mismatch (S002, PLAN DATA-01 T3).
- Adding composite FK `NOT VALID` for all 8 edges, gated on this wave's W02-E01 online-migration-
  protocol acceptance (S002, PLAN DATA-01 T4).
- `VALIDATE CONSTRAINT` for each new composite FK under concurrent writer load, gated the same way
  (S002, PLAN DATA-01 T5).
- Seeded cross-tenant insert negative tests under both `app_rt` and `app_platform` roles (S002,
  PLAN DATA-01 T7).
- Removing redundant single-column FKs, explicitly optional per PLAN's own acceptance row (S002,
  PLAN DATA-01 T8 — "don't block P0 closure on it").

## Out of scope

- **The online-migration protocol's own tooling** (manifest schema, lock budget, expand/backfill/
  validate/canary/switch/contract mechanics) — W02-E01's scope. This epic is the protocol's first
  real consumer for its riskiest steps (T4/T5), not its builder.
- **wowsociety's own `policy_override.rule_version_id` composite-FK migration** — PLAN's own
  wowsociety-impact note states this is "a genuine independent instance of the DATA-01 pattern" that
  requires wowapi's T1 (`UNIQUE (tenant_id, id)` on `rule_versions`) first, then wowsociety follows
  this epic's protocol once it exists — tracked as `PROD-01` in `requirement-inventory.md` §D,
  product-level, excluded from this epic's framework-side closure per mandate §2.3.
- **Remediation of any cross-tenant mismatch the T3 audit finds** — if the mismatch audit discovers
  real cross-tenant data, the remediation decision (delete, reassign tenant, escalate as a security
  incident) is outside this epic's own authority; see RISK-W02-002 and `risks.md`.

## Source requirements

DATA-01. Cross-referenced constraint/closure spec: MATRIX CS-18 ("Tenant FK integrity"), which
cites DATA-01 and DATA-09 together ("DATA-09 (T1–T5 precede the risky steps)").

## Architectural context

This epic sits directly on top of the framework's existing row-level-security architecture: every
affected child table already has RLS enforcing that the row's own `tenant_id` matches the session's
tenant, but RLS by construction only inspects the row being read or written — it has no mechanism to
prove that a foreign-key-referenced parent row belongs to the same tenant as the child claims. A
composite FK on `(tenant_id, id)`, pointing at a parent's `UNIQUE (tenant_id, id)`, converts that gap
from "trusted by convention" to "enforced by the database's own referential-integrity machinery" —
this is precisely why PLAN frames the fix as "encode tenant equality in foreign keys" rather than as
an RLS policy change. Two of the eight affected edges are of genuine architectural interest beyond a
routine FK add: `resources → organizations` sits at the boundary the `kernel/resource` package
(itself the subject of W02-E04's DATA-06 work) writes to, and `document_versions`/
`document_access_grants`/`attachments → documents/document_versions` sit in a three-table chain
where a composite FK on each link is required to make the full chain tenant-safe, not merely the
first hop.

This epic's two stories are grouped by risk tier, not by task count: S001 covers the low-risk,
purely-additive, CI-durable half of the work (parent unique indexes, the catalog scanner, the CI
gate) that can and should ship first and fast, per PLAN's own sequencing note on T6 ("do first if
sequencing allows"); S002 covers the higher-risk half (the mismatch audit, the actual composite-FK
add and validation, the negative tests, optional cleanup) that depends on both S001's own outputs
and — for T4/T5 specifically — this wave's W02-E01 protocol being accepted. This grouping is fixed
by `impl/analysis/wave-allocation-detail.md`'s canonical allocation: "S001
parent-indexes-scanner-gate (T1, T2, T6); S002 audit-fk-validate-negatives (T3, T4, T5, T7, T8 — T4/
T5 gated on E01 S001/S002 acceptance)."

## Included stories

- **W02-E02-S001 — parent-indexes-scanner-gate** (PLAN DATA-01 T1, T2, T6): `UNIQUE (tenant_id, id)`
  on every referenced parent; the tenant-FK catalog scanner; the permanent CI gate.
- **W02-E02-S002 — audit-fk-validate-negatives** (PLAN DATA-01 T3, T4, T5, T7, T8): the mismatch
  audit; the composite `NOT VALID` FK add and `VALIDATE CONSTRAINT` (both gated on W02-E01
  acceptance); seeded cross-tenant negative tests under both roles; optional redundant single-column
  FK cleanup.

## Dependencies

Depends on W02-E01 (this wave's online-migration protocol) — specifically, S002's T4/T5 (the
riskiest steps: adding the composite FK `NOT VALID` and validating it) must not start before
W02-E01-S001 and W02-E01-S002 reach `accepted`, per `impl/analysis/wave-allocation-detail.md`'s
explicit cross-wave sequencing note: "DATA-01 T4/T5 must not start before W02-E01 S001+S002
acceptance." This is a genuine, hard, in-wave epic dependency — not merely a sequencing preference —
because T4/T5 are precisely the "risky steps" PLAN's own PF-DATA cross-cutting note (6) says DATA-09
exists to make safe: "DATA-09 is new infrastructure that DATA-01 and DATA-08 W6-T1 both need before
their riskiest steps ship safely." See `dependencies.md` for the full statement and
`stories/story-002-audit-fk-validate-negatives/story.md`'s own `depends_on` front matter.

## Risks

RISK-W02-002 (the mismatch audit, PLAN DATA-01 T3, may find real cross-tenant data requiring a
remediation decision this epic's own scope cannot make unilaterally) originates at wave scope and
lands entirely within this epic's S002. See `risks.md` for the epic-scoped elaboration.

## Required decisions

None. DATA-01 has no D-0N architecture-decision dependency in the source (confirmed against
`requirement-inventory.md` §B and REVIEW §F/§U — no D-0N row cites DATA-01). This epic's stories
accordingly carry no `decisions/` directory.

## Epic acceptance criteria

- **AC-W02-E02-01**: `UNIQUE (tenant_id, id)` exists on every referenced parent (persons/
  legal_entities/party_contacts/acting_capacities' parent `parties`; `resources`' parent
  `organizations`; document_versions/document_access_grants/attachments' parents
  `documents`/`document_versions`), built `CONCURRENTLY`. The tenant-FK catalog scanner enumerates
  exactly the 8 known FKs with zero silent gaps and is wired as a permanent CI gate that fails a new
  migration adding a single-column tenant FK, proven by a negative fixture migration.
- **AC-W02-E02-02**: The mismatch audit reports zero cross-tenant mismatches against staging/prod-
  shaped data (or a documented, resolved remediation decision per RISK-W02-002 if a mismatch was
  found) before any composite FK is validated. All 8 edges carry a `VALIDATE CONSTRAINT`-clean
  composite FK, added and validated only after W02-E01-S001+S002 acceptance, under a documented
  concurrent-writer-load test for the validation step.
- **AC-W02-E02-03**: A seeded cross-tenant insert fails under both `app_rt` and `app_platform`
  roles — confirming the platform role does not bypass the new FK constraints (PLAN T7's own risk
  note: "Confirm platform role doesn't bypass FK constraints — don't assume").
- **AC-W02-E02-04**: All stories have passed independent review per mandate §14, with S002
  specifically checked for: the T4/T5 gate on W02-E01 acceptance being genuinely honored (not
  silently started early), and the mismatch-audit outcome being honestly recorded whichever way it
  resolved.

## Closure conditions

Both stories reach `accepted` (each satisfying its own `closure.md`); AC-W02-E02-01 through
AC-W02-E02-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; RISK-W02-002's outcome (zero-mismatch confirmation, or a resolved
remediation decision) is recorded, not silently dropped; DATA-01 T8 (optional FK cleanup) is
recorded as intentionally not blocking closure if it was not completed, per its own optional status.

## Status update (2026-07-16)

`status: partially-accepted`. Independent review executed 2026-07-16 superseded the prior
uncorroborated `W02ReviewGate` citation; the "8 edges" → "9 edges" count correction applied to
`story-002-audit-fk-validate-negatives/closure.md` and to
`testkit/tenant_fk_cross_tenant_test.go`'s comment. Story S001 is `accepted`; story S002 was
rolled back to `implemented` because three of its named proof artifacts were never built
(**DEV-PROG-005**) — disposition pending via **DEC-PROG-003**. The epic re-accepts when that
decision is executed.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
