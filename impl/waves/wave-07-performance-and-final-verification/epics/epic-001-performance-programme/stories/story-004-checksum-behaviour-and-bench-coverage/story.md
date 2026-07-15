---
id: W07-E01-S004
type: story
title: Checksum behaviour and bench coverage — required checksums, bounded repair, 7-package hot-path expansion
status: accepted
wave: W07
epic: W07-E01
owner: W07-Scoping-Dispatch.W07E01S004
reviewer: W07-Scoping-Dispatch.W07E01S004ReviewR
priority: P2
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - PERF-05
  - CS-16
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W07-E01-S004-01
  - AC-W07-E01-S004-02
  - AC-W07-E01-S004-03
  - AC-W07-E01-S004-04
  - AC-W07-E01-S004-05
  - AC-W07-E01-S004-06
  - AC-W07-E01-S004-07
artifacts:
  - ART-W07-E01-S004-001
  - ART-W07-E01-S004-002
  - ART-W07-E01-S004-003
  - ART-W07-E01-S004-004
  - ART-W07-E01-S004-005
  - ART-W07-E01-S004-006
evidence:
  - EV-W07-E01-S004-001
  - EV-W07-E01-S004-002
  - EV-W07-E01-S004-003
  - EV-W07-E01-S004-004
  - EV-W07-E01-S004-005
  - EV-W07-E01-S004-006
  - EV-W07-E01-S004-007
decisions: []
risks: []
---

# W07-E01-S004 — Checksum behaviour and bench coverage — required checksums, bounded repair, 7-package hot-path expansion

## Story ID

W07-E01-S004

## Title

Checksum behaviour and bench coverage — required checksums, bounded repair, 7-package hot-path expansion

## Objective

Require framework uploads to always persist canonical checksum metadata, auditing every upload call path
for universality (T1); move the full-hash fallback to an explicit, size/time-bounded import/repair path
(T2); add dedicated metrics for fallback invocations (T3); build a resumable async backfill for legacy
objects (T4); publish before/after evidence against `perf/reference-v1.json` (T5); and, per MATRIX
CS-16's own 7-package hot-path benchmark-coverage expansion, add benchmarks and bench-budget entries for
`kernel/database` (tenant-tx open/commit), `jobs` (claim/finalize loop), `outbox` (relay dispatch
batch), `workflow`, `auth` (token verify), `mfa` (TOTP derive), and `httpclient` (guarded dial) —
expanding `BENCH_PKGS` beyond its current 8/55 coverage.

## Value to the framework

PLAN's own PERF-05 evidence: "`S3.Stat` returns immediately if checksum-signed metadata is present;
otherwise full-downloads and streams through `sha256`. A checksum-on-upload path already exists (per the
code's own comment) — PERF-05 is about making it *required* for framework uploads, not inventing it."
This story converts an already-good mechanism from optional to mandatory, closing the gap where a
missing checksum silently degrades every `Stat` call into a full download. MATRIX CS-16's own evidence
for the bench-coverage half of this story is the more structurally significant finding: "**exactly 8 of
55 non-cmd packages have any `Benchmark*`**... leaving hot-path candidates **kernel/database, jobs,
outbox, workflow, auth, mfa, httpclient** and all adapters unbenched." MATRIX CS-16's own consequence
framing: "a perf regression in the transaction manager, job claim loop, or outbox relay is invisible to
the only performance gate the repo has" — this story's own second half closes exactly that blind spot
for the 7 packages a consumer actually spends the most time in.

## Problem statement

PLAN's own PERF-05 task table: "T1. Require framework uploads to always persist canonical checksum
metadata; audit every upload call path for universality | — | 'Framework uploads always persist and
verify canonical checksum metadata'; normal `Stat` performs no body download | Integration: upload via
framework path, `Stat`, assert no `GetObject` call | `PERF-05/upload-checksum-required.json` | Medium —
enumerate every current upload call site." MATRIX CS-16's own fix specification for the bench-coverage
half: "add benchmarks + budget entries for the 7 named hot-path packages (claim/finalize loop, tenant-tx
open/commit, relay dispatch batch, token verify, TOTP derive, guarded dial)."

## Source requirements

PERF-05 (T1–T5); CS-16 (the 7-package bench-coverage expansion, PERF-06's own bench-coverage item per
`requirement-inventory.md`'s FBL-07 row cross-reference).

## Current-state assessment

Per PLAN's own evidence and MATRIX CS-16 (to be re-confirmed at this story's own execution commit):
`S3.Stat` returns immediately if checksum-signed metadata is present; otherwise it full-downloads and
streams through `sha256`. A checksum-on-upload path already exists but is not required/universally
enforced. `bench-budgets.txt` has 43 budgeted entries; exactly 8 of 55 non-cmd packages
(`kernel/{audit,authz,config,filtering,httpx,pagination,policy,sequence}`) have any `Benchmark*`,
matching `BENCH_PKGS` at `Makefile:206-214`. `kernel/database`, `jobs`, `outbox`, `workflow`, `auth`,
`mfa`, and `httpclient` have zero benchmarks.

