---
id: DEV-W03-E01-S001
type: deviations-record
parent_story: W03-E01-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W03-E01-S001

No deviations from the approved `plan.md` are recorded for this story.

The implementation:

- Created the `identity_grant` migration at the then-next available sequence number
  (`00039_identity_grant.sql`) rather than a pre-assumed number, exactly as `plan.md` instructed
  ("numbering to follow the existing `NNNNN_description.sql` convention, e.g. following on from
  whatever the highest-numbered migration is at this story's start commit").
- Used the exact column set specified in `story.md`.
- Applied RLS FORCE, the one-active-grant-per-actor partial unique index, and
  `app_platform`-only write grants as required.
- Implemented `ActiveTenantAccess` against `user_tenant_access` and called it unconditionally from
  `Verifier.Actor`.
- Rejected zero-UUID tenant claims before any `database.WithTenantID` call.

Out-of-scope fixes required for a green build/test run (concurrent in-flight work) are listed in
`implementation.md` under "Files changed" but are not deviations from this story's plan.
