---
id: TRACK-TECHNICAL-DEBT-REGISTER
type: register
title: Technical debt register — seed items grounded against repo state at HEAD
status: active
created_at: 2026-07-12
updated_at: 2026-07-13
derived: true
---

# Technical debt register

DERIVED VIEW (seeded). Per mandate §11.9: Debt ID | Origin | Description | Reason accepted |
Impact | Target resolution wave | Related stories | Acceptance authority. Canonical source =
the three named seed items below, grounded against the repository at HEAD `0a31186`; extended as
stories close and introduce or resolve debt (each story's own `implementation.md` "technical debt
introduced" field becomes the canonical source for future entries).

## TD-001 — lint-backlog historical exclusions (errcheck carve-outs)

| Field | Value |
|---|---|
| Debt ID | TD-001 |
| Origin | Historical hardening tranche |
| Description | `.golangci.yml` carries two documented, scoped `errcheck` exclusions: (1) `_test.go` files exclude `errcheck`/`unparam` (test helpers use fixed-value assertions, so these linters are noise there); (2) `internal/cli/` excludes `errcheck` scoped specifically to the `fmt.Fprint(f\|ln)?\(` source pattern (CLI commands write user-facing output to `os.Stdout`/`os.Stderr` in production and `bytes.Buffer` in tests; a failed terminal write has no meaningful recovery — mirrors the stdlib's own `fmt.Print*` exclusion). Confirmed live at `.golangci.yml` lines 34–50; genuine errcheck issues elsewhere in `internal/cli` (pool/file/exec errors) are still caught — this is not a blanket disable. Cross-referenced in `docs/working/lint-backlog.md` (B-1). |
| Reason accepted | Pragmatic tranche-by-tranche lint adoption rather than a single big-bang fix; both exclusions are scoped and documented, not blanket-disabled |
| Impact | Low — scoped and documented; does not suppress errcheck on genuine I/O-error paths |
| Target resolution wave | Opportunistic within W01 (zero-dependency-hardening) — if trivial to close during FBL-05/FBL-07 linter-utilisation stories (W01-E01-S001/S002/S003), close it there; otherwise remains accepted debt |
| Related stories | W01-E01-S001 (FBL-05), W01-E01-S002..S003 (FBL-07) |
| Acceptance authority | framework (Fable 5) |

## TD-002 — single-node rate limiter

| Field | Value |
|---|---|
| Debt ID | TD-002 |
| Origin | REVIEW §29 Q8 ("present-but-immature": rate-limiter fixed but single-node) and REVIEW §K retained-customs list (`K-RETAIN` row in `requirement-inventory.md` §C — justified retention, no work planned) |
| Description | The current rate limiter is a documented interim, single-node design; it has no shared/distributed store, so limits are not coordinated across multiple framework instances |
| Reason accepted | No proven multi-node deployment need yet — same P2-overbuild-avoidance logic applied to B11/B12/B13 in `docs/implementation/framework-backlog-p2-decisions.md` (defer until data/need demands, do not build ahead of demonstrated need) |
| Impact | Medium if the framework scales to multi-node deployments without a shared limiter store (limits would be per-instance rather than global, allowing effective rate to scale with instance count) |
| Target resolution wave | None scheduled — deferred debt. Reopen trigger: a demonstrated multi-node deployment need |
| Related stories | None currently targeted |
| Acceptance authority | REVIEW §K (retained-customs justification) |

## TD-003 — pre-push hook DB-silent-skip

| Field | Value |
|---|---|
| Debt ID | TD-003 |
| Origin | FBL-07 utilisation-closure finding (`requirement-inventory.md` FBL-07 row: "Nightly ci schedule EXISTS since #24 (fuzz portion still seed-replay only)") |
| Description | `.githooks/pre-push` line 21–22 runs `go test ./...` with the explicit comment "unit; DB tests skip without a DSN" — i.e. any test gated on a database connection string silently skips locally (via the test's own `t.Skip` when no DSN is configured) rather than warning the pusher. The authoritative DB-backed gate only runs in CI via `make ci-container` (hook header comment, line 5). Grounded fact: `.githooks/pre-push:21-22` — confirmed as still present at HEAD `0a31186`; this is current, not historical, behaviour. |
| Reason accepted | Local-dev ergonomics when no DB is running — most contributors iterate without a local Postgres instance, and requiring one for every push would slow the inner loop |
| Impact | A contributor can push code that fails DB-dependent tests in CI without any local warning; the failure surfaces only after CI runs the DB-backed gate, adding a feedback-loop delay |
| Target resolution wave | W01 (zero-dependency-hardening) |
| Related stories | W01-E01-S002..S003 (FBL-07 utilisation closure) |
| Acceptance authority | framework (Fable 5) |

## TD-004 — CSRF MaxFormBytes not threaded through SecurityChain

| Field | Value |
|---|---|
| Debt ID | TD-004 |
| Origin | W01-E03-S001 (FBL-09 HTTP hardening) — W01Http report |
| Description | CSRF MaxFormBytes not threaded through SecurityChain — known limitation from W01-E03-S001, follow-up only if a product needs >1MiB form-fallback POSTs |
| Reason accepted | The CSRF middleware's MaxBytesReader bound uses its own default; no current product needs >1MiB form-fallback POSTs, so plumbing the config key through SecurityChain would be speculative |
| Impact | Low — form-fallback CSRF POSTs are bounded at the default limit; only a product needing larger form bodies would notice |
| Target resolution wave | None scheduled — reopen trigger: a product needs >1MiB form-fallback POSTs |
| Related stories | W01-E03-S001 (FBL-09) |
| Acceptance authority | Conductor (Main), W01 review gate 2026-07-13 |

## Note on grounding

TD-001 was verified against `.golangci.yml` directly (both exclusions present, scoped as
described, not blanket-disabled). TD-003 was verified against `.githooks/pre-push` directly (the
DB-skip behaviour is stated in the hook's own comment and is unchanged at HEAD) — no re-scoping
needed; the hook still silently skips DB-gated tests locally exactly as FBL-07 describes.
