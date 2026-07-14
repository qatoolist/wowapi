---
id: W06-PROGRESS
type: wave-progress
wave: W06
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W06 progress (initial state)

Per mandate §16.2. Populated at programme-creation time; every item below is at its initial status.

## Epic status

| Epic | Title | Status | Stories | Story status breakdown |
|---|---|---|---|---|
| W06-E01 | consumer-and-dsl | planned | 2 | 2 planned (1 design-investigation) |
| W06-E02 | api-contract-gates | planned | 3 | 3 planned (1 with explicit blocked-entry legs) |
| W06-E03 | release-gating | planned | 3 | 3 planned (1 human-gated) |
| W06-E04 | documentation-gates | planned | 2 | 2 planned |

## Story status

| Story | Title | Status | Task count | Task status breakdown |
|---|---|---|---|---|
| W06-E01-S001 | module-dsl-design | planned | 2 | 2 todo (design-investigation shaped, no independent-review task) |
| W06-E01-S002 | golden-consumer-matrix | planned | 6 | 6 todo (incl. 1 independent-review task) |
| W06-E02-S001 | openapi-merge-complete-or-loud | planned | 4 | 4 todo (incl. 1 independent-review task) |
| W06-E02-S002 | compat-gates-buildable-now | planned | 7 | 7 todo (incl. 1 independent-review task) |
| W06-E02-S003 | compat-gates-unblocked | planned | 4 | 4 todo (incl. 1 independent-review task; entry blocked per-leg) |
| W06-E03-S001 | exact-commit-release-pipeline | planned | 9 | 9 todo (incl. 1 independent-review task) |
| W06-E03-S002 | protection-activation | planned | 2 | 2 todo (human-gated; cannot enter ready until DEC-Q10 resolved) |
| W06-E03-S003 | blocking-security-scans | planned | 5 | 5 todo (incl. 1 independent-review task) |
| W06-E04-S001 | doc-example-compile-gate | planned | 3 | 3 todo (incl. 1 independent-review task) |
| W06-E04-S002 | generated-docs-and-labels | planned | 3 | 3 todo (incl. 1 independent-review task) |

## Blocked items

W06-E03-S002 (protection-activation) is recorded as blocked-entry from creation: it cannot move to
`ready`/`in-progress` until DEC-Q10 (repo-admin action) is resolved by a human with repo-admin access,
per `requirement-inventory.md` §B ("DEC-Q10 | Repo-admin activation... | blocked (human)"). W06-E02-S003
(compat-gates-unblocked) has three legs individually blocked on separate unblocking stories (T3 on
E02-S001, T5 on E01-S001's DX-03 design and W05-E03's AR-03 remainder, T7 on E01-S002) — see `story.md`
for the per-leg detail.

## Critical dependencies

- W06-E02-S003's T3 leg depends on W06-E02-S001 (DX-06) reaching `accepted`.
- W06-E02-S003's T5 leg depends on both W06-E01-S001 (DX-03 design) and W05-E03 (AR-03 remainder)
  reaching `accepted`.
- W06-E02-S003's T7 leg depends on W06-E01-S002 (DX-04) reaching `accepted`.
- W06-E03-S001's T6 (GoReleaser split-mode) depends on ADR-005 (`ADR-W00-E02-S003-005`), already
  ratified at W00.
- W06-E03-S002 depends on DEC-Q10 resolution (human, repo-admin) — untracked by this wave's own
  execution capacity.

## Open decisions

DEC-Q10 (repo-admin activation) is open and human-blocked — tracked, not resolved, by this wave. The
DX-06 T2 validator-dependency decision (`pb33f/libopenapi` or equivalent) is open and recorded as an
implementation-time task in W06-E02-S001, not yet made.

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
