# EV-W02-E05-S001-004 — Empty-catalog fail-first/pass-after readiness report

Commands:
```
# Before-state probe (temporary test, removed after capture):
go test ./app/... -run TestIntegrationCS21BeforeProbeProdBootEmptyCatalogsSilentlyReady -v

# After-state tests:
go test ./app/... -run 'TestIntegrationReadinessEmptyCatalogsFailsNamed|TestIntegrationReadinessAfterSyncReportsHash' -v
```

Result: PASS.

Key proofs:
- **Before:** `app.Readiness(booted, fp, nil)` on an empty-catalog, prod-profile boot with declared
  seeds returned HTTP 200 `ready` and no `seed_catalogs` check — the silent deny-everything defect
  captured in `before-probe.log`.
- **After:** `app.ReadinessWithCatalogs(...)` returns HTTP 503 `not_ready` with a `seed_catalogs`
  error naming `seed sync`; after `seeds.Apply` the same endpoint returns HTTP 200 `ready`.
