---
id: EV-W03-E02-S001-003
type: evidence
task: W03-E02-S001-T003
acceptance_criterion: AC-W03-E02-S001-03
status: accepted
---

# EV-W03-E02-S001-003 — Allowlist change-audit trail

Execution command:

```bash
go test -v ./kernel/config/... \
  -run 'TestRecordAllowlistChange' \
  -count=1
```

`config.RecordAllowlistChange` emits a redacted `AllowlistChange` record
(containing old/new `AllowedHosts` and `AllowedCIDRs`, no secrets) whenever the
allowlist differs from the previous configuration.

Test output: `EV-W03-E02-S001-003-allowlist-audit.txt`.
