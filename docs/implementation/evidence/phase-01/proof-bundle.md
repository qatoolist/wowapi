# Phase 1 — Proof Bundle

Scope (phase-plan row 1): full `kernel/config` loader, `kernel/secrets` provider adapter,
`kernel/logging`, app process views + startup/shutdown skeleton, CLI `config
validate/print/schema/doctor` (framework-side). Date: 2026-07-03.

## 1. Decision evidence
D-0011 (yaml.v3 dependency), D-0012 (binder scope: tags for shape, `Validate()` hook for ranges),
D-0013 (envprovider location), D-0014 (Load/LoadDetailed API), D-0015 (unsafe-knob mechanism now,
real knobs later) — recorded in `docs/implementation/decisions.md` **before** the code.
Post-review: D-0016 (Options final shape), D-0017 (environment not downgradable), D-0018 (module
namespaces file-layer only), D-0019 (unsafe judged on final bound values), D-0005 amendment.
Blueprint 12 §2/§3/§4 updated to match (blueprint deviations never silent).

## 2. Discussion evidence
- Defaults duality (tags vs `Defaults()`): resolved by keeping both with a drift-guard test
  (`TestLoadReproducesDefaults`) — D-0012.
- Where unsafe enforcement runs: inline-at-bind (original) vs post-bind over final values —
  security review reproduced two fail-open holes in the former; moved post-bind (D-0019).
- Environment trust: blueprint §1 lets env vars set `environment`; review showed that permits a
  prod→local downgrade. Narrowed: env var may supply, never override, the file value (D-0017).
- Validate() shadowing (ARCH-10): loader-side auto-call rejected (promotion detection = reflection
  magic beyond R5 budget); contract documented, Phase 10 scaffold will generate delegation.

## 3. Critique/review evidence
`review-findings.md`: 19 findings (2 high security, 1 high architecture, 4 medium, rest low/info),
every one fixed with a named regression test or explicitly accepted with rationale. Reviewer
probes reproduced SEC-3/4/5 before the fixes; all reproductions now fail closed.

## 4. Implementation evidence
New/changed files:
- `kernel/config/`: load.go, bind.go, tree.go, fingerprint.go, schema.go (new); config.go
  (conf/default/doc tags), moduleview.go (Namespaces), load_test.go (~40 tests)
- `kernel/logging/`: logging.go + tests (13)
- `adapters/secrets/envprovider/`: provider + tests (5)
- `app/`: views.go (+SectionFingerprints), context.go, run.go + tests (30+)
- `internal/cli/`: config_cmd.go + tests (21), cli.go (dispatch/usage)
- `scripts/lint_boundaries.sh`: internal/cli composition-root rules
- `Dockerfile`: dev stage cgo fix (race suite in container)
- go.mod/go.sum: gopkg.in/yaml.v3 (only third-party dep)
Team: 3 parallel implementation agents (logging / app+adapter / CLI) + lead (loader, security
fixes); 2 parallel review agents; 1 parallel fix agent — per Goal 2 working style.

## 5. Verification evidence
`command-log.md`: TDD red run (#2), green suites per package (#3, #7–#9), integrated `make ci`
host (#10, #14), CLI live smoke (#11), `make up` (#5), `make ci-container` root-caused and fixed
(#12) then green (#13, #14) — **R4 closed**. Graphify checked at phase start (#6) and updated at
phase end (#15).

## 6. Acceptance evidence
`acceptance-map.md` maps all 13 Phase 1 exit criteria to code, named tests, and command-log
entries. No unchecked item. Carried forward explicitly:
- CLI redaction snapshot test → Phase 2 (first framework `Secret` field, DB DSN).
- Module-namespace env-var coercion → Phase 5 (D-0018).
- Graphify semantic `extract` still blocked on LLM key (R11, unchanged).
