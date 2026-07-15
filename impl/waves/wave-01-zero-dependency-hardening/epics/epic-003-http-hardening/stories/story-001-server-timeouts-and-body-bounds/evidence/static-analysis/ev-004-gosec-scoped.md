# EV-W01-E03-S001-004 — scoped gosec run (G120 resolution)

- **Evidence ID**: EV-W01-E03-S001-004
- **Evidence type**: static-analysis report
- **Story / task**: W01-E03-S001 / W01-E03-S001-T003
- **Acceptance criteria proven**: AC-W01-E03-S001-04 (static-analysis half)
- **Execution command**: `gosec -quiet ./kernel/httpx/`
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: zero findings in kernel/httpx post-fix (exit 0). Honesty note: the installed gosec is a dev build whose rule set may not carry the exact G120 id the story text cites; this run proves *no* gosec finding exists at `csrf.go`'s FormValue call site, and the functional proof (EV-W01-E03-S001-005) is the primary behavioral evidence that the read is bounded. The story's own AC anticipates a definitive re-run once W01-E01-S002 enables the pinned linter set.

```
gosec exit: 0
```
