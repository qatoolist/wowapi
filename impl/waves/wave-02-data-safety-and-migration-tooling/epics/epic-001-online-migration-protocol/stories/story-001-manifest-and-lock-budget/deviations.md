---
id: DEV-W02-E01-S001
type: deviations-record
parent_story: W02-E01-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W02-E01-S001

## DEV-W02-E01-S001-001 — Manifest block added to W02-E05's migration

The enforcement boundary in `plan.md` requires every kernel migration ≥ 00031 to
have a manifest. The working tree already contained an untracked
`migrations/00031_seed_sync_runs.sql` produced by sibling work on W02-E05. To
prevent the new CI gate from failing on a migration owned by another epic, this
story added a `+wowapi:manifest` block to that file and updated
`migrations/migrations_test.go` `expectedFiles`.

- Owner of the affected file: W02Seed (production seed-sync epic).
- Coordination: message sent to W02Seed; no objection received before closure.
- Impact: W02-E05's migration now satisfies the manifest schema; the gate passes.
- Resolution: accepted as a cross-epic integration necessity.
