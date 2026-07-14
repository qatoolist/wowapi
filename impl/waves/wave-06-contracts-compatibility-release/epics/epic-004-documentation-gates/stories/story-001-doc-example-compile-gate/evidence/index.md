---
id: W06-E04-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E04-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E04-S001 — Evidence index

Per mandate §10. Raw outputs and complete evidence records were produced under `evidence/tests/`.
Each record includes the execution command, pinned HEAD lineage plus shared working-tree qualifier,
environment/tool versions, timestamp, result, checksum, and independent reviewer evidence.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E04-S001-001 | extractor-run report | W06-E04-S001-T001 | AC-W06-E04-S001-01 | `go run ./internal/tools/docexamples -root .` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS — 1 tagged example compiled | produced (`tests/EV-W06-E04-S001-001.md`) |
| EV-W06-E04-S001-002 | make docs-check execution output | W06-E04-S001-T002 | AC-W06-E04-S001-02 | `make docs-check` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS — same gate invoked locally/CI | produced (`tests/EV-W06-E04-S001-002.md`) |
| EV-W06-E04-S001-003 | adversarial fixture report (initial) | W06-E04-S001-T002 | AC-W06-E04-S001-03 | focused removed-symbol test | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS, superseded after diagnostic normalization | superseded (`tests/EV-W06-E04-S001-003.md`) |
| EV-W06-E04-S001-003-R1 | adversarial fixture report (retest) | W06-E04-S001-T002 | AC-W06-E04-S001-03 | `go test ./internal/tools/docexamples -run TestRemovedSymbolFixtureFailsAtDocumentationLocation -v` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS — stable location and removed-symbol diagnostic | retested (`tests/EV-W06-E04-S001-003-R1.md`) |
| REV-W06-E04-S001-001 | independent review report | W06-E04-S001-T003 | AC-W06-E04-S001-01/02/03 | review-only agent inspection (no command log supplied) | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS — correct, confidence 1, no issues | produced (`reviews/REV-W06-E04-S001-001.md`) |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
