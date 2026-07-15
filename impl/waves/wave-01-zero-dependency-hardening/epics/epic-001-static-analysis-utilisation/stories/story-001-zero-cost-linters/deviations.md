---
id: DEV-W01-E01-S001
type: deviations-record
parent_story: W01-E01-S001
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W01-E01-S001

## DEV-W01-E01-S001-001 — noctx does not report the named exec.Command sites

- **Approved plan**: AC-02 expects `noctx` fail-before/pass-after runs against
  `internal/cli/config_delegate.go` and `internal/cli/lint_cmd.go` (`exec.Command` → `exec.CommandContext`).
- **Actual**: noctx as shipped in pinned golangci-lint v2.11.4 checks net/http request construction
  (and database/sql), NOT `os/exec` — the two named sites produce no noctx finding before OR after
  the fix. This drift was already flagged by W00-E02-S001's baseline capture ("noctx named exec
  sites unreported by v2.11.4") and reproduced by this story's own fresh runs (146 noctx hits, none
  at the named sites).
- **Disposition**: the sites were fixed to `exec.CommandContext` anyway (story scope is the fix, not
  the analyzer's opinion). Fail-before evidence comes from **gosec G204**, which did flag both sites
  in both triage runs; pass-after = per-linter gosec run at 0 hits + the code diff
  (`evidence/static-analysis/noctx-copyloopvar-site-fix.diff`).
- **Impact**: none on the delivered contract; AC-02's evidence mechanism substituted, not weakened.

## DEV-W01-E01-S001-002 — noctx surfaced 146 hits, not the cited 2

- **Actual**: 145 hits in `_test.go` files (context-less `httptest.NewRequest`) + 1 in
  `testkit/i18n.go` (non-test, shipped test-support helper). Cited "2 prod hits" reflected the
  MATRIX's exec-site expectation, which v2.11.4 does not implement (see DEV-001).
- **Disposition**: `_test.go` excluded for noctx in `.golangci.yml` (documented inline: request
  cancellation is meaningless for httptest-constructed requests in tests); `testkit/i18n.go:33`
  fixed with `NewRequestWithContext`. Recorded per RISK-W01-E01-002's contingency rather than
  silently absorbing.

## DEV-W01-E01-S001-003 — new musttag hit in sibling-new file (zero-cost set not zero at Phase 2)

- **Actual**: `internal/cli/init_version.go:112` (W01Gen's new file, landed mid-wave) tripped
  `musttag` at the Phase-2 enablement run — the zero-cost set was zero-hit at HEAD but not over the
  wave's combined working tree.
- **Disposition**: fixed in-story (json tag on the decode struct), per the conductor's green-light
  instruction to lint the tree as-is including new files. Recorded, not silently absorbed.

## Note (below deviation threshold)

copyloopvar's 6 test-file hits (beyond the 1 named production site) were fixed rather than excluded —
mechanical deletions with zero semantic effect on Go ≥1.22; scope-conformant under the same
RISK-W01-E01-002 contingency.
