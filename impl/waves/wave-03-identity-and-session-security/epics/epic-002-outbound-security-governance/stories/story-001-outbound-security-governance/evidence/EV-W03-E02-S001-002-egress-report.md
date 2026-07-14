---
id: EV-W03-E02-S001-002
type: evidence
task: W03-E02-S001-T002
acceptance_criterion: AC-W03-E02-S001-02
status: accepted
---

# EV-W03-E02-S001-002 — Boot-time egress-exception report sample

Execution command:

```bash
go run /tmp/egress_sample.go
```

(The sample program builds a `config.Framework` with allowlist + trusted-issuer
entries and a secret DSN, then prints `Framework.EgressExceptions()`.)

Sample output, confirmed credential-free:

```json
[
  {
    "kind": "webhook_allowed_hosts",
    "values": [
      "relay.internal.example",
      "partner.example.com"
    ]
  },
  {
    "kind": "webhook_allowed_cidrs",
    "values": [
      "10.0.0.0/8",
      "192.168.1.0/24"
    ]
  },
  {
    "kind": "jwks_trusted_issuers",
    "values": [
      "https://idp.example.com",
      "https://idp2.example.com"
    ]
  }
]
```

The framework DSN (`secretref://env/DATABASE_URL`) does not appear in the
report; only the configured egress-exception values are enumerated.

Unit-test evidence: `EV-W03-E02-S001-002-egress-report.txt`.
