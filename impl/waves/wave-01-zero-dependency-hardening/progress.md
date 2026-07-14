---
id: W01-PROGRESS
type: wave-progress
wave: W01
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01 progress (initial state)

Per mandate §16.2. Populated at programme-creation time; every item below is at its initial status.

## Epic status

| Epic | Title | Status | Stories | Story status breakdown |
|---|---|---|---|---|
| W01-E01 | static-analysis-utilisation | planned | 3 | 3 planned |
| W01-E02 | observability-correlation | planned | 2 | 2 planned |
| W01-E03 | http-hardening | planned | 2 | 2 planned |
| W01-E04 | generator-doc-test-fixes | planned | 3 | 3 planned (incl. 1 investigation-with-decision story: S003) |

## Story status

| Story | Title | Status | Task count | Task status breakdown |
|---|---|---|---|---|
| W01-E01-S001 | zero-cost-linters | planned | 4 | 4 todo |
| W01-E01-S002 | judged-linter-set | planned | 5 | 5 todo |
| W01-E01-S003 | supply-chain-and-hooks | planned | 3 | 3 todo |
| W01-E02-S001 | trace-log-correlation | planned | 2 | 2 todo |
| W01-E02-S002 | pgx-query-tracer | planned | 1 | 1 todo |
| W01-E03-S001 | server-timeouts-and-body-bounds | planned | 3 | 3 todo |
| W01-E03-S002 | central-validation-enforcement | planned | 3 | 3 todo |
| W01-E04-S001 | generator-correctness | planned | 4 | 4 todo |
| W01-E04-S002 | documentation-reconciliation | planned | 3 | 3 todo |
| W01-E04-S003 | e2e-flake-diagnosis | planned | 2 | 2 todo (incl. 1 decision task) |

## Blocked items

None yet — no story has entered `in-progress`.

## Critical dependencies

- W01-E02-S002 (pgx query tracer) depends on W00-E02-S003's D-08 ADR being ratified (design: thin
  in-kernel tracer over the observability port, not `otelpgx`).
- W01-E03-S002 (central validation) is sequenced compatibly with, not blocked by, the future AR-03
  work in W05 — see `wave.md` rationale.
- W01-E04-S001 depends on W01-E04-S001's own T1 (DX-01 T5 scaffold harness) as the shared primitive
  for its generator-output-boots test — internal task-level sequencing, not a cross-story blocker.

## Open decisions

None new to W01 — all design questions this wave implements against (D-08 pgx tracer approach) are
resolved upstream in W00-E02-S003.

## Open risks

See `risks.md`.

## Artifact completeness

0/10 story-level artifact sets populated.

## Evidence completeness

0 evidence records registered.

## Review state

Not yet reviewed.

## Exit-gate readiness

Not ready. 0 of 10 stories accepted.

## Update 2026-07-13 — wave executed and closed

10/10 stories accepted 2026-07-13 (W01-E01: 3, W01-E02: 2, W01-E03: 2, W01-E04: 3).
Independent review gate passed (W01ReviewGate; conductor concurs). Planning-time sections
above retained as-written; story `story.md` front matter and the wave `closure-report.md`
are canonical. Exit gate: ready — see `closure-report.md`.