## Desired state

Every framework upload call path is enumerated and confirmed to persist and verify canonical checksum
metadata; a normal `Stat` call performs no body download, proven by an integration test asserting no
`GetObject` call occurs. The full-hash fallback is reachable only from a labeled repair invocation
(e.g. a distinct `RepairChecksum` method or `Stat` variant), never from ambient `Stat` calls, proven by
a test where a legacy object triggers the fallback only via the labeled path. Dedicated metrics
(counter/histogram) exist for fallback hits, bytes, and duration. A resumable async backfill for legacy
objects exists, proven by an interrupt/resume test with no duplicate work and eventual completion.
Before/after evidence is published against `perf/reference-v1.json`. `BENCH_PKGS` is expanded to
include `kernel/database`, `jobs`, `outbox`, `workflow`, `auth`, `mfa`, and `httpclient`, each with a
benchmark targeting the specific hot path MATRIX CS-16 names (tenant-tx open/commit; claim/finalize
loop; relay dispatch batch; a workflow-specific hot path; token verify; TOTP derive; guarded dial) and a
passing bench-budget entry.

## Scope

- **T1** — Require framework uploads to always persist canonical checksum metadata; audit every upload
  call path for universality.
- **T2** — Move the full-hash fallback to an explicit, size/time-bounded import/repair path.
- **T3** — Dedicated metrics for fallback invocations (counter/histogram for hits, bytes, duration).
- **T4** — Resumable async backfill for legacy objects.
- **T5** — Publish before/after evidence against `perf/reference-v1.json` (the "no body download"
  behavioral proof is independently testable now per PLAN's own framing; only the quantified latency
  claim needs the reference environment).
- **CS-16 bench-coverage expansion** — benchmarks + bench-budget entries for `kernel/database` (tenant-tx
  open/commit), `jobs` (claim/finalize loop), `outbox` (relay dispatch batch), `workflow`, `auth` (token
  verify), `mfa` (TOTP derive), `httpclient` (guarded dial).

## Out of scope

- **PERF-05's own object-checksum-on-upload mechanism design** — already exists per the code's own
  comment; this story makes it required and universal, it does not invent the mechanism from scratch.
- **Extending `BENCH_PKGS` to every one of the 47 currently-unbenched packages** — MATRIX CS-16's own
  fix names exactly 7, not all 55; this story's own scope is bounded to those 7, not a general
  benchmark-everything initiative.
- **Adapter-level benchmarks** — MATRIX CS-16's own evidence notes "all adapters unbenched" as part of
  the broader 47-package gap, but its own fix specification names only the 7 kernel-level packages; adapter
  benchmarking is not this story's own scope unless a future finding names it explicitly.

## Assumptions

- The exact `storage.ObjectInfo` port API-surface decision for T2 (a new `Stat` variant vs. a separate
  `RepairChecksum` method) is not specified by any source document beyond PLAN's own framing — this
  affects other adapters implementing that port, per PLAN T2's own risk note; this story's own T2 design
  work determines the exact API shape.
- The exact per-package benchmark target within each of the 7 CS-16-named packages (beyond the specific
  hot paths MATRIX CS-16 names: tenant-tx open/commit, claim/finalize loop, relay dispatch batch, token
  verify, TOTP derive, guarded dial) — the `workflow` package's own specific target is not named as
  precisely as the other 6 in MATRIX CS-16's own text (it names "workflow" without a specific sub-path
  the way it does for the others) — this story's own implementation determines the exact workflow
  benchmark target, informed by that package's own most request-relevant hot path.

## Dependencies

