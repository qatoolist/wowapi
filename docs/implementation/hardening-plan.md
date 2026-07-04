# Hardening Implementation Plan

> **STATUS (in progress — exit gate NOT met):** the H-phases below built and unit/DB-tested the kernel
> primitives for H1–H5 and the P1 items S2/S3/R1/R5/O1 behind the `make ci` + `make ci-container` gate.
> **An independent verification ([VERIFICATION-wowapi-hardening.md](../../VERIFICATION-wowapi-hardening.md))
> found the roadmap acceptance criteria and exit gate are NOT satisfied**: the dominant gap is
> *built-but-not-wired* (metrics port has zero emission sites; rate limiter, authz cache, signed cursors,
> OTel adapter, and several evidence primitives are not wired by default), plus no hosted CI, R2 load
> characterization missing, and expired idempotency keys re-executing. Closure is tracked as corrective
> actions **CA-1…CA-15** in that report (CA-1…CA-7 are P0). Do **not** treat this plan as "complete" until
> those are closed and re-verified.
>
> Decisions D-0061…**D-0077**; evidence under `evidence/hardening-{H1,H2,H3,H4,H5,P1}/`; migrations to
> **00022**. The roadmap's inaccurate "current state" claims (S4, O3, R8) were verified as NON-gaps; **R2
> is reopened** (declassified here to doc-only, but its acceptance was a load characterization that does
> not exist — see CA-4). Documented follow-ups now escalated to corrective actions: OTel adapter binding +
> cross-process trace propagation (CA-2/CA-9), `module.Context` accessors for the new primitives (CA-5/
> CA-11), and a `wowapi apikey` CLI (CA-3).

Companion to [ROADMAP-wowapi.md](../../ROADMAP-wowapi.md). Derived from a three-track code audit
(security S1–S8, reliability R1–R8, operational O1–O5, evidence E1–E6) run 2026-07-04 against the
v0.1.0 tree. Each roadmap "current state" claim was verified against the actual code; the table below
records the **verified** verdict, which sometimes differs from the roadmap's assumption.

Every hardening phase (`H*`) ends the same way the Goal-2 phases did: failing test first (TDD),
implementation, `make ci` + `make ci-container` green (DB tests forced via `WOWAPI_REQUIRE_DB=1`),
an evidence bundle under `docs/implementation/evidence/hardening-HN/`, a `decisions.md` entry for any
blueprint deviation, and a coherent commit.

## Verified verdicts

| Item | Roadmap claim | Verified verdict | Real gap? | Pri | Phase |
|---|---|---|---|---|---|
| S1 machine auth | only OIDC JWTs | non-human actors exist by hardcoded name; no issuable credential | **yes** (med) | P0 | H3 |
| S2 rate limiting | proxy-delegated | accurate; no in-process limiter | yes (med) | P1 | backlog |
| S3 step-up/MFA | token only | accurate; no challenge hook | yes (med) | P1 | backlog |
| S4 encrypt creds | stored as-is | **inaccurate** — ref-only + compiler-enforced redaction | no | — | doc-only |
| S5 idempotency expiry | kept forever | accurate; `expires_at` exists, no sweep | yes (small) | P1 | H1 |
| S6 audit tamper-evidence | append-only rows | **no audit_logs table at all**; only nil-safe log sink | yes (large) | P0 | H4 |
| S7 reference deploy / headers | proxy's job | blueprint specifies `SecureHeaders→CORS`; **never implemented** | yes (med) | P0 | H1 |
| S8 adversarial | good tests, no fuzz | accurate; zero fuzz / property tests | yes (small) | P0 | H1 |
| R1 authz cache + RO routing | every Evaluate hits DB | accurate; no cache; Evaluate uses caller TenantDB | yes (med) | P1 | backlog |
| R2 advisory-lock contention | serializes per aggregate | accurate; lock correct | no | — | doc-only |
| R3 SLA sweeper | hardcoded, single-runner | **not registered as a job at all**; no leader election | yes (large) | P0 | H2 |
| R4 DLQ operability | works, no tooling | accurate; no inspect/replay/discard | yes (med) | P0 | H2 |
| R5 notification evidence | fire-and-forget | **inaccurate** — status tracked; lacks receipts API + prefs | partial (small) | P1 | backlog |
| R6 retention legal-hold race | checked once | accurate; race window real | yes (med) | P0 | H1 |
| R7 cursor versioning | silent wrong pages | accurate; no sort-spec version in cursor | yes (small) | P1 | H1 |
| R8 webhook breaker granularity | per-endpoint | accurate; correct as intended | no | — | doc-only |
| O1 distributed tracing | request-id only | accurate; no OTel | yes (med) | P1 | backlog |
| O2 migration harness | no tests | **partial** — migration tests exist; no fwd/down CI drill, no expand/contract | yes (small) | P0 | H2 |
| O3 upgrade discipline | fair game | **inaccurate** — `RunModuleContract` tripwire exists | no | — | doc-only |
| O4 config-drift alerting | nothing consumes fp | accurate; fingerprint exposed, no alert convention | yes (small) | P1 | H1 |
| O5 backup/restore drill | nothing | accurate; no runbook/script | yes (large, docs) | P0 | H2 |
| E1 field-level audit | row-level only | accurate; no durable audit table/query API | yes (v.large) | P0 | H4 |
| E2 retention engine | document-only | accurate; no generalized disposition / DSR | yes (large) | P0 | H5 |
| E3 sequence allocator | nothing | accurate; no gap-free numbered series | yes (large) | P0 | H5 |
| E4 snapshot/artifact pipeline | nothing | accurate; no immutable artifact pipeline | yes (v.large) | P0 | H5 |
| E5 scheduler | on-demand only | accurate; `run_at` delay but no cron/recurring | yes (large) | P0 | H2 |
| E6 bulk-operation framework | per-job only | accurate; no chunk/progress/resume | yes (large) | P0 | H5 |

