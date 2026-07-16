---
id: W04-E04
type: epic
title: Compliance and readiness
status: accepted
wave: W04
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-08
  - DX-07
depends_on:
  - W02-E01
stories:
  - W04-E04-S001
  - W04-E04-S002
  - W04-E04-S003
decisions:
  - D-04
risks:
  - RISK-W04-002
  - RISK-W04-004
---

# W04-E04 — Compliance and readiness

## Epic objective

Widen the audit hash chain (`kernel/audit.chainHash`) to cover every persisted field — including the
previously-excluded canonicalized `metadata` and `tx_id` — with a `hash_version smallint`
discriminator column enacting the already-ratified D-04 decision (DATA-08 W6-T1); add external
anchor verification for the audit chain, an encrypted immutable DSR export artifact, a central
legal-hold enforcement wrapper, and explicit per-record-class DSR status reporting (DATA-08 W6-T2
through T5); and make the framework's own readiness and configuration diagnostics truthful by adding
a migration-currency check, seed/rule/model-hash reporting, and a `go env GOMOD`-based `config
doctor` discovery fix (DX-07 T1-T3, with T4 explicitly out of scope).

## Problem being solved

`requirement-inventory.md` row DATA-08 states: "Compliance evidence complete/durable | IMPL | P0/P1 |
partial | W04-E04-S001..S002 | W0 slice EXECUTED (verified ×2); W6-T1 hash widening (D-04) + T2–T5
planned." Row DX-07 states: "Truthful readiness/config diagnostics (T1–T4) | IMPL | P1 | planned |
W04-E04-S003 | T4 dep AR-04 T5 waiver mechanism." Both requirements share a confirmed-real defect
pattern the source evidence documents directly, not speculatively.

For DATA-08: `audit.go`'s `chainHash` explicitly excludes `metadata` (a documented jsonb round-trip
reformatting problem) and — unnamed by the directive but confirmed by reading the code — `tx_id`,
inserted via `pg_current_xact_id()` and never hashed. This means an attacker or bug can alter audit
`metadata`/`tx_id` on a row without breaking the tamper-evidence chain: "for compliance evidence,
[a partial guarantee] is close to none" (MATRIX CS-20). No `hash_version` column exists today, so
naively widening the hash's field coverage would make every historical row unverifiable under
new-scheme verification — this is why D-04 (already ratified) specifies a version discriminator in
the same migration.

For DX-07: the health contract's own documentation describes readiness as including "migration
currency," but the generated `cmd/api/main.go.tmpl` readiness map registers only `"db"` and
`"seeds"` — no migration-currency check exists, directly contradicting the documented contract.
`config_delegate.go`'s product-checker discovery is CWD-relative `os.Stat`, silently falling back to
framework-only validation if the product root is not found there. `CapacityMode` defaults to
`"advisory"` (never enforced); `HTTPMaxInFlight` defaults to `0` (backpressure fully disabled) — but
the fix for that gap (T4) is out of this epic's scope, per its own forward dependency on a waiver
mechanism that does not yet exist.

## Scope

- **DATA-08 W6-T1** — widen `chainHash` to cover every persisted field (canonicalized `metadata`,
  `tx_id`, all nullable fields, sequence, ID, timestamps, previous hash); add a `hash_version
  smallint NOT NULL DEFAULT 1` column in the same migration; branch verification by row version so
  historical rows verify under v1 and new rows under v2. Enacts D-04 (S001).
- **DATA-08 W6-T2** — external anchor verification for the audit chain, so tampering is detectable
  even if `head_hash` were locally compromised (S002).
- **DATA-08 W6-T3** — persist the DSR export as an encrypted, immutable artifact with a manifest,
  per-class results, checksum, expiry, access policy, and download audit, replacing
  `retention/engine.go`'s bare in-memory map return (S002).
- **DATA-08 W6-T4** — a central legal-hold enforcement wrapper every `Dispose`/`Erase` callback must
  pass through, replacing today's per-callback responsibility (S002).
- **DATA-08 W6-T5** — explicit partial/not-applicable DSR results for record classes without
  export/erase callbacks, so the result set never silently omits a registered class (S002).
- **DX-07 T1** — add a migration-currency check to the generated readiness template (S003).
- **DX-07 T2** — add seed/rule/model-hash checks to readiness (S003).
- **DX-07 T3** — `config doctor` discovers the product root via `go env GOMOD`/`--project`, not
  CWD-relative `os.Stat` (S003).

