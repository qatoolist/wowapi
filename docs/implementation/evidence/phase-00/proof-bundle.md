# Phase 0 — Proof Bundle

## 1. Decision evidence
- D-0001…D-0005 (preflight blueprint fixes: kernel/secrets package, config naming, generated
  configcheck for CLI, no pipe shorthand, acyclicity) — [decisions.md](../../decisions.md), blueprint
  diffs to 04/06/11/12 in this phase's commit.
- D-0006 (minimal module.Context growth path), D-0007 (Go 1.26 floor), D-0008 (version via
  buildinfo + go.mod scan) — [decisions.md](../../decisions.md).

## 2. Discussion evidence
- Question: fold secrets into kernel/config vs separate package → separate `kernel/secrets`
  (D-0001) so adapters implement Provider without importing config; keeps graph layered.
- Question: how can a prebuilt CLI validate product-typed config → generated product-local
  checker `tools/configcheck` (D-0003); plugin loading rejected as runtime magic.
- Question: ship full module.Context now vs grow per phase → grow per phase (D-0006); stubbing 15
  unimplemented accessors would violate the "no broad partial implementations" preflight rule.

## 3. Critique/review evidence
- Two parallel review agents (architecture/boundaries; security/config) ran over the slice;
  findings + resolutions in [review-findings.md](review-findings.md).
- Self-caught defect: module name regex rejected single-char names → cycle test failed first run;
  fixed before review (command-log #6).

## 4. Implementation evidence (files added/changed this phase)
- Blueprint preflight edits: docs/blueprint/{04,06,11,12}-*.md
- Planning: docs/implementation/{phase-plan,progress,decisions,test-strategy,risk-register,readiness-checklist}.md, evidence/README.md
- Code: go.mod; kernel/secrets/secrets.go; kernel/config/{config,secret,moduleview}.go;
  module/module.go; app/app.go; internal/buildinfo/buildinfo.go; internal/cli/cli.go;
  cmd/wowapi/main.go
- Tests: kernel/secrets/secrets_test.go; kernel/config/{config,secret}_test.go; app/app_test.go;
  internal/cli/cli_test.go
- Tooling: scripts/lint_boundaries.sh; Makefile; Dockerfile; deployments/compose.yaml; .golangci.yml

## 5. Verification evidence
- [command-log.md](command-log.md): build/vet/test/race/boundaries/CLI/make-ci all exit 0;
  compose statically validated; Docker daemon + Graphify-extract gaps documented with residual risk.

## 6. Acceptance evidence
- [acceptance-map.md](acceptance-map.md): all Phase 0 exit criteria ✅ except live `make up`
  (deferred with recorded risk R4 and a hard gate before Phase 2).
