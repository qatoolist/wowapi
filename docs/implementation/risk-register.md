# Risk Register

| ID | Risk | Likelihood | Impact | Mitigation | Trigger/monitor | Status |
|---|---|---|---|---|---|---|
| R1 | RLS/tenant isolation subtly wrong (pool state leakage, missing FORCE, bypass path) | M | Critical | `SET LOCAL` inside tx only; TenantDB sole door; catalog-driven `AssertRLSIsolation` sweep; security review gate in Phase 2/4 | isolation test failures; review findings | open |
| R2 | Public API surface churns after external consumers exist | M | High | v0.x until Phase 12; scratch-consumer test from Phase 5 onward; API review before v1 tag | breaking-change diff in CI | open |
| R3 | Package graph decays (cycles, kernel importing upward) | M | High | boundary lint script from Phase 0 (`go list -deps` rules + vocabulary denylist) wired into `make ci` | `make lint-boundaries` red | open |
| R4 | Docker unavailable in some dev environments (observed: daemon not running on author machine at Phase 0) | H | Medium | all container files validated in CI; Makefile targets degrade with clear errors; unit tests run host-side without services | `make up` failures | **realized at Phase 0 — container run deferred, documented in evidence/phase-00/command-log.md** |
| R5 | Config system becomes reflection-heavy/magical | L | Medium | one audited struct binder; strict unknown-key rejection; hot-path immutability tests; 12 §10 anti-patterns lint | config code review | open |
| R6 | Secret leakage via logs/dumps/errors | M | Critical | structural `Secret` type from Phase 0; `AssertNoSecretsInLogs`; CLI output snapshot tests | secret-scan CI job | open |
| R7 | Workflow engine scope creep (Phase 7 biggest build) | M | High | closed step-type set; WorkflowSim test-first; defer vote/override polish behind core approve/reject path | phase 7 slippage | open |
| R8 | River / pgx / goose version drift vs Go 1.26 | L | Medium | pin versions in go.mod; upgrade only at phase boundaries with test runs | dependabot/CI | open |
| R9 | Generated code (CLI templates) drifts from framework API | M | Medium | golden tests compile generated output against the same wowapi version in CI (Phase 10) | golden test failures | open |
| R10 | Evidence discipline decays under velocity | M | Medium | proof bundle is a phase exit criterion; commit checklist includes bundle update | missing bundle at phase end | open |
| R11 | Graphify extraction unavailable (no LLM key configured) | M | Low | `check`/`update` (non-LLM) run regularly; `extract` deferred until key present; noted per phase | graphify check output | open |
