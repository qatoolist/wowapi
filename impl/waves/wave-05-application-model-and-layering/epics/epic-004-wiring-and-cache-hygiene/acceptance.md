---
id: W05-E04-ACCEPTANCE
type: epic-acceptance
epic: W05-E04
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E04 — Epic-level acceptance

## AC-W05-E04-01 — Constructor-bypass surface closed

The constructor-boundary lint fails CI on a reintroduced ad hoc infrastructure constructor outside
composition packages; the `kernel/kernel.go` audit confirms/refutes, with evidence, that the
closure-captures-a-fresh-instance pattern is isolated to the one already-fixed line. Traces to
W05-E04-S001.

## AC-W05-E04-02 — Authorization cache bounded and epoch-invalidated; DATA-07 T4 closed

The cache never exceeds its configured maximum under adversarial cardinality; idle entries are
evicted with full admission/eviction metrics; N concurrent misses collapse to one DB load via
singleflight; a simulated cross-pod revocation is visible on the second pod without a full TTL wait,
via the D-06 epoch table; `Decision` metadata distinguishes hit from miss/epoch-observed; a `prod`
profile with the cache enabled and no explicit bound fails boot. DATA-07 T4's cache-invalidation
acceptance criterion is closed by this AC. Traces to W05-E04-S002.

## AC-W05-E04-03 — Independent review passed

Both stories have passed independent review per mandate §14. S002's review specifically confirms
SEC-04 T4's epoch-bump wiring is complete across every enumerated framework-side mutation path, given
its own "Highest-risk task" PLAN framing.

## Acceptance authority

Framework architecture lead, jointly with the product-security lead for SEC-04, per
`../../wave.md`'s wave-level acceptance authority.