## Phase sequence

Ordered cheapest-highest-leverage first so each phase ships an independently valuable, fully
QA-gated hardening increment.

- **H1 — Edge & pagination hardening** (self-contained, no new heavy schema):
  S7 `SecureHeaders`+`CORS` middleware (+ generated-api wiring + reference proxy config + smoke),
  S8 fuzz tests (filter DSL parser, cursor decode), R7 cursor sort-spec versioning,
  S5 idempotency expiry sweep, R6 retention legal-hold race fix, O4 config-drift alert convention.
- **H2 — Async operability** (clusters on jobs/outbox): E5 cron scheduler (recurring job registry),
  R3 SLA sweeper registered on the scheduler + advisory-lock leader guard, R4 DLQ inspect/replay/discard
  CLI, O2 migration fwd/down CI drill + expand/contract doc, O5 backup/restore runbook + drill script.
- **H3 — Machine identity**: S1 issuable, scoped, rotatable, revocable, audited service principals /
  API keys; `Authenticator` for API keys composes with the H1/OIDC gate.
- **H4 — Durable audit + tamper-evidence**: E1 field-level `audit_logs` (entity/field/before/after/
  actor/capacity/impersonator/request-id/tx-id; append-only; query API; redaction hooks) written inside
  TenantDB; S6 per-tenant-per-period hash-chaining + exportable anchors + verification tool.
- **H5 — Evidence primitives** (largest; sub-sequenced): E3 gap-free per-tenant sequence allocator with
  audited voids; E2 generalized retention/disposition + legal hold + DSR export/erasure; E6 bulk-op
  framework (chunk/progress/partial-failure ledger/resume); E4 snapshot/artifact pipeline.

## P1 backlog (subsequently implemented at the kernel level — wiring/finishers tracked as CAs)

These were built after this plan's first draft (commits for S2 rate limiting, S3 step-up/MFA, R1 authz
caching, R5 receipts + channel prefs, O1 OTel seam), so the "not in this pass" wording above is **stale**.
However, independent verification found each landed as a **kernel primitive that is not wired by default**
or with acceptance residuals — closure is tracked as: S2→CA-1/CA-2, S3→CA-13, R1→CA-2/CA-10, R5→CA-15,
O1→CA-2/CA-9, O2 expand/contract helper→CA-12.

## Doc-only (verified NOT gaps — recorded so the roadmap's inaccurate rows don't reopen)

S4, O3, R2, R5 (core), R8. Each gets a one-line note in the relevant evidence bundle explaining why the
roadmap's "current state" was wrong, with the file:line proving the capability already exists.
