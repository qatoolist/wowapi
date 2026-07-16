---
id: TRACK-STATUS-REGISTER
type: register
title: Status register — derived roll-up of canonical front-matter statuses
status: active
created_at: 2026-07-12
updated_at: 2026-07-16
derived: true
generated_by: miscellaneous/regen_status_register.py
---

# Status register

**DERIVED VIEW — generated 2026-07-16 by `miscellaneous/regen_status_register.py`.**
Canonical status lives in each wave/epic/story's own front matter (mandate §6). Do not
hand-edit this file; regenerate it after any canonical status change.

## wave-00-baseline-and-verification — wave status: `accepted`

| Item | Level | Title | Status |
|---|---|---|---|
| W00-E01 | epic | Executed-slice verification | planned |
| W00-E01-S001 | story | Verify workflow and boot composition slices at current HEAD | accepted |
| W00-E01-S002 | story | Verify performance and benchmark-budget-gate slices at current HEAD | accepted |
| W00-E01-S003 | story | Verify data-durability and CI-integration slices at current HEAD | accepted |
| W00-E02 | epic | Baseline capture | planned |
| W00-E02-S001 | story | Quality baselines | accepted |
| W00-E02-S002 | story | Dependency and toolchain inventory | accepted |
| W00-E02-S003 | story | ADR-ification of D-01 through D-09 | accepted |

## wave-01-zero-dependency-hardening — wave status: `accepted`

| Item | Level | Title | Status |
|---|---|---|---|
| W01-E01 | epic | Static-analysis utilisation | planned |
| W01-E01-S001 | story | Enable the zero-cost leak-detection linter set | accepted |
| W01-E01-S002 | story | Enable and triage the judged linter set | accepted |
| W01-E01-S003 | story | Close supply-chain and pre-push hook hygiene gaps | accepted |
| W01-E02 | epic | Observability correlation | planned |
| W01-E02-S001 | story | Trace/log correlation | accepted |
| W01-E02-S002 | story | Pgx query tracer | accepted |
| W01-E03 | epic | HTTP hardening | planned |
| W01-E03-S001 | story | Server timeouts and body bounds | accepted |
| W01-E03-S002 | story | Central validation enforcement | accepted |
| W01-E04 | epic | Generator, documentation, and test-diagnosis fixes | planned |
| W01-E04-S001 | story | Generator correctness — source-built CLI path validity and boot-safe CRUD generation | accepted |
| W01-E04-S002 | story | Documentation reconciliation — plan traceability fix, DX-05 residual, wowsociety upstream register | accepted |
| W01-E04-S003 | story | E2E flake diagnosis — reproduction-first investigation of the intermittent internal/e2e full-suite failure | accepted |

## wave-02-data-safety-and-migration-tooling — wave status: `accepted`

| Item | Level | Title | Status |
|---|---|---|---|
| W02-E01 | epic | Online migration protocol | accepted |
| W02-E01-S001 | story | Migration manifest schema and online-DDL lock budget | accepted |
| W02-E01-S002 | story | Expand-phase tooling, resumable backfill harness, and validation-phase tooling | accepted |
| W02-E01-S003 | story | Canary, switch, and contract-phase tooling with the full CI drill pipeline | accepted |
| W02-E02 | epic | Tenant foreign-key integrity | accepted |
| W02-E02-S001 | story | Parent tenant-scoped unique indexes, FK catalog scanner, and CI gate | accepted |
| W02-E02-S002 | story | Cross-tenant mismatch audit, composite FK validation, and negative tests | accepted |
| W02-E03 | epic | Version allocation and GC | accepted |
| W02-E03-S001 | story | Version-allocation races and upload-blob GC | accepted |
| W02-E04 | epic | Aggregate write contract | accepted |
| W02-E04-S001 | story | Typed aggregate write contract with mandatory mirror, audit, and outbox | accepted |
| W02-E05 | epic | Production seed-sync | accepted |
| W02-E05-S001 | story | Production catalog seed-sync path | accepted |

