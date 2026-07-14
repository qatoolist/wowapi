---
id: W00-DEPS
type: wave-dependencies
wave: W00
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W00 — Dependencies

## Upstream (waves this wave depends on)

None. W00 is the first wave in the programme.

## Downstream (waves that depend on this wave)

All of W01-W07 depend on W00's exit gate per `impl/index.md`'s wave map ("Execution order: strictly
W00→W07 for wave entry"). Specific load-bearing dependencies:

| Downstream item | Depends on (from W00) | Why |
|---|---|---|
| W03-E01 (SEC-01) | W00-E02-S003 (D-01 ADR) | D-01 resolves grant-table authority split (framework vs wowsociety) that SEC-01's design assumes |
| W05-E02 (AR-02) | W00-E02-S003 (D-02 ADR) | D-02 resolves the single-Registrar-type-with-typed-keys design AR-01 T2/AR-02 T1 implement directly |
| W05-E01 (AR-01) | W00-E02-S003 (D-03 ADR) | D-03 resolves post-seal-mutation error-vs-panic policy that AR-01 T8/AR-04 T4 implement |
| W04-E04 (DATA-08 W6) | W00-E02-S003 (D-04 ADR) | D-04 resolves the hash_version discriminator design W6-T1 implements |
| W06-E03 (REL-01) | W00-E02-S003 (D-05 ADR) | D-05 resolves GoReleaser split-mode approach for REL-01 T6 |
| W05-E04 (SEC-04) | W00-E02-S003 (D-06 ADR) | D-06 resolves cross-pod cache invalidation transport (epoch table, not message bus) |
| W03-E02 (SEC-06) | W00-E02-S003 (D-07 ADR) | D-07 resolves JWKS-client governance model |
| W01-E02 (FBL-06) | W00-E02-S003 (D-08 ADR) | D-08 resolves pgx query tracing approach (thin in-kernel tracer, not otelpgx) |
| W01 (secrets docs, CS-25) | W00-E02-S003 (D-09 ADR) | D-09 resolves secrets rotation contract (restart-based, v1) |
| W01's AR-04/AR-06 remainder tasks | W00-E01-S001 | Re-verification confirms AR-04 T1/AR-06 T1's current state before T2-T5/T2-T3 build on it |
| W03's SEC-02 T4/T5 (ratification) | W00-E01-S001 | Confirms SEC-02's Wave-0 fail-closed fix (T1-T3) is genuinely intact before layering ratification on top |
| W07's PERF-02..06 remainder | W00-E01-S002 | Confirms PERF-01/PERF-06's fixes and the #25-recalibrated budgets are the correct "before" baseline for W07's relative-comparison work |
| W02/W04's DATA-08 W6 tasks | W00-E01-S003 | Confirms the DATA-08 W0 slice (attachment outbox propagation, legal-delivery audit) is intact before W6 widens the hash contract over it |
| W03's SEC-01/DATA-07 | W00-E01-S003 | REL-04 T1-T4's S3/TOTP wiring underpins the parallel-CI pipeline state later waves' test infrastructure assumes is stable |

## Cross-wave dependencies

None beyond the strict W00→W01..W07 entry ordering above — Wave 00 does not itself depend on any
later wave's output (that would violate mandate §15's "a later wave must not be marked ready when
mandatory predecessor capabilities remain unaccepted" — here it's the inverse direction check: W00
has no predecessor to violate).

## External dependencies

- Test infrastructure: Postgres (via `testkit`/`make ci-container`), MinIO (S3-gated REL-04 tests),
  network access for `go mod download` if the dependency inventory step needs to resolve modules not
  already cached.
- No GitHub org-admin action required for any W00 story (that blocker belongs to W06's REL-01/REL-02
  work).

## Tooling dependencies

- `golangci-lint` v2.11.4 (pinned, `Makefile:16`) for the lint baseline in W00-E02-S001.
- `make bench-budget` / `internal/tools/benchbudget` for the bench-budget baseline.
- `go list -m all`, `go mod graph` for the dependency inventory in W00-E02-S002.

## Decision dependencies

W00-E02-S003 is itself the mechanism that resolves D-01..D-09 into ratified ADR files — see
`dependencies.md` table above for which downstream epics each ADR unblocks. No W00 story depends on
an *unresolved* decision; the decisions being ratified in this wave are inputs the wave produces, not
consumes.
