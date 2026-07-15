---
id: PLAN-W03-E02-S001
type: plan
parent_story: W03-E02-S001
status: ready
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W03-E02-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

No new package. This story adds: a regression test proving `SharedFingerprint()`'s existing scope
(or a small extension to that scope); a boot-time reporting addition to the existing readiness/log
path; an audit-write addition to the existing config-change path; a new trusted-issuer config field
consumed at `prod`-profile boot validation; and a new static fitness check (likely a custom linter
rule or a boot-time/CI-time assertion, exact mechanism TBD).

## Implementation strategy

1. Confirm `SharedFingerprint()`'s current field coverage at this story's actual start commit; write
   a fingerprint-diff test that mutates the allowlist and asserts the fingerprint changes. If the
   test fails (fingerprint does not change), extend `SharedFingerprint()`'s scope to cover the
   allowlist fields, then re-run.
2. Implement the boot-time egress-exception report: enumerate `AllowedHosts`/`AllowedCIDRs` and any
   other configured escape hatch, formatted for readiness/log output, with an explicit review step
   confirming no credential or secret value is included.
3. Implement the allowlist change-audit trail: on a configuration change touching the allowlist,
   write an audit-visible record (exact audit sink — existing `kaudit`-style audit writer or a
   dedicated config-change log — TBD at implementation time).
4. Design and implement the JWKS trusted-issuer config field per D-07 (`ADR-W00-E02-S003-007`): a
   declared, fingerprinted config field; `prod`-profile boot validation rejects a custom JWKS client
   injection with no declared trusted-issuer allowlist.
5. Implement the fitness check: a static assertion (custom linter rule, `go vet` analyzer, or a
   dedicated CI-time test) that allowlist/JWKS-client construction call sites never read from a
   request-scoped or tenant-scoped context value.
6. Write the full test suite for all five tasks.

## Expected package or module changes

`kernel/auth` (`jwks.go`), `httpclient` (`client.go`), the config layer (`SharedFingerprint()`
scope, new trusted-issuer field), the readiness/boot-reporting layer, and — for T5 — either
`.golangci.yml` (if implemented as a custom analyzer, unlikely to be zero-cost/off-the-shelf) or a
dedicated fitness-check test file.

## Expected file changes where determinable

- `kernel/auth/jwks.go:59` — `JWKSConfig.Client`'s governance gate (T4).
- `httpclient/client.go:142` — cross-referenced for T1's fingerprint-scope confirmation, not
  necessarily modified unless the scope-confirmation test fails.
- The config layer's `SharedFingerprint()` implementation (exact file TBD at implementation time).
- The readiness/boot-reporting layer (exact file TBD).
- A new fitness-check test file (T5, exact path TBD).

## Contracts and interfaces

A new trusted-issuer config field is added (T4) — additive to the config struct. No existing public
interface is removed or renamed by this story.

## Data structures

The new trusted-issuer config field's exact shape (a list of trusted issuer URLs/hostnames,
presumably) — to be finalized against D-07's exact wording ("a declared, fingerprinted `config`
field") at implementation time.

## APIs

None affected — this story is boot-time/config-time governance, not a runtime HTTP API change.

## Configuration changes

New trusted-issuer config field (T4). Possibly a new boot-time report toggle if the egress-exception
report (T2) needs to be optionally suppressed in non-prod profiles — a judgment call for
implementation time, not specified here.

## Persistence changes

None anticipated unless the allowlist change-audit trail (T3) requires a durable audit-table write
rather than a log-based record — to be determined against whatever existing audit-writing pattern
the framework uses elsewhere (e.g. `kaudit`).

## Migration strategy

Not applicable unless T3's audit trail requires a new table, in which case a small additive
migration would be needed — not assumed here.

## Concurrency implications

None material — boot-time and config-change-time operations are not concurrency-sensitive in the
way request-path code is.

## Error-handling strategy

