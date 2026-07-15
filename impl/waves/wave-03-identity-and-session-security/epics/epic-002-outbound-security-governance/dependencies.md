---
id: W03-E02-DEPS
type: epic-dependencies
epic: W03-E02
wave: W03
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E02 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W00-E02-S003** — this epic's S001 references `ADR-W00-E02-S003-007` (D-07) as an already-
  ratified design premise for T4's JWKS-client governance gate.

## Downstream (epics/waves that depend on this epic)

None recorded — no other epic in `impl/analysis/wave-allocation-detail.md` or `requirement-
inventory.md` names a dependency on SEC-06.

## Internal (within this epic)

Single story (S001) — no internal cross-story dependency.

## Cross-wave dependencies

None beyond the W00-E02-S003 upstream dependency stated above.

## External dependencies

None beyond what W00/W01/W02 already establish.

## Repository dependencies

wowapi-internal. Per `requirement-inventory.md` row SEC-06 and PLAN's own wowsociety-impact note:
"Affected, config-only, not source-code... wowsociety never constructs `httpclient.New`/
`auth.JWKSConfig` directly (wired by wowapi's `kernel.New`), and configures OIDC/JWKS purely via
static YAML... with no tenant/user-controlled data feeding these values." PLAN also flags a
"Genuine evidence gap... wowsociety's actual deployment config for allowlist entries or custom
JWKS-client injection was not read in this pass — needs a follow-up config audit," and states
breaking exposure only for T4, "only if wowsociety currently injects a custom JWKS client with no
declaration path (unconfirmed)." This epic's S001 should re-confirm this at implementation time
rather than assume it, consistent with the wave-01 pattern of re-verifying cited evidence.

## Tooling dependencies

None beyond `SharedFingerprint()`'s existing mechanism (T1 confirms/extends its scope, does not
replace it).

## Decision dependencies

- D-07 (`ADR-W00-E02-S003-007`) — from W00-E02-S003, referenced by S001, not re-authored.
