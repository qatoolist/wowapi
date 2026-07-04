# wowapi — Reusable Internal Scripts Policy

All internal check / audit / validation / test-inspection / duplicate-detection / reporting / analysis
scripts live in **`miscellaneous/`** at the repo root. Their catalog + usage table is
[`miscellaneous/README.md`](../../miscellaneous/README.md).

## Why
One-off scripts scattered across the repo get lost and rewritten. A single home makes them discoverable
and reusable, and lets the review gate (`miscellaneous/review_gate.sh`) run them as a batch. This
directory was created for exactly the recurring risks reviews surfaced (unregistered migrations, hollow
coverage, unwired primitives, doc overclaims).

## Rules for every script
1. **Location** — in `miscellaneous/`. Do not create disposable scripts elsewhere (use the session
   scratchpad for truly throwaway work, never the repo).
2. **Clear name** — `check_<thing>.sh`, `find_<thing>.sh`, `<verb>_<thing>.sh`. Reusable, not
   goal-specific.
3. **Documented header** — a comment block stating: what it checks · when to run · usage · exit-code
   meaning.
4. **Read-only & safe** — never modify project files. Read/grep/report only. (`review_gate.sh --full` may
   build via `make`, which is non-destructive.)
5. **Self-locating** — `cd "$(dirname "$0")/.."` so it works from any CWD.
6. **No hardcoded temporary assumptions** — parameterize (e.g. `check_unwired.sh <pkg>`); use repo-relative
   paths; write scratch output only under `/tmp` if needed.
7. **Preserve useful output format** — stable, greppable lines; a clear `OK` / `FAIL:` prefix; a summary
   line. Advisory scripts print and exit 0; gating scripts exit 1 on issues.
8. **Register it** — add a row to `miscellaneous/README.md` and, if it maps to a recurring risk, reference
   it from [quality-gate-checklist.md](quality-gate-checklist.md).
9. **Keep it dependency-light** — POSIX-ish `bash` + `grep`/`awk`/`git`/`go`. No new toolchain.

## Lifecycle
- When a review finds a mechanically-detectable class of defect, add or extend a script here and wire it
  into `review_gate.sh` so it can't recur silently (this is how the migration-ledger and stray-tag checks
  were born).
- Scripts are code: keep them working. `review_gate.sh` exercises the others, so a broken check surfaces.
