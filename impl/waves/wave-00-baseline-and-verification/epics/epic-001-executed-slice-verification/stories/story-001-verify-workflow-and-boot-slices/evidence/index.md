---
id: W00-E01-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W00-E01-S001
status: executed
created_at: 2026-07-12
updated_at: 2026-07-13
derived: false
---

# Evidence index — W00-E01-S001

Per `evidence-policy.md` required fields (mandate §10). All four records below were **actually
executed** on 2026-07-13 against pinned commit `0a31186cada5c275a588c74081cf977adf346e61`
(branch `main`). Raw logs live in `evidence/tests/` (category subdirectory created on first real
content, per policy). Checksums are SHA-256 (first 16 hex chars).

Shared environment for all rows: local workstation, macOS 26.5.2 (Darwin 25.5.0), arm64;
`go1.26.5 darwin/arm64`; Postgres 17-class local instance via `make up` compose
(`wowapi-postgres-1`, healthy), `DATABASE_URL=postgres://wowapi:***@localhost:5432/wowapi?sslmode=disable`;
Docker Desktop container runtime; **concurrent load present** (sibling W00 workers running test
suites on the same machine — noted per wave guidance; none of these records is timing-sensitive).

| Evidence ID | Evidence type | Story and task | AC proven | Execution command | Commit SHA | Branch/tag | Environment | Tool versions | Date/time | Result | File/URI | Checksum | Reviewer | Superseded evidence |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EV-W00-E01-S001-01 | Test-execution log (race-detector) | W00-E01-S001 / W00-E01-S001-T001 | AC-W00-E01-S001-01 | `go test -v ./kernel/workflow/... -race` | `0a31186cada5c275a588c74081cf977adf346e61` | main | shared env above (DB-backed integration tests ran against local Postgres) | go1.26.5 darwin/arm64 | 2026-07-13T12:13:43+05:30 → 12:13:54 | **pass** — exit 0, 0 FAIL, no race warnings; `TestNewRuntimePanicsOnNilDeps`, `TestIntegrationOverrideAuthzGate`, `TestIntegrationOverrideFailsClosedWithoutPermission`, `TestIntegrationWorkflowOverride` all PASS | `evidence/tests/sec02-workflow-race.log` | sha256:0a17e85ea35ecdce | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E01-S001-02 | Test-execution log + full-suite green check | W00-E01-S001 / W00-E01-S001-T002 | AC-W00-E01-S001-02 | `go test -v ./app/... -run Boot` AND `go test ./...` | `0a31186cada5c275a588c74081cf977adf346e61` | main | shared env above | go1.26.5 darwin/arm64 | 2026-07-13T12:14:09 (boot) / 12:15:25→12:17:33 (full suite) | **pass** — both exit 0; `TestBootFailsOnUnknownConfigNamespace` PASS; full suite 57 packages, 0 FAIL | `evidence/tests/ar04-boot-run-boot.log`; `evidence/tests/ar04-full-suite.log` | sha256:d04aec5132af0008; sha256:91427e58ded80d82 | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E01-S001-03 | Test-execution log (race-detector) | W00-E01-S001 / W00-E01-S001-T003 | AC-W00-E01-S001-03 | `go test -v ./kernel/authz/... -race` AND `go test -v ./kernel/ -run 'TestIntegrationRulesResolverOrgAncestry' -race -count=1` (equivalent covering `kernel_rules_test.go`; planned `-run TestKernelRules` matched no tests — see `deviations.md` DEV-01) | `0a31186cada5c275a588c74081cf977adf346e61` | main | shared env above | go1.26.5 darwin/arm64 | 2026-07-13T12:14:12 (authz) / 12:15:13 (kernel rules) | **pass** — both exit 0, no race warnings; sentinel test `TestCachingStoreOrgAncestorsRoutesToComposedInner` PASS; `TestIntegrationRulesResolverOrgAncestry` + `...WithAuthzCache` PASS | `evidence/tests/ar06-authz-race.log`; `evidence/tests/ar06-kernel-rules-race.log` | sha256:b954cb0cbc1c15b0; sha256:97441fa6cb69364c | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E01-S001-04 | Doc-drift grep + interface-diff log | W00-E01-S001 / W00-E01-S001-T004 | AC-W00-E01-S001-04 | `grep -rn "RunAPI\|RunWorker\|RunMigrate" README.md docs/blueprint/` AND method-set diff of `docs/blueprint/06-module-sdk.md` `Context` listing vs `module/module.go` | `0a31186cada5c275a588c74081cf977adf346e61` | main | shared env above (no DB needed) | go1.26.5 darwin/arm64; BSD grep (macOS); git 2.x (`git grep` cross-check at `345e4ce`) | 2026-07-13T12:20 (approx; exact in log header) | **failed** (as AC literally worded) — T2 diff EMPTY (40/40 methods match, pass); T1 grep returned **7 hits** in `docs/blueprint/` (04:15,37-39; 06:207; 10:94; 12:171) instead of zero. README.md: zero hits; blueprint 11: zero hits; no `RunAPI/RunWorker/RunMigrate` function exists in Go source. `git grep` at fix commit `345e4ce` shows the **identical 7-hit set** → no drift since the executed fix; the hits are future-state design prose whose labeling is AR-05 **T5** (planned, W06-E04-S002). See `deviations.md` DEV-02 and `verification.md` findings. | `evidence/tests/ar05-doc-drift.log` | sha256:3f0c10fa413d04f4 | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |

## Notes

- Per `evidence-policy.md`'s revision-pinning rule every row cites the exact commit SHA
  (`0a31186cada5c275a588c74081cf977adf346e61`); no moving-target reference is used.
- EV-W00-E01-S001-04 is preserved with status **failed** per the failed-evidence preservation rule.
  It is *not* retried-until-green and *not* silently reinterpreted as pass. The accompanying
  analysis (identical hit set at the reviewed fix commit `345e4ce`; executed T1 scope —
  `README.md` + blueprint 11 — clean) is recorded for the conductor's adjudication: either the AC
  wording is re-scoped to the executed slice, or the 7 future-state references are routed to
  AR-05 T5's canonical target (`W06-E04-S002`). Any later re-run is a new `retested` record
  referencing this one.
- Log files carry full command output plus a header with command, cwd, commit, start/finish
  timestamps, exit code, and environment (including the concurrent-load note).
