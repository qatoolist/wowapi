---
id: W06-E03-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E03-S003
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E03-S003 — Evidence index

Focused raw evidence is preserved under `evidence/tests/`. Each record names the exact command,
revision, exit status, stdout, and stderr.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E03-S003-001 | seeded-vulnerability fail-then-pass test report | W06-E03-S003-T001 | AC-W06-E03-S003-01 | `scripts/validation/tests/test_trivy_seed.sh` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | seeded vulnerable lock rejected; removal accepted | verified (`tests/seeded-trivy.txt`) |
| EV-W06-E03-S003-002 | waiver-schema fixture test report | W06-E03-S003-T002 | AC-W06-E03-S003-02 | `python3 -m unittest scripts.validation.tests.test_security_contracts` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | scoped active waiver passed; malformed/expired/mismatched waivers rejected | verified (`tests/security-contracts.txt`) |
| EV-W06-E03-S003-003 | forced-private guard-regression test report | W06-E03-S003-T003 | AC-W06-E03-S003-03 | focused security-contract unit tests | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | missing hosted result rejected; private fallback selected | verified (`tests/security-contracts.txt`) |
| EV-W06-E03-S003-004 | seeded-SAST-fixture fallback test report | W06-E03-S003-T004 | AC-W06-E03-S003-04 | `python3 scripts/validation/security_contract.py private-fallback --path .` and focused fixtures | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | local SAST/posture/actionlint/govulncheck/Trivy passed; unsafe seeded pattern rejected | verified (`tests/private-fallback.txt`, `tests/security-contracts.txt`) |
| EV-W06-E03-S003-005 | cross-reference test report | W06-E03-S003-T005 | AC-W06-E03-S003-05 | focused security-contract unit tests | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | duplicate/missing scanner gate entries rejected; release candidate requires artifact/image reports | verified (`tests/security-contracts.txt`) |

## Sequential retest — 2026-07-14

- `python3 -m unittest scripts.validation.tests.test_security_contracts`: 8/8 passed.
- `scripts/validation/tests/test_trivy_seed.sh`: seeded vulnerability rejected; removal accepted.
- `python3 scripts/validation/security_contract.py private-fallback --path .`: local SAST,
  repository posture, actionlint, govulncheck, and Trivy passed.
- `python3 scripts/validation/security_contract.py workflow-policy --visibility auto --source-sha 0a31186cada5c275a588c74081cf977adf346e61`:
  live hosted-scanner meta-check passed for remote `main`.

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