None within W07-E01 for T1-T4 (checksum behavior is disjoint from S001-S003's own scope). T5 consumes
W07-E01-S001's `perf/reference-v1.json` for its own quantified-latency portion (though the "no body
download" behavioral proof is independently testable now, per PLAN's own framing). The CS-16 bench-
coverage expansion has no dependency on S001-S003 either — it is a package-level micro-benchmark
addition, distinct from S001's own complete-request benchmarks.

## Affected packages or components

`adapters/storage/s3` (the checksum-required enforcement, the bounded repair path); the `storage.
ObjectInfo` port (possible API-surface change for T2); new benchmark files in `kernel/database`,
`kernel/jobs` (referred to as "jobs" in MATRIX CS-16), `kernel/outbox`, `kernel/workflow`, `kernel/auth`,
`kernel/mfa`, `kernel/httpclient`; `bench-budgets.txt` (7 new entries); `Makefile` (`BENCH_PKGS`
extended).

## Compatibility considerations

T2's own possible `storage.ObjectInfo` port API-surface change is explicitly flagged by PLAN as
"affecting the `storage.ObjectInfo` port other adapters implement" — this story's own implementation
must confirm any other adapter (beyond `adapters/storage/s3`) implementing that port still compiles and
behaves correctly after the change.

## Security considerations

Required, universal checksum enforcement (T1) is itself a data-integrity control — ensuring every
framework upload's integrity is verifiable, not merely "verifiable if the upload happened to include
checksum metadata."

## Performance considerations

This story IS the performance work (T1-T5's own checksum-cost-reduction and CS-16's own bench-coverage
expansion); no separate performance concern beyond what these tasks already address.

## Observability considerations

T3's own dedicated fallback-invocation metrics are this story's own primary observability addition for
the PERF-05 half; the CS-16 half's own benchmark additions are themselves an observability-adjacent
capability (making regressions visible where they were previously invisible).

## Migration considerations

T4's own resumable backfill needs an inventory mechanism for "legacy objects lacking checksum metadata,"
which PLAN's own T4 risk note flags as "doesn't obviously exist yet" — this story's own T4 implementation
must build or confirm this inventory mechanism as part of its own scope, not assume it already exists.

## Documentation requirements

Document the checksum-required enforcement's own call-site audit results (T1); document the labeled
repair path's own invocation convention (T2); document each of the 7 new benchmark packages' own
specific hot-path target, so a future maintainer understands what each benchmark is actually measuring.

## Acceptance criteria

- **AC-W07-E01-S004-01**: Every framework upload call path persists and verifies canonical checksum metadata; a
  normal `Stat` call performs no body download, proven by an integration test asserting no `GetObject`
  call occurs.
- **AC-W07-E01-S004-02**: The full-hash fallback is reachable only from a labeled repair invocation, never
  ambient `Stat`, proven by a test where a legacy object triggers the fallback only via the labeled
  path.
- **AC-W07-E01-S004-03**: Dedicated metrics (counter/histogram) exist for fallback hits, bytes, and duration.
- **AC-W07-E01-S004-04**: The resumable async backfill survives an interrupt-and-resume cycle with no duplicate
  work and reaches eventual completion.
- **AC-W07-E01-S004-05**: Before/after evidence is published against `perf/reference-v1.json`; the "no body
  download" behavioral proof (AC-01) stands independently of the reference environment.
- **AC-W07-E01-S004-06**: `BENCH_PKGS` includes `kernel/database`, `jobs`, `outbox`, `workflow`, `auth`, `mfa`,
  and `httpclient`, each with a benchmark targeting its own named hot path (tenant-tx open/commit;
  claim/finalize loop; relay dispatch batch; a workflow-specific target; token verify; TOTP derive;
  guarded dial).
- **AC-W07-E01-S004-07**: `make bench-budget` exits 0 with all 7 new package entries present and passing.

## Required artifacts

- The checksum-required enforcement + call-site audit (T1).
- The bounded repair path (T2).
- Fallback-invocation metrics (T3).
- The resumable backfill mechanism (T4).
- The published comparison report (T5).
- 7 new benchmark files with bench-budget entries (CS-16 expansion).
See `artifacts/index.md`.

## Required evidence

- Integration test output confirming no `GetObject` call on normal `Stat` (T1).
- Labeled-repair-path test output (T2).
- Metric-emission test output (T3).
- Interrupt/resume backfill test output (T4).
- The published comparison report (T5).
- `make bench-budget` passing output including all 7 new entries (CS-16 expansion).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all seven acceptance criteria numbered and measurable, no dependency, owner/
reviewer assignment pending, T2's own port-API-surface decision and T4's own inventory-mechanism gap
recorded as unresolved questions rather than silently assumed solved.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all seven acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming each of the 7 new benchmarks genuinely targets the
specific hot path MATRIX CS-16 names (not a generic, loosely-related benchmark that happens to live in
the right package).

## Risks

None recorded at this story's own scope beyond the general port-API-surface-change risk (T2) and the
inventory-mechanism gap (T4), both mitigated by this story's own explicit scoping of those decisions as
implementation-time work, not pre-decided facts.

## Residual-risk expectations

Once all seven acceptance criteria are verified, residual risk is expected to be low — this is a
well-bounded, source-derived closure story combining two distinct but individually well-specified
finding sets (PERF-05, CS-16's bench-coverage expansion).

## Plan

See `plan.md`.