T4's `prod`-profile readiness check fails closed: a custom JWKS client with no declared trusted-
issuer config causes readiness to fail, not merely a warning log — consistent with this framework's
established fail-closed config-validation pattern (per epic.md's own citation of "an established
pattern already exists in `config.go`," referencing a comparable SEC-04 T6 precedent).

## Security controls

T4's fail-closed readiness gate is the central security control this story adds. T5's fitness check
is a durable, mechanically enforced invariant preventing future regression.

## Observability changes

T2's egress-exception report and T3's change-audit trail are themselves the observability
deliverables of this story.

## Testing strategy

- T1: fingerprint-diff test — mutate the allowlist, assert `SharedFingerprint()`'s output changes.
- T2: report-output test — confirm every configured egress exception is enumerated, confirm no
  credential/secret value appears in the output.
- T3: change-audit test — mutate the allowlist config, assert an audit-visible record is produced.
- T4: negative-fixture test — boot with a `prod` profile, a custom JWKS client injected, and no
  declared trusted-issuer allowlist; assert readiness fails.
- T5: fitness-check test — assert (via static analysis or a dedicated test walking construction call
  sites) that no allowlist/JWKS-client construction reads request- or tenant-scoped data.

## Regression strategy

Each test above, run in CI, is the regression guard for its own acceptance criterion.

## Compatibility strategy

T1/T2/T3/T5 are additive, non-breaking. T4 is breaking only for a currently-unconfirmed wowsociety
JWKS-client-injection usage (see `story.md` "Compatibility considerations") — this story does not
soften T4's fail-closed behavior to avoid a hypothetical break; it implements the correct governance
gate and records the compatibility risk honestly.

## Rollout strategy

Single story, all five tasks land together since they share the same outbound-security-governance
theme and largely disjoint but related file surface.

## Rollback strategy

Each task's change is independently revertible: T1's fingerprint-scope extension, T2's report, T3's
audit trail, T4's config gate, and T5's fitness check do not share tight coupling beyond touching
the same general area of the config/auth layer.

## Implementation sequence

T1 first (establishes the fingerprint baseline other tasks may want to reference), then T2/T3 (can
proceed in parallel, disjoint concerns), then T4 (the highest-risk task per PLAN's own framing, now
de-risked by D-07's ratified design), then T5 (a fitness check that ideally covers the
newly-added T4 surface too, so sequencing T5 last ensures it checks the final state).

## Task breakdown

- **W03-E02-S001-T001** — Fingerprint-scope confirmation (SEC-06 T1).
- **W03-E02-S001-T002** — Boot-time egress-exception report (SEC-06 T2).
- **W03-E02-S001-T003** — Allowlist change-audit trail (SEC-06 T3).
- **W03-E02-S001-T004** — JWKS-client governance gate, D-07 enactment (SEC-06 T4).
- **W03-E02-S001-T005** — No-tenant-controlled-allowlist fitness check (SEC-06 T5).
- **W03-E02-S001-T006** — Independent review (mandate §14).

## Expected artifacts

Fingerprint regression test; boot-time report implementation; change-audit trail implementation;
JWKS trusted-issuer config field and gate; fitness-check implementation.

## Expected evidence

Fingerprint-diff test log; report-output sample; change-audit test log; JWKS-governance negative-
fixture test log; fitness-check test log.

## Unresolved questions

- Whether `SharedFingerprint()` already covers the allowlist fields structurally (PLAN's own framing:
  "likely already covers... pending a direct scope-confirmation test") — T1's own work answers this;
  not assumed here.
- The exact audit sink for T3 (existing `kaudit`-style writer vs. a dedicated config-change log) —
  to be determined against the framework's existing audit-writing conventions at implementation
  time.
- The exact trusted-issuer config field's shape and validation rule (T4) — to be finalized against
  D-07's ratified ADR text and the existing `JWKSConfig` struct's conventions.
- The exact mechanism for T5's fitness check (custom linter analyzer vs. a dedicated CI-time test
  walking construction call sites) — a judgment call for implementation time; a dedicated test is
  likely simpler than authoring a new golangci-lint-compatible analyzer for a single, narrow
  invariant.
- wowsociety's actual current JWKS-client-injection usage (T4's compatibility risk) — PLAN's own
  admission is that this was not read in its evidence-gathering pass; this story does not invent the
  answer, and flags the need for a follow-up config audit as an out-of-scope item.

## Approval conditions

This plan is approved for implementation once: (a) D-07 (`ADR-W00-E02-S003-007`) is confirmed
ratified, (b) the unresolved questions above are answered by implementation-time investigation, and
(c) the owner and reviewer are assigned.
