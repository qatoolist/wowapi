---
id: DEV-W03-E02-S001
type: deviations-record
parent_story: W03-E02-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W03-E02-S001

No deviations from the approved `plan.md`.

- T3 audit sink: `plan.md` left the exact sink TBD and cited structured logging
  as an acceptable option. The implementation chose a structured-log record
  emitted at boot time because no tenant transaction exists at that point.
- T4 trusted-issuer field shape: `plan.md` left the exact shape TBD. The
  implementation added `Security.TrustedIssuers []string` and mirrored it on
  `auth.JWKSConfig`, which matches D-07's requirement for a declared,
  fingerprinted config field.
- T5 fitness-check mechanism: `plan.md` left the mechanism TBD and noted a
  dedicated test was likely simpler than a custom linter. The implementation
  chose an AST-based test.
