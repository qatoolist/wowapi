---
id: W06-E03-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E03-S002
status: blocked
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E03-S002 — Evidence index

Read-only live API evidence exists under `evidence/tests/`; it proves the activation is absent and
therefore blocks, rather than satisfies, the acceptance criterion.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E03-S002-001 | live API readiness and retest output | W06-E03-S002-T001 | AC-W06-E03-S002-01 (not proven) | `gh api repos/qatoolist/wowapi/branches/main/protection`; `.../environments/release`; `.../rulesets` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | 2026-07-13 and 2026-07-14: branch unprotected (404), environment absent (404), rulesets empty | failed/retested/still blocking (`tests/activation-readiness.txt`) |
| EV-W06-E03-S002-002 | workflow re-verification report | W06-E03-S002-T002 | AC-W06-E03-S002-02 | requires real post-activation GitHub Actions run | none | not executable before DEC-Q10 | not yet produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
