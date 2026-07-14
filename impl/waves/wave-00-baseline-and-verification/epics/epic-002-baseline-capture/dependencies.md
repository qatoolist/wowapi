---
id: W00-E02-DEPS
type: epic-dependencies
epic: W00-E02
wave: W00
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W00-E02 — Dependencies

## Upstream (what this epic depends on)

None. W00-E02 has no dependency on W00-E01 (executed-slice-verification) for its own execution —
S001/S002/S003 read and record the current state of coverage, lint, dependencies, and the D-01..09
decisions; none of that requires W00-E01's re-verification work to have completed first.

## Internal sequencing recommendation (non-blocking)

Per `../../wave.md`'s progress rationale, **S003 (adr-ification) should logically follow S001
(quality-baselines)** in execution order, though this is a recommendation, not a hard dependency
enforced in `story.md` `depends_on`:

- Rationale: S001's lint-baseline task captures the current `golangci-lint` state including
  gosec/errorlint/nilerr/exhaustive adjudications; several of those adjudications (nilerr,
  exhaustive, errorlint "deliberate, not gaps" per `requirement-inventory.md` §C) share
  provenance with the same REVIEW/MATRIX passes that produced D-01..D-09. Running S001 first means
  S003's ADR authors are working from a freshly-confirmed baseline snapshot rather than a stale
  one, reducing the (low-severity) risk that an ADR cites a fact about repository state that has
  since drifted.
- This is recorded as an internal sequencing recommendation only. S003 does not list S001 in its
  `depends_on` front matter, and the epic's `plan.md`-equivalent scheduling does not block S003
  from starting before S001 completes — both may proceed in parallel if resourcing prefers that.

## Downstream (what depends on this epic)

Nine downstream epics depend on the D-01..D-09 ADRs this epic's S003 produces. Reproduced from
`../../dependencies.md` (wave-level), scoped to the producing story:

| Downstream item | Depends on | Why |
|---|---|---|
| W03-E01 (SEC-01) | W00-E02-S003 (D-01 ADR) | D-01 resolves grant-table authority split (framework vs wowsociety) |
| W05-E02 (AR-02) | W00-E02-S003 (D-02 ADR) | D-02 resolves single-Registrar-type-with-typed-keys design |
| W05-E01 (AR-01) | W00-E02-S003 (D-03 ADR) | D-03 resolves post-seal-mutation error-vs-panic policy |
| W04-E04 (DATA-08 W6) | W00-E02-S003 (D-04 ADR) | D-04 resolves the hash_version discriminator design |
| W06-E03 (REL-01) | W00-E02-S003 (D-05 ADR) | D-05 resolves GoReleaser split-mode approach |
| W05-E04 (SEC-04) | W00-E02-S003 (D-06 ADR) | D-06 resolves cross-pod cache invalidation transport |
| W03-E02 (SEC-06) | W00-E02-S003 (D-07 ADR) | D-07 resolves JWKS-client governance model |
| W01-E02 (FBL-06) | W00-E02-S003 (D-08 ADR) | D-08 resolves pgx query tracing approach |
| W01 (secrets docs, CS-25) | W00-E02-S003 (D-09 ADR) | D-09 resolves secrets rotation contract |

Additionally, every later wave's "before" baseline claims (coverage/lint/bench/CI regression
comparisons) cite W00-E02-S001's evidence as the reference point, and W04-E02-S003 (FBL-04,
adopting `cenkalti/backoff/v5`) cites W00-E02-S002's dependency inventory confirming that package
is already approved and present-but-unused.

## Cross-wave dependencies

None beyond the downstream table above — this epic does not itself depend on any later wave's
output.

## External dependencies

- `go mod download` / module cache reachability for S002's `go list -m all` / `go mod graph`
  commands.
- `golangci-lint` v2.11.4 binary availability (pinned) for S001's 25-analyzer capture run.
- Real Postgres DB availability (per project history) for S001's coverage-baseline command to be
  measured against the real DB rather than a mocked subset.

## Tooling dependencies

- `golangci-lint` v2.11.4, `make bench-budget` / `internal/tools/benchbudget`, `go test -cover`,
  `go list -m all`, `go mod graph` — see each story's `plan.md` for exact commands.

## Decision dependencies

None block this epic's start (see `epic.md` "Required decisions"). The decision *dependency
direction* runs outward from this epic's S003 to nine downstream epics, not inward.
