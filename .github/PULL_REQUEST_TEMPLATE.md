<!--
PR title MUST follow Conventional Commits, e.g.:
  feat(authz): add step-up challenge for sensitive routes
  fix(outbox): dispatch events in occurred_at order per aggregate
  docs: expand the module SDK guide
Types: feat, fix, docs, refactor, perf, test, build, ci, chore, revert
-->

## What & why
<!-- What does this change do, and what problem does it solve? Link issues: Closes #123 -->

## Type of change
- [ ] Bug fix (non-breaking)
- [ ] New feature (non-breaking)
- [ ] Breaking change (API/behavior)
- [ ] Docs / tooling / CI only

## Checklist
- [ ] `make check` passes locally (fmt, vet, lint-new, tidy, unit tests)
- [ ] Tests added/updated for the change (TDD where applicable)
- [ ] No new kernel domain-specific terms (`make lint-boundaries` clean)
- [ ] Docs updated (SRS/tracker/user-guide) if behavior or surface changed
- [ ] Migrations are reversible (if any were added)
- [ ] Secrets handled by reference only; nothing sensitive logged

## Notes for reviewers
<!-- Anything that needs special attention, risk areas, follow-ups. -->
