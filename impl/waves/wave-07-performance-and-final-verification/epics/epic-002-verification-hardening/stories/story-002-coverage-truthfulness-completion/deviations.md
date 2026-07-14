---
id: DEV-W07-E02-S002
type: deviations-record
parent_story: W07-E02-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Deviations record — W07-E02-S002

## DEV-W07-E02-S002-001 — Execution inventory exceeded the historical 22 sites

- **Approved plan:** classify the 22 skip sites inventoried by the 2026-07-11 architecture review.
- **Actual implementation:** the execution-time AST scan found 39 sites after intervening test/package
  expansion. One probabilistic TOTP skip was eliminated; all remaining 38 were classified and
  registered.
- **Reason:** the historical count was evidence at an earlier revision, not a frozen allowlist.
- **Impact:** scope increased; no site was omitted. The manifest now protects the current repository.
- **Risks:** none beyond normal manifest maintenance.
- **Approval:** implementation-time correctness decision; subject to independent review.
- **Compensating controls:** AST validation rejects both new unapproved sites and stale approvals.
- **Follow-up work:** none.

## Resolutions of planned open choices

These are recorded explicitly but are not plan divergences:

- T7 runs on each code-changing PR/merge/main CI execution, scoped to seven DB/S3-backed packages,
  rather than scheduled-only.
- T8 uses 10 seconds per target for the short profile and 1 minute per target for the scheduled
  profile.
- Generated corpus is retained via `.fuzzcache` using per-run save keys and stable restore prefixes;
  proof reports are uploaded as 14/30-day workflow artifacts.
