---
id: VER-W06-E04-S001
type: verification-record
parent_story: W06-E04-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W06-E04-S001

| Acceptance criterion | Verification method | Actual result | Evidence | Reviewer |
|---|---|---|---|---|
| AC-W06-E04-S001-01 | `go run ./internal/tools/docexamples -root .` | PASS — 1 normative example compiled; every Go fence explicitly classified | EV-W06-E04-S001-001; REV-W06-E04-S001-001 | W06-E01-E04-Execution.W06E04ReviewR |
| AC-W06-E04-S001-02 | `make docs-check`; inspect CI unit job | PASS — Make and CI invoke the same gate | EV-W06-E04-S001-002; REV-W06-E04-S001-001 | W06-E01-E04-Execution.W06E04ReviewR |
| AC-W06-E04-S001-03 | removed-symbol fixture + current repository gate | PASS — `app.RunAPI` rejected at fixture line 7 with deterministic diagnostics; current docs pass | EV-W06-E04-S001-003-R1; REV-W06-E04-S001-001 | W06-E01-E04-Execution.W06E04ReviewR |

## Execution record

- **Date/time:** 2026-07-13T16:46:25Z.
- **Revision:** `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus shared uncommitted W05/W06 changes.
- **Branch:** `main`.
- **Environment:** macOS Darwin 25.5.0 arm64; Go 1.26.5.
- **Retest:** `go test ./internal/tools/docexamples -v`, direct extractor, and `make docs-check` all passed.
- **Findings:** a TDD regression test found `go build` initially leaked `example-001`/`example-002`
  binaries into the repository root; fixed by directing outputs into each throwaway package. No leaked
  artifacts remain and the regression test passes.
- **Conclusion:** all acceptance checks and mandate §14 independent review passed; no open issues.
