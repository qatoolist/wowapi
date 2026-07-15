---
id: W06-E04-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E04-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E04-S002 — Evidence index

Per mandate §10. Raw outputs and complete evidence records were produced under `evidence/tests/`.
Each record includes the execution command, pinned HEAD lineage plus shared working-tree qualifier,
environment/tool versions, timestamp, result, checksum, and independent reviewer evidence.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E04-S002-001 | integration golden-diff test report | W06-E04-S002-T001 | AC-W06-E04-S002-01 | `go test ./internal/tools/docexamples -run TestGeneratedReferenceByteMatchesAuthoritativeExport -v` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS — generated table byte-matches AR-03 projection export | produced (`tests/EV-W06-E04-S002-001.md`) |
| EV-W06-E04-S002-002 | lint fixture test report | W06-E04-S002-T002 | AC-W06-E04-S002-02 | `go test ./internal/tools/docexamples -run 'TestFutureStateLintRequiresLabelAfterFutureHeading|TestFutureStateLintIgnoresCodeAndCurrentState' -v` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS — unlabeled fails, labeled/current passes | produced (`tests/EV-W06-E04-S002-002.md`) |
| REV-W06-E04-S002-001 | independent review report | W06-E04-S002-T003 | AC-W06-E04-S002-01/02 | review-only agent inspection (no command log supplied) | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` + working-tree changes | PASS — correct, confidence 1, no issues | produced (`reviews/REV-W06-E04-S002-001.md`) |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
