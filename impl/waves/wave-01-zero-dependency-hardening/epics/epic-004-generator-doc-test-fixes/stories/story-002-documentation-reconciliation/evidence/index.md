---
id: W01-E04-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E04-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04-S002 — Evidence index

Per mandate §10. All three planned records produced 2026-07-13. Common fields for all records
below (stated once, per evidence-policy field list): **code revision** = pinned at
`0a31186cada5c275a588c74081cf977adf346e61` (HEAD at capture start; each record also cites the
working-tree diff it captures, since workers do not commit). **Carry-forward note (evidence-policy
"explicitly carried forward" path):** HEAD advanced to
`05dce5c8a548f7dce3222637ab2c82024236a2a0` mid-story (conductor's Wave-00 acceptance commit);
`git diff 0a31186 05dce5c8` touches only added `impl/` governance files — zero changes under
`docs/`, `internal/`, `cmd/`, `kernel/`, `adapters/`, or `go.mod` — so no re-run is required for
any AC here; the HEAD-clean retest in `reviews/ev-002-command-log.md` was additionally executed
from a `git archive HEAD` extraction at the advanced SHA and reproduced all results.
**branch** = `main` (working tree); **environment** = local macOS (darwin/arm64),
go1.26.5; **date** = 2026-07-13T07:25Z; **executed by** = W01Docs (wave-01 worker);
**reviewer** = developer-experience lead (per-task verification tables) — conductor acceptance
pending, records marked `produced` until then.

| Evidence ID | Type | Task | ACs proven | Execution command | File | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W01-E04-S002-001 | Doc diff (review) | T001 | AC-W01-E04-S002-01 | `git diff docs/implementation/premier-framework-implementation-plan.md` | `reviews/ev-001-t001-plan-doc.diff` | §6 DX-05 row now reads `**[EXECUTED — T1+T2 only, see §9; T3–T5 PLANNED]**`, matching §9's record (§9 line "AR-05 T1/T2 + DX-05 T1/T2 … EXECUTED", "DX-05 T3-T5 … remain PLANNED"); no §6-vs-§9 contradiction remains — PASS | produced |
| EV-W01-E04-S002-002 | Doc diff + decision table + command log | T002 | AC-W01-E04-S002-02 | `git diff docs/blueprint/11-framework-distribution-and-consumption.md`; CLI runs per command log | `reviews/ev-002-t002-blueprint11.diff`, `reviews/ev-002-command-log.md`, `../artifacts/dx05-t3-cli-example-decision-table.md` | 20/20 blueprint CLI examples decided (14 keep, 4+2 implement-corrections, 2 delete); every stale form's failure captured before correction (fail-first); every corrected form verified against a HEAD-built binary; T4 design note states the S001 plumbing dependency explicitly; T5 deferral recorded with REL-03/W06 cross-reference — PASS | produced |
| EV-W01-E04-S002-003 | Coordination note (review) | T003 | AC-W01-E04-S002-03 | read-only inspection of `wowsociety/docs/upstream/` + IRC confirmation with S001 owner | `../artifacts/fbl03-wowsociety-register-coordination-recommendation.md` | PF-2 contingency on S001's DX-02 stated explicitly (not softened); PF-6/RFF-001 targeted as already-resolved per REVIEW Answer 18, with exact register files and index rows named — PASS | produced |

The produced documents serve as both artifact and evidence for this documentation-only story
(dual role recorded at planning time, retained). Checksums omitted (text files under version
control — "where appropriate" carve-out). No superseded records.
