---
id: W05-E03-ACCEPTANCE
type: epic-acceptance
epic: W05-E03
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E03 — Epic-level acceptance

## AC-W05-E03-01 — Manifest and derived-projection tooling proven

The manifest schema (identity + projection inputs) round-trips against ≥1 existing internal fixture
module; a golden-fixture manifest change deterministically produces the expected full projection
diff with no other hand-edited file — the golden-delta test (PLAN's own "this test IS the acceptance
gate") passes; a lint rule fails on hand-maintained duplicate identity or an omitted projection;
AR-03 T2 (OpenAPI merge) is correctly recorded out-of-scope, single-owned by DX-06. Traces to
W05-E03-S001.

## AC-W05-E03-02 — Boot strictness and waiver mechanism proven

Every collector rejects a second write to the same identity; a module declaring a required-but-empty
fragment fails boot; the D-03 error-not-panic contract extends to config/namespace/collector state;
a `prod` profile with a required-but-no-op/missing adapter and no waiver fails readiness by name,
`local` with the same configuration succeeds, and a policy-approved waiver suppresses the failure
with an audit record — proven by the named profile × waiver × adapter integration matrix. Traces to
W05-E03-S002.

## AC-W05-E03-03 — Independent review passed

Both stories have passed independent review per mandate §14. S001's review specifically confirms
AR-03 T3's golden-delta test genuinely ran (not skipped) and genuinely covers the full named
projection surface, given PLAN's own framing of this test as the acceptance gate itself.

## Acceptance authority

Framework architecture lead, per `../../wave.md`'s wave-level acceptance authority.
