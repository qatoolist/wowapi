# Phase 0 — Review Findings

Two parallel critique agents reviewed the walking skeleton (2026-07-03):
**A** = architecture/package-boundary/API-surface reviewer; **S** = security/config/redaction reviewer.
Plus one defect self-caught by the test suite before review (**T**).

| ID | Sev | Source | Finding (file:line) | Resolution | Status |
|---|---|---|---|---|---|
| ARCH-1 | medium | A | `scripts/lint_boundaries.sh` prefix match `index($1,p)==1` also matches sibling packages (`kernelx` vs `kernel`), silently missing/misattributing violations once such packages exist | rewrote matching to path-segment-aware: `$1 == p":" \|\| index($1, p"/") == 1` for the package and `$i == f \|\| index($i, f"/") == 1` for the import | **fixed** |
| ARCH-2 | medium | A | testkit rule was note-only (not `fail=1`) because prod+test imports were concatenated — the hard blueprint rule "production never imports testkit" was unenforced | script now builds separate prod/test import files; production import of testkit is a hard failure; test imports of testkit allowed | **fixed** |
| ARCH-3 | low | A | rules 1–4 mixed TestImports into production checks → future false positive when `app`'s own tests legitimately import testkit | same split as ARCH-2; layering rules run against prod imports, with a narrower law applied to test imports (kernel tests can't import app, etc.) | **fixed** |
| ARCH-4 | low | A | `app.Ordered()` ran Validate + topo-sort twice | refactored to shared `validateAndOrder()`; `Validate`/`Ordered` are thin wrappers | **fixed** |
| ARCH-5 | info | A | `MapView.Decode` marshal-side errors (non-JSON-marshalable values) may surprise callers expecting only target-struct errors | accepted: Decode returns an error either way; doc comment already states the JSON round-trip. No change. | **rejected (agreed)** |
| SEC-1 | medium* | S | `Defaults()` env = `local` → a prod deploy that fails to set `environment` silently validates under local rules (silent security downgrade). *Reported informational by reviewer (loader not yet built); treated as medium spec gap.* | blueprint 12 §4 now mandates fail-closed explicit `environment` in deployed processes (loader errors when absent from every layer); recorded as D-0010; Phase 1 loader test must cover it | **fixed (spec) / carried to Phase 1 (code)** |
| SEC-2 | low | S | `secrets.redactCandidate` echoed first 2 chars — 40% disclosure of a 5-char secret | now always returns `****`; test asserts no echo | **fixed** |
| T-1 | low | tests | module name regex `{1,63}` rejected single-char names; cycle-detection test failed with misleading "unknown module" errors | regex → `{0,63}` (min length 1); full suite green | **fixed** |

Residual risk (both reviewers): Reveal()/vocabulary greps are heuristic until the Phase 5 AST lint
(accepted, D-0009); prod-safety depends on the Phase 1 loader implementing the fail-closed
environment rule (tracked as a Phase 1 exit criterion).

Re-verification after fixes: `go vet` / `go test ./...` / `-race` / boundary lint all exit 0
(command-log #15).
