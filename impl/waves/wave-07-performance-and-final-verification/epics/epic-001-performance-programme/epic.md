---
id: W07-E01
type: epic
title: Performance programme
status: accepted
wave: W07
owner: W07-Phase-A-Execution.W07E01S001
reviewer: W05ReviewGateRerun
priority: high
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - PERF-02
  - PERF-03
  - PERF-04
  - PERF-05
  - CS-16
depends_on: []
stories:
  - W07-E01-S001
  - W07-E01-S002
  - W07-E01-S003
  - W07-E01-S004
decisions:
  - DEC-Q9
risks:
  - RISK-W07-001
---

# W07-E01 — Performance programme

## Epic objective

Execute PERF-02 (complete-request benchmarks against real PostgreSQL), PERF-03 (rules resolution
collapsed to bounded SQL), PERF-04 (sweeper/worker N+1 and unbounded-materialization removal), and
PERF-05 (explicit object-checksum behavior) as relative/container comparisons now, per REVIEW §12's own
framing, with absolute-SLO acceptance criteria explicitly conditional on DEC-Q9; and expand hot-path
benchmark coverage from 8/55 non-cmd packages to include the 7 MATRIX CS-16-named packages
(`kernel/database`, `jobs`, `outbox`, `workflow`, `auth`, `mfa`, `httpclient`).

## Problem being solved

