---
id: EV-W03-E02-S001-005
type: evidence
task: W03-E02-S001-T005
acceptance_criterion: AC-W03-E02-S001-05
status: accepted
---

# EV-W03-E02-S001-005 — No-tenant-controlled-allowlist fitness check

Execution command:

```bash
go test -v ./kernel/config/... \
  -run 'TestFitnessCheck' \
  -count=1
```

`TestFitnessCheckDetectsKnownViolation` proves the AST-based checker catches a
deliberate violation. `TestFitnessCheckKernelAndAppAreClean` walks the framework
source files that construct `httpclient.Config`, `httpclient.New`,
`auth.JWKSConfig`, and `auth.NewJWKSKeySource` and asserts no call site reads
`context.Context` or `*http.Request` data. `TestFitnessCheckTemplateJWKSUsesConfigOnly`
asserts the generated api main template constructs the JWKS source purely from
static config.

Test output: `EV-W03-E02-S001-005-fitness-check.txt`.