## Out of scope

- **DATA-08 Wave-0 tasks (W0-T1, W0-T2)** — already executed and verified ×2 elsewhere in the
  programme, per `requirement-inventory.md`'s own note ("W0 slice EXECUTED (verified ×2)"). This
  epic references them only as already-done current-state context in S001; it does not create
  implementation tasks for them and does not re-mark their status as this epic's own work.
- **DX-07 T4** — production-profile capacity/backpressure enforcement. T4's own dependency column
  states "T1-T3, AR-04's waiver framework," and `AR-04 T5` (the shared waiver mechanism) is
  W05-E03-S002 scope, which does not yet exist as of this epic's planning. Deferred-linked by
  requirement ID only, per RISK-W04-004 (wave-level) — not implemented here, no task created for it.
- **PROD-05** — the wowsociety-side staging audit re-verification drill before `FRAMEWORK_VERSION`
  is bumped past W6-T1's commit. This is a product-level compliance drill
  (`requirement-inventory.md` §D), excluded from this epic's framework-side closure per mandate
  §2.3. Recorded as a noted, non-blocking coordination item in S001, not implemented here.
- **PROD-03** — wowsociety's already-committed `cmd/api/main.go` readiness backport. DX-07 T1's fix
  changes the generated template only; it does not retroactively alter wowsociety's own
  already-generated, hand-edited `main.go`. Recorded as a non-blocking follow-up note in S003, not
  implemented here.
- **wowsociety's own `tools/configcheck/main.go`** — already exists and already engages product-aware
  validation correctly today, per DX-07's own wowsociety-impact note; DX-07 T3's discovery fix is a
  wowapi-side improvement only and has no wowsociety-side action item.

## Source requirements

DATA-08 (W6-T1 through W6-T5), DX-07 (T1, T2, T3; T4 explicitly excluded). D-04 (already ratified,
enacted by S001, not authored here).

## Architectural context

This epic sits at the intersection of two independent closure specs sharing one epic slot per the
wave's own grouping rationale (`wave.md`'s "Rationale": "DX-07 closes this wave because MATRIX CS-21
ties it to the same deployment-readiness closure spec as FBL-02... DX-07 T1-T3 are buildable now and
are grouped here per `wave-allocation-detail.md`'s canonical allocation"). DATA-08 W6 (S001, S002)
and DX-07 (S003) do not share a task dependency chain with each other — S003 has no dependency on
S001/S002, per `dependencies.md` (wave-level): "DX-07 (S003) has no dependency on S001/S002 — it is
an independent readiness/diagnostics concern grouped into the same epic by MATRIX CS-21's shared
closure-spec framing, not by a task dependency." Internally, DATA-08 W6 is a strict two-phase
pipeline: S001 (W6-T1, the hash-widening migration and its `hash_version` discriminator) must land
before S002 (W6-T2 through W6-T5, which build on the widened, versioned hash chain and the DSR/
legal-hold surface).

S001 (W6-T1) is confirmed by the source as "**Single highest-risk task in PF-DATA's Wave-6
scope, and directly hits wowsociety's live audit rows**" (PLAN DATA-08 W6-T1's own risk column) — a
breaking audit-hash format change against a live-production compliance-evidence chain, gated on
W02-E01's online-migration protocol existing first (per `wave.md`'s entry criteria: "this is the one
concrete predecessor capability this wave's own stories require, and only for W04-E04-S001"). This
is not a softened characterization; the epic's own risk treatment (`risks.md`) and S001's `story.md`/
`plan.md` must convey it as stated, not paraphrase it into a generic "risky migration."

## Included stories

- **W04-E04-S001 — audit-hash-widening** (DATA-08 W6-T1 only; enacts D-04): the widened `chainHash`
  covering every persisted field, the `hash_version` discriminator migration, and version-branched
  verification. Single highest-risk task in this epic; ships via W02-E01's online-migration
  protocol.
- **W04-E04-S002 — anchor-dsr-hold** (DATA-08 W6-T2, T3, T4, T5): external anchor verification;
  encrypted immutable DSR export artifact; central legal-hold enforcement wrapper; explicit
  per-class DSR status reporting.
