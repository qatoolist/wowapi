# Phase 12 — Proof Bundle (capstone)

Scope (phase-plan row 12): end-to-end acceptance — a scratch product repo scaffolded by the CLI that
builds working api/worker/migrate binaries, an api/worker/migrate smoke, the complete 28-criterion
acceptance map, release notes, and the final review pass. Date: 2026-07-04.

## 1. Decision evidence
D-0059 (Phase 12: `wowapi init` produces framework-wired mains — a blank repo builds working
api/worker/migrate binaries; E2E scaffold-build-run test; final acceptance sweep).

## 2. Discussion evidence
- **`wowapi init` now wires the framework, not stubs.** The generated `cmd/api|worker|migrate` mains
  construct the pool (runtime AS app_rt with the RLS guard), `kernel.New`, `app.Boot`, serve the router
  behind the observability middleware chain + `/healthz`//`/readyz`, and shut down gracefully; the worker
  runs `app.StartWorker`; migrate runs kernel migrations then each module's. Modules are registered
  through a generated `internal/wire/modules.go` the mains share.
- **Config templates ship secret REFERENCES, not raw DSNs.** The scaffolded `configs/local.yaml` uses
  `secretref://env/DATABASE_URL` (raw/empty DSN strings fail `Secret.UnmarshalText` by design) — the
  secret-reference-only guarantee is visible in the scaffold itself.
- **The E2E test is the acceptance, run through the real CLI.** It doesn't hand-write a module; it runs
  `wowapi init`, points the repo at the local framework, and builds it — then, with a database, runs the
  actual migrate binary and curls the actual api binary's `/healthz`. It follows the consumer test's
  offline-skip discipline so it degrades cleanly on a cold module cache.

## 3. Critique/review evidence
`review-findings.md`: the final review pass verified the scaffolded repo builds and runs end-to-end
(migrate applies, api serves `/healthz` 200 with the observability access-log line), the scaffold golden
tests still hold, and the 28-criterion sweep is complete with no unmet criteria. Honest residuals are
enumerated (durable audit_logs writer, OpenAPI strict CI-diff, module auto-registration, OTel export).

## 4. Implementation evidence
Agent: the wired `cmd/api|worker|migrate` + `internal/wire/modules.go` templates, the config-template
secret-ref fix, `init_cmd.go` registration of the new template, and `internal/e2e/e2e_test.go`. Lead:
the 28-criterion `acceptance-map.md`, `CHANGELOG.md` (v0.1.0 release notes), evidence + decision log.

## 5. Verification evidence
`command-log.md`: `go build ./...`; scaffold golden tests; the E2E test (scaffold → build → vet →
migrate → api `/healthz` 200); the external scratch-consumer contract test; host `make ci` (incl.
bench-budget) + `make ci-container` green (no flakes — the Phase-11 root fix holds).

## 6. Acceptance evidence
`acceptance-map.md`: all 28 framework acceptance criteria (blueprint 10 §2) mapped to their delivering
phase and concrete proof — the two consumer-facing E2E criteria (#19 build a working API binary, #22
kernel + module migrations from cmd/migrate) proven by the Phase 12 scaffold-and-run test. Goal 2 is
complete: Phases 0–12 delivered, reviewed, and evidenced. Graphify semantic `extract` remains blocked on
an LLM key (R11) — the only carried-forward item that is environmental, not a framework gap.
