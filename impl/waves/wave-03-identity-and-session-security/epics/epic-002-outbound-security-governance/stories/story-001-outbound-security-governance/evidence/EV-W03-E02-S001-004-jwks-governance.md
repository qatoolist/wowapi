---
id: EV-W03-E02-S001-004
type: evidence
task: W03-E02-S001-T004
acceptance_criterion: AC-W03-E02-S001-04
status: accepted
---

# EV-W03-E02-S001-004 — JWKS-client governance gate

Execution command:

```bash
go test -v ./kernel/auth/... \
  -run 'TestNewJWKSKeySource_Prod' \
  -count=1
```

`auth.NewJWKSKeySource` enforces D-07: when `Env` is `prod` and a custom
`Client` is injected with an empty/nil `TrustedIssuers` list, construction fails
closed. Non-prod profiles and default clients are unaffected.

Test output: `EV-W03-E02-S001-004-jwks-governance.txt`.