- **W04-E04-S003 — readiness-truthfulness** (DX-07 T1, T2, T3 only; T4 explicitly deferred-linked to
  W05-E03-S002's AR-04 T5 waiver mechanism, not implemented here): migration-currency readiness
  check; seed/rule/model-hash readiness reporting; `config doctor` product-root discovery fix.

## Dependencies

Depends on **W02-E01** (the DATA-09 online-migration protocol), narrowly for **W04-E04-S001 only** —
per `wave.md`'s entry criteria and `wave-allocation-detail.md`'s W04-E04 row ("dep W02-E01
protocol"), and confirmed at wave scope by
`impl/waves/wave-02-data-safety-and-migration-tooling/dependencies.md`'s downstream table: "W04-E04-
S001 (DATA-08 W6-T1 audit hash widening) | W02-E01 (DATA-09 protocol) | ... the audit hash-chain
widening migration (a breaking format change touching wowsociety's live audit rows) is expected to
ship via DATA-09's protocol, not ad hoc." S002 and S003 have no dependency on W02-E01. See
`dependencies.md` for the full statement.

## Risks

RISK-W04-002 (DATA-08 W6-T1's confirmed highest-risk-in-wave status, breaking audit-hash format
change hitting wowsociety's live rows) and RISK-W04-004 (DX-07 T4's forward dependency on W05-E03-
S002's not-yet-built waiver mechanism) both originate at wave scope (`../../risks.md`) and land
entirely within this epic's S001 and S003 respectively. See `risks.md` for the epic-scoped
elaboration.

## Required decisions

**D-04** (audit `hash_version smallint` discriminator, version-branched verification) — already
ratified in `impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/
story-003-adr-ification/decisions/adr-004-audit-hash-version-column.md`. Enacted by S001 only. S001
carries the epic's one `decisions/` directory, referencing this ADR, not authoring a new one. No
other story in this epic enacts a D-0N decision.

## Epic acceptance criteria

- **AC-W04-E04-01**: The widened `chainHash` covers every persisted field, including canonicalized
  `metadata` and `tx_id`; mutating any declared field independently breaks verification, proven by a
  per-field tamper test. A `hash_version smallint NOT NULL DEFAULT 1` column exists; historical rows
  verify under the v1 branch; new rows verify under the v2 branch (metadata + tx_id included).
- **AC-W04-E04-02**: The audit chain is periodically anchored externally, with tamper detectable even
  if the local `head_hash` were compromised. DSR export completes only after an encrypted immutable
  artifact (manifest, per-class results, checksum, expiry, access policy, download audit) is
  successfully written. A deliberately non-compliant `Dispose`/`Erase` callback is still blocked by
  the central legal-hold wrapper. The DSR result set explicitly lists every registered record class
  with a status, never a silent omission.
- **AC-W04-E04-03**: `/readyz` fails when the applied-migration version lags the expected version.
  Readiness reports migration version, seed/rule hash, and model hash. `config doctor` discovers the
  product root via `go env GOMOD`/`--project` regardless of invocation directory and explicitly
  reports whether product validation ran.
- **AC-W04-E04-04**: All three stories have passed independent review per mandate §14. S001's review
  specifically confirms the per-field tamper test covers every declared field independently (not a
  generic tamper test) and that D-04's version-branch design was implemented exactly as ratified.
  S003's review specifically confirms DX-07 T4 was correctly scoped out (not silently attempted or
  silently dropped without a forward reference).

## Closure conditions

All three stories reach `accepted` (each satisfying its own `closure.md`); AC-W04-E04-01 through
AC-W04-E04-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; RISK-W04-002 is recorded as not-fully-resolvable within this epic's
own framework-side closure (final confidence requires the product-side PROD-05 staging drill,
tracked but not owned here); RISK-W04-004 is recorded as remaining open by design (resolves only once
W05-E03-S002 lands); PROD-05 and PROD-03 are recorded as noted, non-blocking coordination items, not
silently dropped.

## Status update (2026-07-16)

`status: accepted` (reconfirmed) — all three stories independently reviewed and accepted per
`review-gate-2026-07-16.md`, including normalizing the non-vocabulary `closed-pending-review`
token used by S001 and S002 to `accepted`, and reversing S003's prior `unsupported-by-evidence`
autopsy verdict (a time-budget limitation of that pass, not a defect).

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
