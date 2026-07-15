---
id: EV-W02-E01-S003-005
type: consolidated-evidence-bundle
parent_story: W02-E01-S003
commit_sha: 1626b1132622aacc3e85475e4190e16a457ad1f6
---

# DATA-09 six-drill consolidated evidence bundle

| Drill | Test function | Evidence file | Result |
|---|---|---|---|
| N-1 code on expanded N schema | `TestCanaryNAndNMinusOne` N-1 legs | `../tests/EV-W02-E01-S003-001.txt` | pass |
| N code before/after backfill | `TestCanaryNAndNMinusOne` N legs | `../tests/EV-W02-E01-S003-001.txt` | pass |
| Interrupted/resumed backfill | `TestBackfillInterruptedAndResumed` | `../story-002-expand-backfill-validate/evidence/tests/EV-W02-E01-S002-002.txt` | pass |
| Partial fleet rollout | `TestPartialFleetRollout` | `../tests/EV-W02-E01-S003-001.txt` | pass |
| Application rollback after switch | `TestSwitchRollbackAfterSwitch` | `../tests/EV-W02-E01-S003-002.txt` | pass |
| Forward recovery + delayed contract gate | `TestContractGateAndForwardRecovery` | `../tests/EV-W02-E01-S003-003.txt` | pass |

Pipeline run artifact: `EV-W02-E01-S003-004.txt`.
CI pipeline definition: `.github/workflows/migration-drills.yml`.

## Accepted residual risk

Soak duration and error-threshold numeric values remain uncalibrated due to the
absence of a production telemetry baseline (RISK-W02-003). The canary tooling
accepts configurable `SoakConfig` parameters; the values are a per-rollout human
judgment call. This is recorded as accepted residual risk, not silently resolved.
