---
id: ART-W02-E01-S001-001
type: artifact
title: Migration manifest schema — design, format options, and locked format
parent_story: W02-E01-S001
parent_task: W02-E01-S001-T001
status: draft-pending-external-review
created_at: 2026-07-13
updated_at: 2026-07-13
---

# Migration manifest schema — design record (DATA-09 T1)

Base commit: 1626b1132622aacc3e85475e4190e16a457ad1f6. Current-state re-confirmation executed at
this commit: `Makefile` `migrate` target is a plain forward-apply (`go run ./internal/tools/migrate`);
`miscellaneous/check_migrations.sh` checks registration/markers/numbering only; zero occurrences of
`lock_timeout` in Go production code; zero manifest concept in `migrations/*.sql`. The story's
"confirmed absence" assessment holds.

## Storage-format options considered

| Option | Pros | Cons |
|---|---|---|
| **A. Inline structured header comment in the migration `.sql` file** | One file per migration — no sibling-file drift possible; travels through the existing `//go:embed *.sql` unchanged (no embed-pattern change, no template-hash change beyond the file itself); reviewable in the same diff as the DDL it classifies; parseable with stdlib only (no YAML dependency in the validator path); goose ignores non-`+goose` comment lines | Schema is line-oriented, not nested; long validation queries must fit a line (mitigated: `validation_query` accepts one SQL statement; multi-check validation belongs to the validation-phase tooling's own check list, not the manifest) |
| B. Sibling YAML/JSON file per migration (`00031_x.manifest.yaml`) | Rich nesting; independent evolution | Requires widening the embed pattern and the template-name hash inputs; introduces file-pair drift (`check_migrations.sh`-class trap this repo has already been bitten by — unregistered/mispaired files); YAML parse in validator |
| C. Central registry file (one YAML for all migrations) | Single place to audit | Merge-conflict magnet with 5 concurrent W02 workers appending migrations; distance between DDL and its declaration invites stale entries |

**Selected: Option A — inline structured header comment block.** Rationale: it makes manifest
omission structurally visible in the migration's own diff, cannot drift from its migration, costs no
new dependency or embed change, and is the same mechanism goose itself uses for its markers, so
migration authors already think in header-annotation terms. Option C was rejected primarily on the
concurrent-authorship conflict surface (this very wave has four sibling workers adding migrations);
Option B primarily on the file-pair drift trap.

## Locked format (pending external review)

A block anywhere in the file (conventionally at the top, before `-- +goose Up`):

```sql
-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: none
-- rollback_plan: goose Down drops the objects this migration creates; no data loss possible (additive-only).
-- +wowapi:end
```

Grammar: every line between the markers is `-- <key>: <value>`; keys are snake_case; unknown keys
are rejected (catches typos — a misspelled required key would otherwise read as "missing" with a
confusing message); duplicate keys are rejected.

## Required fields (all nine; PLAN DATA-09 T1's verbatim field list)

| Field | Type / domain | Validation rule |
|---|---|---|
| `classification` | `online` \| `maintenance` | enum, required. `online` = safe to apply under live traffic; `maintenance` = requires a maintenance window. |
| `rows_estimate` | integer ≥ 0 | required. Estimated rows affected/backfilled (0 for pure catalog DDL). |
| `bytes_estimate` | integer ≥ 0 | required. Estimated bytes written/rewritten. |
| `lock_timeout_ms` | integer > 0 | required. For `classification: online` MUST be ≤ 2000 (the DATA-09 T2 budget — the manifest cannot declare its way out of the online lock budget). |
| `statement_timeout_ms` | integer > 0 | required; MUST be ≥ `lock_timeout_ms`. |
| `nn1_compatible` | `true` \| `false` | required. Whether N-1 application binaries keep working while/after this migration is applied (the N/N-1 flag). |
| `backfill_owner` | non-empty string | required. `none` when the migration needs no backfill; otherwise the owning person/team/story ID. |
| `validation_query` | non-empty string | required. A single SQL statement returning a mismatch count (0 = pass), or the literal `none` with the implicit claim the migration is self-evidently correct (catalog-only DDL). |
| `rollback_plan` | non-empty string | required. Rollback and/or forward-fix plan in one sentence or more. |

## Enforcement boundary (compatibility decision, recorded not defaulted)

The CI gate requires a valid manifest for every kernel migration numbered **≥ 00031** (the first
migration authored after this schema exists). Migrations 00001–00030 predate the schema;
retroactively classifying them is per-migration human judgment (PLAN T1's own classification
column) over already-applied history with no remaining execution risk, and is deliberately NOT
absorbed into this story (story.md "Out of scope" authorizes exactly this split). A manifest block
present on a pre-00031 migration is still parsed and validated (opt-in). The boundary constant
lives next to the validator with this rationale.

## Enforcement mechanism

`migrations/manifest_test.go` — a plain Go test (no DB) that walks `migrations.Kernel()` and
validates every file ≥ 00031 via `kernel/migration.ParseManifest`. It runs in every `go test
./migrations` invocation, i.e. in the existing CI test job — no new workflow step, no bash
duplication in `check_migrations.sh`. Field-specific failure messages name the file, the field, and
the rule violated. Positive/negative fixture pairs live in `kernel/migration/testdata/` and are
exercised by the validator's own unit tests.

## Relationship to the lock-budget mechanism (T2)

`lock_timeout_ms` is the declared budget; `kernel/migration.ExecDDL` is the enforcement: it sets
the session/transaction `lock_timeout` from the caller-supplied budget (default 2000ms), aborts
cleanly on SQLSTATE 55P03, and retries within a bounded ceiling. The manifest declares; the
executor enforces; the CI gate proves the declaration exists and is well-formed.