MATRIX CS-16's own evidence: "benchbudget gate fails closed (PERF-06 verified at HEAD); `bench-
budgets.txt` = 43 budgeted entries; **exactly 8 of 55 non-cmd packages have any `Benchmark*`**
(kernel/{audit,authz,config,filtering,httpx,pagination,policy,sequence}, matches `BENCH_PKGS`,
`Makefile:206-214`), leaving hot-path candidates **kernel/database, jobs, outbox, workflow, auth, mfa,
httpclient** and all adapters unbenched; no reference environment exists." MATRIX CS-16's own consequence
framing: "a perf regression in the transaction manager, job claim loop, or outbox relay is invisible to
the only performance gate the repo has; and PERF-02..05 were framed as wholly blocked on a reference env
when only absolute-SLO gating is." Each of PERF-02..05's own PLAN task tables confirms the specific
current gaps: `BenchmarkDispatch`'s own doc comment states auth/authz are exercised via fakes, no real
DB (PERF-02); one sequential SQL query per org ancestor, worst case `len(ancestors) + 2` round trips
(PERF-03); `SweepSLA` loads ALL due rows unbounded, per-row UPDATE+load+emit (PERF-04); `S3.Stat` full-
downloads and streams through `sha256` when checksum-signed metadata is absent, no required-checksum
enforcement (PERF-05).

## Scope

- PERF-02 T1-T5: reference-environment stand-up (T1, shared prerequisite); DB-backed benchmarks across
  6 workload profiles; cold/warm cache × 1/10/100-concurrent-tenant variants; cost-breakdown
  attribution; publication against `perf/reference-v1.json` (S001).
- PERF-03 T0-T6: index-definition audit (T0, gap-fill); set-based rules-resolution query design;
  index confirmation; `EXPLAIN` fixtures; result-parity + SQL-count-constant tests; live-update-
  visibility non-regression; before/after publication (S002).
- PERF-04 T1-T8: bounded-batch sweeper claiming; set-based UPDATE conversion; a partial index on
  `remind_after`; batch-loaded webhook endpoints; the leased-state-machine outbox rework (hard dependency
  on W04's DATA-02/DATA-03 lease primitives); queue-lag/batch-duration metrics; bounded-batch
  benchmarks; before/after publication (S003).
- PERF-05 T1-T5: required-checksum-on-upload enforcement; the bounded, labeled repair path; dedicated
  fallback-invocation metrics; resumable async legacy-object backfill; before/after publication
  (S004).
- CS-16's own 7-package bench-coverage expansion, folded into S004 per this wave's own task brief:
  benchmarks + budget entries for `kernel/database` (tenant-tx open/commit), `jobs` (claim/finalize
  loop), `outbox` (relay dispatch batch), `workflow`, `auth` (token verify), `mfa` (TOTP derive), and
  `httpclient` (guarded dial).

## Out of scope

- **DEC-Q9's own resolution** (whether a dedicated bare-metal runner replaces the provisional GH-runner
  default) — a later SRE decision per REVIEW §F row 9's own framing, not this epic's own scope to make;
  this epic proceeds against the provisional default.
- **PERF-01's own token-bucket sweep fix** — already `EXECUTED` and verified at W00-E01-S002; not
  re-implemented here.
- **PERF-06's own gate-mechanism fixes** (T1, fail-closed missing-benchmark path) — already `EXECUTED`
  at W00-E01-S002; PERF-06's own remaining T3/T4 fuzz scope is W07-E02-S002's scope (owned by REL-04 T8
  per CONFLICT-02), not this epic's.

## Source requirements

PERF-02, PERF-03, PERF-04, PERF-05. MATRIX CS-16 is the consolidated closure spec covering the bench-
coverage-expansion portion.

## Architectural context

This epic groups PERF-02..05 plus CS-16's coverage expansion because all five share the identical
structural constraint MATRIX CS-16 names explicitly: each is individually a real, well-specified
performance finding with its own task table, but all were originally framed (per PLAN's own §5.5
heading) as "blocked on §14 reference environment" — a framing REVIEW §12 and MATRIX CS-16 both correct:
"Relative/container now (REVIEW §12); absolute SLO gated on DEC-Q9." `impl/analysis/wave-allocation-
detail.md`'s own W07-E01 grouping states this exactly: "S001 request-benchmarks-real-pg (PERF-02
relative); S002 rules-resolution-sql (PERF-03); S003 sweeper-materialization (PERF-04); S004
checksum-behaviour-and-bench-coverage (PERF-05 + CS-16's 7 hot-path package benchmarks + budgets).
DEC-Q9 tracked at epic level; absolute-SLO ACs conditional." The per-finding story split (S001-S004,
one per PLAN finding) mirrors PLAN's own §5.5 structure exactly, with CS-16's own cross-cutting
bench-coverage work folded into S004 (the last story) rather than given a fifth story, since CS-16's own
7-package list is itself sized similarly to PERF-05's own T1-T5 and shares the same "no reference-env
blocker" property.

## Included stories

- **W07-E01-S001 — request-benchmarks-real-pg** (PLAN PERF-02 T1-T5): the §14 reference-environment
  stand-up (T1, shared prerequisite across all four PERF stories in this epic) plus DB-backed
  benchmarks, concurrency-matrix variants, cost-breakdown attribution, and publication.
- **W07-E01-S002 — rules-resolution-sql** (PLAN PERF-03 T0-T6): the index-audit gap-fill plus the
  set-based query rewrite, index confirmation, parity/SQL-count tests, and publication.
- **W07-E01-S003 — sweeper-materialization** (PLAN PERF-04 T1-T8): bounded-batch sweeper/webhook fixes
  plus the leased-state-machine outbox rework, consuming W04's own DATA-02/DATA-03 lease primitives.
- **W07-E01-S004 — checksum-behaviour-and-bench-coverage** (PLAN PERF-05 T1-T5 plus MATRIX CS-16's
  7-package benchmark expansion): required-checksum enforcement, bounded repair path, resumable
  backfill, plus the 7 named hot-path benchmarks and their budget entries.

## Dependencies

No dependency on any other W07 epic — this epic's performance-programme scope is disjoint from W07-E02
(verification-hardening), W07-E03 (product-alignment), and W07-E04 (closure), though W07-E04-S001's own
final-gate re-run consumes this epic's own closure state as one of many inputs. This epic depends on
W04-E01/E02 (DATA-02/DATA-03 lease primitives) for S003's own T5, and transitively on this wave's own
all-prior-waves entry gate.

## Risks

RISK-W07-001 (DEC-Q9 remaining unresolved, leaving absolute-SLO acceptance criteria permanently
conditional) originates at wave scope and lands entirely within this epic's four stories. See `risks.md`
for the epic-scoped elaboration.

## Required decisions

DEC-Q9 (reference-performance-environment ownership) is tracked at this epic's own level, per this
wave's task brief instruction ("DEC-Q9 must be tracked at the EPIC level... since it constrains S001
most directly but the whole epic's absolute-SLO framing depends on it"), not buried in only S001's own
story. REVIEW §F row 9's own provisional default (a Linux amd64 GitHub Actions runner + committed
`perf/reference-v1.json`, advisory-only initially) is already in effect and unblocks this epic's
relative/container work; DEC-Q9's own full resolution (a dedicated bare-metal runner, a later SRE
decision) remains open and is not required for this epic's own closure.

## Epic acceptance criteria

- **AC-W07-E01-01**: PERF-02's relative/container evidence is published against `perf/reference-v1.json`
  across all 6 workload profiles and all cold/warm × 1/10/100-concurrent-tenant variants, with cost
  attributed by pool wait / tx setup / authz query / handler query / serialization / middleware
  separately. Absolute-SLO acceptance is explicitly conditional on DEC-Q9.
- **AC-W07-E01-02**: PERF-03's set-based rules-resolution query preserves exact precedence semantics,
  proven by result-parity tests; SQL-count-constant-with-depth holds across 3/10/50-level ancestries;
  live per-request rule updates continue passing (no B13-style stale-read regression).
- **AC-W07-E01-03**: PERF-04's bounded-batch sweeper/webhook fixes hold fixed query count and memory
  across due-row cardinalities (10/1k/100k); the leased-state-machine outbox rework passes its inherited
  crash/duplicate-worker tests from DATA-02/DATA-03's own gate, with no outer transaction spanning
  tenant handlers.
- **AC-W07-E01-04**: PERF-05's required-checksum enforcement proves no body download on a normal `Stat`
  call; the full-hash fallback is reachable only via a labeled repair invocation; the resumable backfill
  survives an interrupt-and-resume cycle with no duplicate work.
- **AC-W07-E01-05**: BENCH_PKGS covers the 7 MATRIX CS-16-named hot-path packages, each with a passing
  bench-budget entry; `make bench-budget` exits 0 with the new entries present.
- **AC-W07-E01-06**: All four stories have passed independent review per mandate §14, specifically
  confirming every absolute-SLO acceptance criterion across the epic is genuinely conditional on DEC-Q9
  in the story's own written text, not silently asserted as unconditional.

## Closure conditions

All four stories reach `accepted`; AC-W07-E01-01 through AC-W07-E01-06 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date; DEC-Q9's
continued-open status (if still unresolved at this epic's own closure) is recorded honestly, not
silently presented as settled.
