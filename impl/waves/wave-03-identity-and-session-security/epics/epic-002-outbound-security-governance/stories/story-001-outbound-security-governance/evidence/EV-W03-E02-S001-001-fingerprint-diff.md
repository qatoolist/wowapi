---
id: EV-W03-E02-S001-001
type: evidence
task: W03-E02-S001-T001
acceptance_criterion: AC-W03-E02-S001-01
status: accepted
---

# EV-W03-E02-S001-001 — Fingerprint-diff regression test

Execution command:

```bash
go test -v ./kernel/config/... \
  -run 'TestSharedFingerprintChangesWithOutboundAllowlist|TestSharedFingerprintChangesWithTrustedIssuers' \
  -count=1
```

`SharedSection()` was extended to include `Security` and `Webhook`; therefore
mutations to `Webhook.Outbound.AllowedHosts`, `Webhook.Outbound.AllowedCIDRs`,
or `Security.TrustedIssuers` change `SharedFingerprint()`.

Test output: `EV-W03-E02-S001-001-fingerprint-diff.txt`.