## wave-03-identity-and-session-security — wave status: `in-progress`

| Item | Level | Title | Status |
|---|---|---|---|
| W03-E01 | epic | Server-side session state | in-progress |
| W03-E01-S001 | story | Grant schema and unconditional membership enforcement | accepted |
| W03-E01-S002 | story | Capacity selection and privileged-session resolver | accepted |
| W03-E01-S003 | story | Assurance freshness and credential-scheme distinction | verified |
| W03-E01-S004 | story | Cross-repo cutover plan for the wowsociety impersonation-flow breaking change | implemented |
| W03-E02 | epic | Outbound-security governance | accepted |
| W03-E02-S001 | story | Outbound-security escape-hatch governance | accepted |
| W03-E03 | epic | Webhook authenticated replay | accepted |
| W03-E03-S001 | story | Bind webhook replay and dedup to provider-authenticated data | accepted |
| W03-E04 | epic | Relationship semantics | accepted |
| W03-E04-S001 | story | Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation governance | accepted |
| W03-E05 | epic | Workflow privileged completion | accepted |
| W03-E05-S001 | story | Workflow privileged completion — ratification and durable override audit | accepted |

## wave-04-jobs-and-durable-delivery — wave status: `in-progress`

| Item | Level | Title | Status |
|---|---|---|---|
| W04-E01 | epic | Lease-fencing primitive and jobs | accepted |
| W04-E01-S001 | story | Shared lease/fencing primitive | accepted |
| W04-E01-S002 | story | Jobs lease columns, fenced finalize, and fenced reclaim | accepted |
| W04-E01-S003 | story | Worker idempotency contract and the shared duplicate-worker chaos harness | accepted |
| W04-E02 | epic | Remote I/O outside transactions | in-progress |
| W04-E02-S001 | story | Notify and webhook three-stage remote-I/O protocol | accepted |
| W04-E02-S002 | story | Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos test | planned |
| W04-E02-S003 | story | Adopt cenkalti/backoff/v5 for duplicated retry logic | accepted |
| W04-E03 | epic | Bulk multi-worker safety | accepted |
| W04-E03-S001 | story | Bulk multi-worker stopgap — correct false safety claim, enforce single-processor | accepted |
| W04-E03-S002 | story | Leased claims, finalize fencing, lifecycle controls, and the named multi-worker chaos test | accepted |
| W04-E04 | epic | Compliance and readiness | accepted |
| W04-E04-S001 | story | Audit hash-chain widening with hash_version discriminator | accepted |
| W04-E04-S002 | story | External anchoring, DSR export artifact, central legal-hold, and explicit per-class status | accepted |
| W04-E04-S003 | story | Readiness and configuration diagnostics truthfulness | accepted |

## wave-05-application-model-and-layering — wave status: `planned`

| Item | Level | Title | Status |
|---|---|---|---|
| W05-E01 | epic | Application model | planned |
| W05-E01-S001 | story | ApplicationModel lifecycle skeleton and Registrar capability type | planned |
| W05-E01-S002 | story | Owner-bound registry wrappers across all declaration classes | planned |
| W05-E01-S003 | story | Snapshot immutability, post-seal rejection, model hash, and race safety | planned |
| W05-E01-S004 | story | Legacy module/context compatibility adapter | planned |
| W05-E02 | epic | Typed ports | planned |
| W05-E02-S001 | story | Typed port-key API and registrar-forge safety proof | planned |
| W05-E02-S002 | story | Zero-reflection provider graph, boot-time validation, and profile projection | planned |
| W05-E02-S003 | story | Lifecycle manifest retirement and legacy port adapter | planned |
| W05-E03 | epic | Authoritative declarations | planned |
| W05-E03-S001 | story | Manifest schema and derived-projection tooling | planned |
| W05-E03-S002 | story | Boot-time strictness and the shared no-op-adapter waiver mechanism | planned |
| W05-E04 | epic | Wiring and cache hygiene | planned |
| W05-E04-S001 | story | Constructor-boundary lint and kernel.go audit | ready-for-review |
| W05-E04-S002 | story | Bounded, epoch-invalidated authorization cache | planned |
| W05-E05 | epic | Kernel re-home | planned |
| W05-E05-S001 | story | Foundation tree, package moves, and mfa forwarding shim | planned |
| W05-E05-S002 | story | Kernel package-count and wowsociety identity-suite verification | planned |

