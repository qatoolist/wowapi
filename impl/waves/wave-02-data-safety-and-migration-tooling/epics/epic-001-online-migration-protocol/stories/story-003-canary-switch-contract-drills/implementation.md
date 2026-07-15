---
id: IMPL-W02-E01-S003
type: implementation-record
parent_story: W02-E01-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W02-E01-S003

## What was actually implemented

- Canary/deploy-N tooling in `kernel/migration/canary.go` with configurable soak
  duration/threshold parameters.
- Switch-phase tooling in `kernel/migration/switch.go` (observable compatibility
  flag, application rollback after switch without destructive Down).
- Contract-phase gate in `kernel/migration/contract.go` (fail-closed evidence
  check for no-N-1-remains, forward-recovery dispatch).
- CI drill pipeline in `.github/workflows/migration-drills.yml`.
- Consolidated six-drill evidence bundle in `evidence/pipeline/consolidated-bundle.md`.

## Components changed

- `kernel/migration/canary.go`
- `kernel/migration/switch.go`
- `kernel/migration/contract.go`
- `.github/workflows/migration-drills.yml`

## Files changed

- `kernel/migration/canary.go`
- `kernel/migration/canary_test.go`
- `kernel/migration/switch.go`
- `kernel/migration/switch_test.go`
- `kernel/migration/contract.go`
- `kernel/migration/contract_test.go`
- `.github/workflows/migration-drills.yml`
- `impl/.../story-003-canary-switch-contract-drills/evidence/pipeline/consolidated-bundle.md`

## Interfaces introduced or changed

- `SoakConfig`, `CanaryLeg`, `RunCanary`, `CanaryResult`.
- `CompatibilityFlag`, `SetCompatibility`, `GetCompatibility`, `RollbackAfterSwitch`.
- `ContractGate`, `NoN1Remains`, `RegisterActiveProcess`, `DeregisterActiveProcess`,
  `ForwardRecovery`.

## Configuration changes

Soak duration/threshold configurable via `SoakConfig`. Pipeline trigger is a
nightly cron plus manual dispatch.

## Persistence changes

New `migration.compat_flag` and `migration.active_process` tables created lazily
by `EnsureCompatFlagTable` and `EnsureActiveProcessTable`.

## Migration strategy

No application data migration performed; drills run against fixture migrations.

## Concurrency implications

Contract gate checks the active-process registry; a late-starting N-1 process is
a human operational concern, but the gate fails closed on any positive evidence
of N-1 presence.

## Error-handling strategy

Contract gate returns `ErrContractGateDenied` when evidence is missing or
ambiguous. Forward recovery propagates handler errors.

## Security changes

Contract gate fails closed on missing/ambiguous N-1-absence evidence.

## Observability changes

Compatibility flag is observable via `GetCompatibility`. Soak metrics are
recorded in `CanaryResult`.

## Tests added or modified

- `TestCanaryNAndNMinusOne`
- `TestPartialFleetRollout`
- `TestSwitchRollbackAfterSwitch`
- `TestContractGateAndForwardRecovery`

## Commits

Working tree at base commit `1626b1132622aacc3e85475e4190e16a457ad1f6`.

## Pull requests

Not tracked in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

Soak duration/threshold values are not calibrated; recorded as accepted residual
risk (RISK-W02-003).

## Follow-up items

None.

## Relationship to the approved plan

Matches `plan.md`. The soak-threshold judgment gap is recorded honestly as an
accepted residual risk, not silently resolved.
