---
id: W06-E03-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E03-S001
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W06-E03-S001 — Evidence index

Focused raw evidence is preserved under `evidence/tests/`. Each record names the exact command,
revision, exit status, stdout, and stderr.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W06-E03-S001-001 | unit-test report (malformed-manifest fixture) | W06-E03-S001-T001 | AC-W06-E03-S001-01 | `python3 -m unittest scripts.validation.tests.test_release_contracts` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | 10 tests passed | verified (`tests/release-contracts.txt`) |
| EV-W06-E03-S001-002 | manifest cross-reference report | W06-E03-S001-T002 | AC-W06-E03-S001-02 | same focused unit-test command | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | schema/cross-reference assertions passed | verified (`tests/release-contracts.txt`) |
| EV-W06-E03-S001-003 | workflow test report (seeded-failure attestation) | W06-E03-S001-T003 | AC-W06-E03-S001-03 | same focused unit-test command | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | failed exact-SHA gate blocked candidate | verified (`tests/release-contracts.txt`) |
| EV-W06-E03-S001-004 | workflow syntax/exact-SHA wiring report | W06-E03-S001-T004 | AC-W06-E03-S001-04 | `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 ...` | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | no diagnostics | verified (`tests/actionlint.txt`) |
| EV-W06-E03-S001-005 | seeded-failure fixture report (moving tag) | W06-E03-S001-T005 | AC-W06-E03-S001-05 | focused release-contract unit tests | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | retargeted tag rejected | verified (`tests/release-contracts.txt`) |
| EV-W06-E03-S001-006 | tamper-test report | W06-E03-S001-T006 | AC-W06-E03-S001-06 | focused release-contract unit tests | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | gate/tag/manifest/archive/image/candidate mutations rejected | verified (`tests/release-contracts.txt`) |
| EV-W06-E03-S001-007 | unmanifested-artifact test report | W06-E03-S001-T007 | AC-W06-E03-S001-07 | focused release-contract unit tests | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | extra file rejected; only attested bytes copied | verified (`tests/release-contracts.txt`) |
| EV-W06-E03-S001-008 | clean-verifier golden-failure report | W06-E03-S001-T008 | AC-W06-E03-S001-08 | focused release-contract unit tests | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | SHA/signature/SBOM/provenance/platform/version/hash failures rejected | verified (`tests/release-contracts.txt`) |
| EV-W06-E03-S001-009 | disposable-repo/local-registry dry-run report | W06-E03-S001-T008 | AC-W06-E03-S001-08 | focused release-contract unit tests | `733ef3e930cbb3f89f5bbc53d8f562c60e426513` | temporary git tag and immutable promotion scenarios passed | verified (`tests/release-contracts.txt`) |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
