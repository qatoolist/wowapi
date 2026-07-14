---
id: W01-E03-RISKS
type: epic-risks
epic: W01-E03
wave: W01
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W01-E03 — Risks

Epic-scoped elaboration of the two wave-level risks (`../../risks.md`) that name this epic's stories
as affected items, plus one epic-specific risk not otherwise recorded at wave level.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W01-002 | FBL-08's boot-time rejection of undeclared-mutating-routes breaks an existing route that currently works only because validation was silently skipped | Medium | Medium — could break an existing deployment path if the profile-flag compat strategy isn't honored | Medium | W01-E03-S002 | Ship behind a profile flag first (explicit in FBL-08's plan note: "compat: profile-flag first"); audit all existing mutating routes for a declared contract before flipping the flag to enforced-by-default | Revert to advisory-only (log, don't reject) if an undeclared route is found in a code path this epic didn't anticipate | unassigned | open | Low once the flag strategy is honored |
| RISK-W01-003 | FBL-09's prod-profile zero-timeout rejection breaks an existing deployment that relies on the current infinite-default connection timeouts | Low | Medium | Low-medium | W01-E03-S001 | Default timeout values are explicitly "safe defaults" per MATRIX CS-09 (read 30s / write 60s / idle 120s / header 10s), not zero — the rejection only fires on an explicit zero-value config, not on unset config falling through to the new defaults | Document the new defaults prominently in the story's implementation record so a downstream deployment isn't surprised | unassigned | open | Low |
| RISK-W01-E03-001 | S001's four new timeout config keys (`ReadTimeout`/`WriteTimeout`/`IdleTimeout`/`HeaderTimeout`) are easily confused with the three already-existing, differently-scoped keys (`ReadHeaderTimeout`, `RequestTimeout`, `MaxBodyBytes` — `kernel/config/config.go:104-114`), risking either accidental duplication of an existing key or an inconsistent validation policy (existing keys reject `<=0` unconditionally at `Framework.Validate()`; this story's task brief specifies a prod-profile-only rejection, mirroring the SSRF-disable pattern instead) | Medium | Low-medium — a naming collision or an inconsistent validation policy is a design-review catch, not a runtime failure, if caught before merge | Low-medium | W01-E03-S001-T001, W01-E03-S001-T002 | `plan.md` explicitly distinguishes the three existing keys (out of scope, already validated unconditionally) from the four new connection-level keys this story adds; T002's design note flags the validation-policy choice (unconditional vs. prod-only) as an explicit implementation-time decision to confirm against the story's stated fail-first tests, not to silently resolve by copying either existing pattern without checking | If the unconditional-vs-prod-only inconsistency is flagged in review, resolve by matching whichever of the two existing precedents the reviewer judges more consistent, and record the choice in `implementation.md`/`deviations.md` if it differs from the task brief's stated prod-only framing | unassigned | open | Low |

## Notes

This epic inherits RISK-W01-002 and RISK-W01-003 from the wave-level register verbatim (same risk
IDs — not re-numbered, per mandate §5's "never reuse an identifier" combined with "never renumber
existing identifiers"). RISK-W01-E03-001 is a new epic-specific risk, identified during this epic's
planning pass against the actual current state of `kernel/config/config.go`, and is not present in
the wave-level register — it may be promoted there if a later wave-level risk review determines it
has broader relevance.
