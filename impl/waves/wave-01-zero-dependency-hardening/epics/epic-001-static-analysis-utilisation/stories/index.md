---
id: W01-E01-STORIES-INDEX
type: stories-index
epic: W01-E01
wave: W01
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E01 — Stories index

| Story | Title | Status | Priority | Source requirement | Tasks | Objective |
|---|---|---|---|---|---|---|
| [W01-E01-S001](story-001-zero-cost-linters/story.md) | zero-cost-linters | planned | P1 | FBL-05 | 4 | Enable the zero-cost leak-detection linter set at zero hits; fix noctx/copyloopvar's named prod hits; add pool-lifetime config keys |
| [W01-E01-S002](story-002-judged-linter-set/story.md) | judged-linter-set | planned | P1 | FBL-07 (part) | 7 | Enable gosec/errorlint/exhaustive/forcetypeassert/usestdlibvars; triage every hit per the named site list (gosec split into G704/G115/G304 tasks, plus a T007 closure task — see story `tasks/index.md` grouping rationale); reject wrapcheck/revive |
| [W01-E01-S003](story-003-supply-chain-and-hooks/story.md) | supply-chain-and-hooks | planned | P1/P2 | FBL-07 (remainder) | 4 | `go mod verify` in CI; license-scanning signal; nightly-fuzz-schedule confirmation (own task, see story `tasks/index.md` grouping rationale); pre-push hook DB-silent-skip fix |
