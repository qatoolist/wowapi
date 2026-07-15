---
id: W07-E03-DEPS
type: epic-dependencies
epic: W07-E03
wave: W07
status: blocked
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E03 — Dependencies

## Upstream (epics/waves this epic depends on)

- **W02-E01/E02** (DATA-09, DATA-01) for PROD-01's own verification.
- **W05-E05** (FBL-01) for PROD-02's own verification.
- **W04-E04-S003, W01-E03-S001** (DX-07 T1, FBL-09) for PROD-03's own verification.
- **W03-E01** (SEC-01 T1/T5) for PROD-04's own verification.
- **W04-E04-S001, W00-E02-S003** (D-04, `hash_version`) for PROD-05's own verification.

Their lifecycle records satisfied the wave entry gate, but direct consumability did not: W02's
DATA-01 parent key is absent for `rule_versions`, and W03-E01-S004's rollout document is stale.
Those two upstream inputs remain blocking dependencies.

## Downstream (epics/waves that depend on this epic)

| Downstream item | Depends on (from this epic) | Why |
|---|---|---|
| W07-E04-S001 (final verification gate, this wave) | This epic | The final gate's own re-run scope includes confirming PROD-01..05's own coordination-artifact status as one of many programme-wide inputs. |

## Internal (within this epic)

Not applicable — single-story epic.

## Cross-wave dependencies

All five listed under "Upstream" above.

## External dependencies

None new. This epic does not read or require access to the wowsociety repository itself — it verifies
that wowapi's own framework-side capabilities are documented as consumable, which is entirely a wowapi-
repository-internal exercise.

## Repository dependencies

**None on wowsociety** — by explicit design, per mandate §2.3's framework/product boundary. This epic's
own verification does not require reading, cloning, or modifying the wowsociety repository.

## Tooling dependencies

None new.

## Decision dependencies

None. See `epic.md` "Required decisions."
