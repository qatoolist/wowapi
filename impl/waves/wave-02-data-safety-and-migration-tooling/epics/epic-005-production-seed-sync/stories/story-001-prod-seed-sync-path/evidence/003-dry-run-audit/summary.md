# EV-W02-E05-S001-003 — Dry-run + audit-record test report

Commands:
```
go test ./kernel/seeds/... -run 'TestApplyDryRunNoWrites|TestApplyRecordsAuditRow' -v
go test ./internal/cli/... -run 'TestSeedSyncDBDryRun' -v
```

Result: PASS.

Key proofs:
- `Apply` with `DryRun: true` writes nothing to the catalog tables and emits a deterministic plan.
- A successful `Apply` writes a `seed_sync_runs` row with `manifest_hash`, `actor`, `outcome`, and
  `counts` JSONB.
- The CLI `--dry-run` flag produces a plan and exits 0 without writes.
