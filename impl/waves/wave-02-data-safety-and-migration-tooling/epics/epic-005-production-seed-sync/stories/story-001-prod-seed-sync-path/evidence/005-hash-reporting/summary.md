# EV-W02-E05-S001-005 — Seed/catalog-hash reporting test report

Command:
```
go test ./app/... -run 'TestIntegrationReadinessAfterSyncReportsHash' -v
```

Result: PASS.

Key proof:
- After `seeds.Apply`, `/readyz` returns HTTP 200 with
  `details.seed_catalog_hash` equal to the manifest hash reported by `Apply`.
