---
id: PLAN-W06-E03-S002
type: plan
parent_story: W06-E03-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E03-S002

Per mandate §8.5. This plan is unusual in that its "implementation" is entirely a human, repo-admin-only
action set — no code is written by this story. Confirmed facts and the explicit human-gating are stated
throughout; nothing about the exact rule-set parameters is invented.

## Proposed architecture

No code architecture — this story's own "architecture" is a GitHub repository-settings configuration:
branch protection on `main`, a protected `release` Environment with required reviewers, and a tag
protection ruleset, configured via the GitHub web console or `gh`/`gh api` by a human with repo-admin
access.

## Implementation strategy

1. **[Human, repo-admin]** Re-confirm the current absence of branch protection, the `release`
   environment, and tag protection via the same `gh api` calls the source review used, at this story's
   own actual start date (fail-first re-confirmation).
2. **[Human, repo-admin]** Configure branch protection on `main`, referencing W06-E03-S001's own gate
   manifest's required status checks where applicable.
3. **[Human, repo-admin]** Create the protected `release` GitHub Environment with required reviewers.
4. **[Human, repo-admin]** Configure a tag protection ruleset for release tags.
5. **[Any tier]** Re-run W06-E03-S001's `publish` job and its unmanifested-artifact rejection test
   against the now-real protected environment, confirming it still passes (this step can be performed by
   a coding agent or any programme worker once the human-only steps 1-4 are complete).
6. **[Any tier]** Record the activated configuration and verification evidence.

## Expected package or module changes

None — no code is produced by this story.

## Expected file changes where determinable

None in the repository's own source tree. This story's own `evidence/index.md` and `artifacts/index.md`
will record the activation's configuration and verification output once performed.

## Contracts and interfaces

None — no code contract is defined or changed.

## Data structures

None.

## APIs

None affected in the code sense — GitHub's own repository-settings API is the surface this story's
human action operates against, not a framework API.

## Configuration changes

The GitHub repository's own settings (branch protection rules, environment configuration, tag
protection ruleset) — not application configuration.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

Not applicable in the code sense; if the activation is misconfigured (e.g. branch protection blocks a
legitimate release), the repo administrator adjusts the configuration directly.

## Security controls

The activation itself is the security control — see `story.md` "Security considerations."

## Observability changes

None beyond GitHub's own platform-provided audit log for protection-related events.

## Testing strategy

- AC-W06-E03-S002-01: live `gh api` call re-run, mirroring the exact calls used to confirm the controls' prior
  absence.
- AC-W06-E03-S002-02: re-run of W06-E03-S001's `publish` job and its unmanifested-artifact rejection test
  against the real protected environment.

## Regression strategy

Once active, GitHub's own platform enforcement is itself the ongoing regression guard — a future attempt
to bypass branch/tag/environment protection is blocked by GitHub directly, not by anything this
programme's own CI configuration must separately re-check.

## Compatibility strategy

Not applicable beyond the direct-push-to-`main` behavior change noted in `story.md` "Compatibility
considerations," which is the intended effect of this story, not an unintended regression.

## Rollout strategy

Single activation event, performed once by a human with repo-admin access. No phased rollout.

## Rollback strategy

If the activated branch-protection/environment/tag-protection configuration proves too restrictive in
practice (blocks a legitimate release path), the repo administrator adjusts the configuration directly
— this is a configuration change, not a code rollback.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–6). Steps 1-4 are strictly human-only and
cannot be performed or simulated by a coding agent; steps 5-6 may be performed by any programme worker
once steps 1-4 are complete.

## Task breakdown

- **W06-E03-S002-T001** — Repo-admin activation (branch protection, release Environment, tag protection
  ruleset) — human-only, blocked until DEC-Q10 resolves.
- **W06-E03-S002-T002** — Post-activation verification (re-run W06-E03-S001's publish job and rejection test
  against the real environment; record evidence).

## Expected artifacts

A record of the activated branch-protection configuration, the `release` environment's required-
reviewer configuration, and the tag-protection ruleset.

## Expected evidence

Live `gh api` call output confirming each of the three controls is active; re-verification output of
W06-E03-S001's `publish` job against the real environment.

## Unresolved questions

- Exact branch-protection rule parameters (required review count, exact required status checks) — left
  to the repo administrator's own judgment at activation time, informed by but not dictated by this
  planning document.
- Exact timing of DEC-Q10's resolution — genuinely unknown; this story's own status may remain blocked
  indefinitely.

## Approval conditions

This plan is approved for implementation — meaning T001 may begin — only once DEC-Q10 is resolved,
i.e., a human with repo-admin access has committed to performing the activation. The plan document
itself (this file) requires only owner/reviewer assignment to exist; it does not require DEC-Q10's
resolution to be drafted, only to be executed.
