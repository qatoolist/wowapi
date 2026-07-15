# EV-W02-E05-S001-002 — Idempotency + RLS-posture test report

Command:
```
go test ./kernel/seeds/... -run 'TestApplyIdempotentNoop|TestApplyRLSPostureRespectsPlatformRole|TestApplyHashStableAcrossOrdering|TestApplyHashExcludesVersionLabel|TestLoadParsesVersion|TestLoadRejectsConflictingVersion' -v
```

Result: PASS.

Key proofs:
- Second `Apply` with identical manifest returns outcome `noop` and leaves row `xmin` unchanged.
- Sync runs as `app_platform`; the role does not have `BYPASSRLS`.
- `app_rt` cannot INSERT into `seed_sync_runs`.
- Canonical hash is stable across declaration ordering and excludes the `version` label.
- Conflicting manifest versions are rejected at load time.
