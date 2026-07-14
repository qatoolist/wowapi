---
id: ART-W07-E02-S002-001
type: implementation-artifact
parent_story: W07-E02-S002
status: produced
version: 1
created_at: 2026-07-14
updated_at: 2026-07-14
---

# Test-skip classification

The source review cited a historical inventory of 22 `t.Skip` sites. The execution-time AST scan found
39 sites after intervening package/test expansion. The probabilistic TOTP skip was removed by making the
wrong-code test deterministic; all remaining 38 sites are registered in
`miscellaneous/test-skip-manifest.json` with a stable path/function/method/message identity, owner,
classification, and rationale.

| Classification | Count | Manifest IDs | Enforcement |
|---|---:|---|---|
| Required, fail closed | 15 | SKIP-001..003, 005, 009, 017..021, 034..038 | The named `WOWAPI_REQUIRE_DB` or `WOWAPI_REQUIRE_S3` guard reaches `t.Fatal`/`t.Fatalf` before the local-only skip. |
| Legitimately optional | 23 | Every other registered ID | Limited to explicit `-short` profiles, an inapplicable privilege/permission environment, or another documented environment-specific branch. |

The manifest is the exhaustive per-site record. `internal/tools/testskipmanifest` parses Go AST rather
than line-oriented text, rejects any unapproved site, rejects stale approvals, and rejects entries
without an assigned owner or rationale. `miscellaneous/check_test_skip_fixtures.sh` proves an unapproved
fixture fails while a complete approved fixture passes.

Required prerequisite conversions cover DB-backed CLI, E2E, tenant-FK, RLS-guard, outbox, and testkit
paths plus all MinIO paths. The authoritative gate exports both requirement flags. The E2E Go-toolchain
and cold-module-cache branches now also fail with actionable remediation when either authoritative
integration flag is set.
