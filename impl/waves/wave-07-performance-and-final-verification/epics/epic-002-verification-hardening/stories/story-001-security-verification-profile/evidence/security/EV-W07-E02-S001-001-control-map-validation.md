---
id: EV-W07-E02-S001-001
type: control-map-validation
task: W07-E02-S001-T001
acceptance_criteria:
  - AC-W07-E02-S001-01
status: accepted
---

# EV-W07-E02-S001-001 — Control-map completeness and focused execution

## Required evidence fields

- **Evidence ID:** EV-W07-E02-S001-001
- **Evidence type:** control-map completeness + focused executable-test report
- **Story and task:** W07-E02-S001 / W07-E02-S001-T001
- **Acceptance criterion proven:** AC-W07-E02-S001-01
- **Execution commands:**
  1. `python3 SEC-05/validate_control_map.py`
  2. `python3 SEC-05/test_validate_control_map.py`
  3. `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1 python3 SEC-05/validate_control_map.py --run-tests`
- **Code revision / commit SHA:** `733ef3e930cbb3f89f5bbc53d8f562c60e426513` (working-tree execution; see revision caveat below)
- **Branch:** `main`
- **Execution environment:** Darwin 25.5.0 arm64; PostgreSQL `localhost:5432/wowapi`; required DB and S3 flags set as shown
- **Relevant tool versions:** `go version go1.26.5 darwin/arm64`; `Python 3.14.2`
- **Date/time:** 2026-07-13T21:17:50Z
- **Result:** PASS — 412/412 source inventory entries and their committed inventory digests resolved; 33 applicable controls mapped to executable tests; 379 explicitly not applicable with rationale; 0 waivers. Five focused Go package invocations and six validator regression tests passed.
- **File/URI:** `SEC-05/control-map.json`, `SEC-05/control-map.md`, `SEC-05/validate_control_map.py`, `SEC-05/test_validate_control_map.py`
- **Checksums:**
  - `control-map.json`: `890cacbcc71ce0c1b8fbbe8a24c0f618badea5b570559920f10a1479aa44bf1a`
  - `control-map.md`: `002923b2129b683592c67409d64db293d9d85437cbe47d0c03572c8ee92d2a44`
  - `validate_control_map.py`: `7bf72732c1f66f71ade63dcd384941263337c3a3c4fde005b3c367e9a81a54de`
  - `test_validate_control_map.py`: `3e92c9dc4a69cb5b544b98106056589260267cc3f11e5ebb8146061695c3eb1c`
- **Reviewer:** W05ReviewGateFinal — PASS; no open actionable story-scope issue (EV-W07-E02-S001-004)
- **Superseded evidence:** not applicable (first execution)

## Observed output

```text
control-map valid: total=412 applicable=33 not-applicable=379 waived=0
foundation/webhook: ok
kernel/auth: ok
kernel/authz: ok
kernel/database: ok
kernel/httpclient: ok
validator regression tests: Ran 6 tests ... OK
```

The validator compared the map with all 345 ASVS 5.0.0 CSV requirements, all ten OWASP API Security Top 10 2023 categories, and all 57 normative units in the pinned final NIST SP 800-63-4 main-publication inventory. It also resolved every linked test function in its named source file before executing the de-duplicated focused set.

## Revision caveat

The repository is a shared concurrent execution workspace with unrelated staged, unstaged, and untracked changes. This record identifies the observed HEAD and hashes the story-owned artifacts, but it must be re-run against the eventual clean integration commit before it can be treated as final proof under `impl/governance/evidence-policy.md`'s strict commit-pinning rule. It is not used to claim the blocked story is accepted.
