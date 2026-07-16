---
id: W02-E02-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E02-S001
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-16
---

# W02-E02-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` are created on first real content, not pre-populated
empty.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E02-S001-001 | fixture-schema test report (`TestScannerEnumerateFixture`) | W02-E02-S001-T001/T002 | AC-W02-E02-S001-01, AC-W02-E02-S001-02 | `go test ./internal/tools/tenantfk/... -run TestScannerEnumerateFixture -v -count=1` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced (corrected 2026-07-16 from stale "TBD"/"not yet produced" — see EV-004 Finding 2) |
| EV-W02-E02-S001-002 | CI-gate negative-fixture test report (`TestScannerGateNegativeFixture`) | W02-E02-S001-T002/T003 | AC-W02-E02-S001-02, AC-W02-E02-S001-03 | `go test ./internal/tools/tenantfk/... -run TestScannerGateNegativeFixture -v -count=1` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced (corrected 2026-07-16, same as above) |
| EV-W02-E02-S001-003 | CLI gate run output (`tenantfk gate` against migrations) | W02-E02-S001-T003 | AC-W02-E02-S001-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go run ./internal/tools/tenantfk gate --since=36 --migrations=migrations` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass ("no migration files to check") | produced (corrected 2026-07-16, same as above) |
| EV-W02-E02-S001-004 | review report (independent review, task-004, re-verifying AC-01/AC-02/AC-03 and confirming CI wiring via `.github/workflows/ci.yml` tenantfk-gate job) | W02-E02-S001-T004 | AC-W02-E02-S001-01, AC-W02-E02-S001-02, AC-W02-E02-S001-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./internal/tools/tenantfk/... -run 'TestScannerEnumerateFixture\|TestScannerGateNegativeFixture' -v -count=1` | HEAD 43b6e12 + remediation working tree 2026-07-16 | pass | produced |

Corrections applied 2026-07-16 by Independent review agent (Claude Sonnet 4.5), dispatched by
Fable 5 conductor (autopsy remediation R-3): EV-001..003's "TBD"/"not yet produced" placeholders
were stale — the underlying `.txt` output files in `evidence/tests/` were genuinely produced at
commit 1626b1132622aacc3e85475e4190e16a457ad1f6 but the index table itself was never updated to
reflect it. This correction does not delete or alter the original `.txt` evidence files, only the
metadata describing them. Environment for EV-004: macOS (darwin/arm64), go1.26.5.
