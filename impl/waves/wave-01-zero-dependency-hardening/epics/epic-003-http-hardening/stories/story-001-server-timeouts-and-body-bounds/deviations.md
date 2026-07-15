---
id: DEV-W01-E03-S001
type: deviation-record
parent_story: W01-E03-S001
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations — W01-E03-S001

## DEV-W01-E03-S001-001 — `HeaderTimeout` delivered as the existing `ReadHeaderTimeout` key

- **Approved plan / AC text**: AC-W01-E03-S001-01 names a `HeaderTimeout` field among "four new"
  config fields with a 10s default. plan.md unresolved question 2 explicitly flagged this naming
  for implementation-time resolution rather than treating the AC wording as settled.
- **Actual implementation**: three new fields (`ReadTimeout`/`WriteTimeout`/`IdleTimeout`) plus a
  default bump of the EXISTING `ReadHeaderTimeout` from 5s to 10s (option (a) of plan.md Q2).
- **Reason**: Go's `http.Server` exposes exactly one header-read timeout (`ReadHeaderTimeout`),
  already wired in the template. A second `HeaderTimeout` key would wire to nothing distinct and
  create the exact key-confusion RISK-W01-E03-001 warns about.
- **Impact**: all four `http.Server` timeout fields are configured with CS-09's values — the
  closure spec's intent is fully met; only the AC's literal field-name enumeration differs.
  Behavior change: products relying on the old 5s header default now get 10s (MORE permissive,
  no tightening risk); `configs_base.yaml.tmpl`, docs, and `load_test.go` updated consistently.
- **Risks**: none beyond the (permissive-direction) default change; unconditional `> 0`
  validation on `ReadHeaderTimeout` unchanged.
- **Approval**: recorded for the wave review gate (conductor) per mandate §2.6; flagged in
  verification.md AC-01 row.
- **Compensating controls / follow-up**: none required.