## wave-06-contracts-compatibility-release — wave status: `in-progress`

| Item | Level | Title | Status |
|---|---|---|---|
| W06-E01 | epic | Consumer and DSL | in-progress |
| W06-E01-S001 | story | Module DSL design — state-of-the-art DSL design record (target, not implemented) | verified |
| W06-E01-S002 | story | Golden consumer matrix — framework-repo-owned CLI/generator proof fixture | accepted |
| W06-E02 | epic | API contract gates | in-progress |
| W06-E02-S001 | story | OpenAPI merge complete-or-loud — full-field merge, validation, semantic diff | verified |
| W06-E02-S002 | story | Compatibility gates buildable now — REL-03a (Go API diff, compile matrix, config compat, migration drill, arch smoke, SBOM verify) | accepted |
| W06-E02-S003 | story | Compatibility gates unblocked — REL-03b (OpenAPI diff, event/schema compat, generated-consumer upgrade) | blocked |
| W06-E03 | epic | Release gating | in-progress |
| W06-E03-S001 | story | Exact-commit release pipeline — REL-01 T1-T8 buildable-now set | verified |
| W06-E03-S002 | story | Protection activation — branch/tag/environment protection (human-gated, DEC-Q10) | blocked |
| W06-E03-S003 | story | Blocking security scans — Trivy flip, waiver schema, visibility-guard regression check, private fallback | verified |
| W06-E04 | epic | Documentation gates | in-progress |
| W06-E04-S001 | story | Doc-example compile gate — CI-enforced normative Go example compilation | accepted |
| W06-E04-S002 | story | Generated docs and labels — model-export byte-match and future-state labeling lint | accepted |

## wave-07-performance-and-final-verification — wave status: `in-progress`

| Item | Level | Title | Status |
|---|---|---|---|
| W07-E01 | epic | Performance programme | accepted |
| W07-E01-S001 | story | Request benchmarks against real PostgreSQL — reference environment + DB-backed benchmarks | accepted |
| W07-E01-S002 | story | Rules resolution collapsed to bounded SQL — set-based query, index verification, parity proof | accepted |
| W07-E01-S003 | story | Sweeper and worker materialization — bounded batches, leased outbox, N+1 removal | accepted |
| W07-E01-S004 | story | Checksum behaviour and bench coverage — required checksums, bounded repair, 7-package hot-path expansion | accepted |
| W07-E02 | epic | Verification hardening | blocked |
| W07-E02-S001 | story | Security verification profile — version-pinned control map + external assessment | blocked |
| W07-E02-S002 | story | Coverage truthfulness completion — fail-not-skip E2E, skip manifest, race schedule, real fuzz | accepted |
| W07-E03 | epic | Product alignment verification | blocked |
| W07-E03-S001 | story | wowsociety readiness check — framework-side PROD-01..05 coordination-artifact verification | blocked |
| W07-E04 | epic | Programme closure | planned |
| W07-E04-S001 | story | Final verification gate — programme-wide REVIEW §30 re-run, traceability completeness, disposition audit | planned |
| W07-E04-S002 | story | Closure and claim decision — programme closure report + production-readiness claim-upgrade decision package | planned |

## Story status totals

- `accepted`: 49
- `blocked`: 4
- `implemented`: 1
- `planned`: 15
- `ready-for-review`: 1
- `verified`: 5
