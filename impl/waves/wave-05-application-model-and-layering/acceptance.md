---
id: W05-ACCEPTANCE
type: wave-acceptance
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05 — Wave-level acceptance

## AC-W05-01 — Ownership-bound application model operational end-to-end

The `ApplicationModel` compiles via `collect → validate → seal → expose read-only snapshot`; the
owner-bound `Registrar` capability type is mintable only by the compiler and rejects cross-module
ownership claims for every declaration class (resource, rules, authz permission registration, and
the ~9+ remaining classes); all exported registry readers return cloned/immutable data; a
deterministic model hash is emitted at startup/readiness; `go test -race` is clean on concurrent
legitimate reads; existing modules (wowapi-internal and wowsociety) boot unchanged through the
legacy adapter. Traces to W05-E01-S001, W05-E01-S002, W05-E01-S003, W05-E01-S004.

## AC-W05-02 — Typed port API and compiled provider graph operational

`port.Key[T]` and the four generic free functions (`Define`/`Provide`/`Require`/`Resolve`) compile
and resolve correctly bound to the AR-01 `Registrar`; the provider graph is boot-time validated
against duplicate providers, missing requirements, undeclared edges, cycles, and invalid scope/
lifetime edges, each proven by a dedicated adversarial fixture; zero `reflect.*` calls occur at
`Resolve` time on request hot paths, proven by benchmark and lint; API/worker/migrate profiles
compile as three projections of one graph with no hand-copied wiring template; the hand-maintained
`kernel/lifecycle` manifest is retired in favor of the generated graph with existing lint-failure
classes still passing; the legacy port adapter compiles/resolves unchanged for any existing caller.
Traces to W05-E02-S001, W05-E02-S002, W05-E02-S003.

## AC-W05-03 — Authoritative manifest and derived projections proven; boot strictness and waivers enforced

A golden-fixture manifest change deterministically produces the expected full projection diff
(route/permission/resource/schema/lifecycle/profile/test/doc) with no other hand-edited file — the
golden-delta test passes as the acceptance gate itself; a lint rule fails on hand-maintained
duplicate identity or an omitted projection; AR-03 T2 (OpenAPI merge) is correctly recorded as
out-of-scope, single-owned by DX-06 (W06), not silently implemented or silently dropped. Separately:
every collector rejects a second write to the same identity; a module declaring a required-but-empty
fragment fails boot; the post-seal error-not-panic contract (D-03) extends to config/namespace/
collector state; a `prod` profile with a required-but-no-op/missing adapter and no waiver fails
readiness by name, the same configuration under `local` succeeds, and a policy-approved waiver
suppresses the failure with an audit record — proven by the named profile × waiver × adapter
integration matrix. Traces to W05-E03-S001, W05-E03-S002.

## AC-W05-04 — Constructor-bypass surface closed; authorization cache bounded and epoch-invalidated

`kernel/kernel.go`'s `orgAncestry`-pattern audit confirms (or refutes, with evidence) that the fixed
closure-captures-a-fresh-instance pattern is isolated to the one already-fixed line; a lint rule
fails CI on any reintroduced ad hoc infrastructure constructor outside composition packages.
Separately: the authorization cache never exceeds its configured maximum under adversarial
cardinality; idle entries are evicted with a full admission/eviction metric set; N concurrent misses
collapse to one DB load via singleflight; a simulated cross-pod revocation is visible on the second
pod without a full TTL wait, via the D-06 per-tenant epoch table; `Decision` metadata distinguishes
cache-hit from cache-miss/epoch-observed; a `prod` profile with the cache enabled but no explicit
max-size/stale-allow bound fails boot validation. This wave's own acceptance criterion explicitly
closes DATA-07 T4's cache-invalidation acceptance criterion (W03-E04 scope) per
`impl/analysis/wave-allocation-detail.md`'s stated cross-wave closure relationship. Traces to
W05-E04-S001, W05-E04-S002.

## AC-W05-05 — Kernel re-homed to the correct four-level layering

`go list ./kernel/...` returns a package count at or below the target-list count (the nine
non-kernel packages — webhook, notify, document, artifact, attachment, comment, bulk, integration,
mfa — moved to `foundation/`; `kernel/storage` retained as the correct port); the extended depguard
rule denying `kernel → foundation` imports and the extended `scripts/lint_boundaries.sh` allowlist
are both green; the boundary-lint fixture that fails today against the nine packages is confirmed to
pass after the re-home; the `kernel/mfa` deprecated forwarding shim is in place; wowsociety's build
and full identity/authz test suite run green against the new `foundation/mfa` path or the shim.
Traces to W05-E05-S001, W05-E05-S002.

## AC-W05-06 — Independent review passed

Every W05 story has passed independent review per mandate §14. W05-E01-S001 (the security-boundary
`Registrar` capability type) and W05-E01-S002 (T5's authz-registration ownership gap, PLAN's own
"widest gap of the six") are specifically checked for genuine adversarial-test proof, not
pattern-matched confidence; W05-E03-S001 is specifically checked for the golden-delta test having
actually run as the acceptance gate PLAN frames it as; W05-E05-S001 and W05-E05-S002 (kernel
re-home, "the largest single architectural correction") are specifically checked for the
`kernel/mfa` shim's correctness and for wowsociety's identity/authz suite having genuinely run green,
not merely asserted.

## Acceptance authority

Framework architecture lead, jointly with the product-security lead for SEC-04 (W05-E04-S002) — per
`wave.md`'s "Acceptance authority."
